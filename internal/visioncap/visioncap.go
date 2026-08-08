package visioncap

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"deepseek-orca/internal/config"
	"deepseek-orca/internal/provider"
)

const (
	Supported           = "supported"
	Unsupported         = "unsupported"
	Unknown             = "unknown"
	Probing             = "probing"
	OverrideAuto        = "auto"
	SourceProbe         = "probe"
	SourceMetadata      = "metadata"
	SourceManual        = "manual"
	CurrentProbeVersion = 2
)

type Capability struct {
	ModelRef     string `json:"modelRef"`
	Key          string `json:"key"`
	Status       string `json:"status"`
	CheckedAt    int64  `json:"checkedAt,omitempty"`
	Reason       string `json:"reason,omitempty"`
	Attempts     int    `json:"attempts,omitempty"`
	Source       string `json:"source,omitempty"`
	Override     string `json:"override,omitempty"`
	ProbeVersion int    `json:"probeVersion,omitempty"`
}

type Store struct {
	mu    sync.Mutex
	path  string
	Items map[string]Capability `json:"items"`
}

// Store instances are intentionally cheap and are loaded independently by the
// desktop settings and boot paths. Serialize file updates across those
// instances so parallel model probes cannot overwrite each other's result.
var storeFileMu sync.Mutex

func DefaultPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		home, _ := os.UserHomeDir()
		dir = filepath.Join(home, ".deepseek-orca")
	} else {
		dir = filepath.Join(dir, "deepseek-orca")
	}
	return filepath.Join(dir, "vision-capabilities.json")
}

func Load(path string) *Store {
	if strings.TrimSpace(path) == "" {
		path = DefaultPath()
	}
	s := &Store{path: path, Items: map[string]Capability{}}
	if b, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(b, s)
		s.path = path
	}
	if s.Items == nil {
		s.Items = map[string]Capability{}
	}
	for key, item := range s.Items {
		if item.Status == Probing {
			item.Status = Unknown
			item.Reason = "previous vision probe was interrupted"
			s.Items[key] = item
		}
		if item.Source == SourceProbe && item.ProbeVersion < CurrentProbeVersion {
			item.Status = Unknown
			item.Reason = "vision probe needs refresh"
			item.Attempts = 0
			s.Items[key] = item
		}
	}
	return s
}

func Key(e *config.ProviderEntry) string {
	if e == nil {
		return ""
	}
	base := strings.TrimRight(strings.ToLower(strings.TrimSpace(e.BaseURL)), "/")
	return strings.ToLower(strings.TrimSpace(e.Kind)) + "|" + base + "|" + strings.TrimSpace(e.Model)
}

func ModelRef(e *config.ProviderEntry) string {
	if e == nil {
		return ""
	}
	return strings.TrimSpace(e.Name) + "/" + strings.TrimSpace(e.Model)
}

func (s *Store) Get(e *config.ProviderEntry) Capability {
	c := s.Stored(e)
	if c.Override == Supported || c.Override == Unsupported {
		c.Status = c.Override
		c.Source = SourceManual
	}
	return c
}

// Stored returns the last automatic result without applying a manual override.
// Callers that update or clear the override need this underlying value.
func (s *Store) Stored(e *config.ProviderEntry) Capability {
	k := Key(e)
	s.mu.Lock()
	defer s.mu.Unlock()
	if c, ok := s.Items[k]; ok {
		c.ModelRef = ModelRef(e)
		return c
	}
	return Capability{ModelRef: ModelRef(e), Key: k, Status: Unknown, Override: OverrideAuto}
}

func (s *Store) Put(c Capability) error {
	storeFileMu.Lock()
	defer storeFileMu.Unlock()
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.Items == nil {
		s.Items = map[string]Capability{}
	}
	if b, err := os.ReadFile(s.path); err == nil {
		var disk struct {
			Items map[string]Capability `json:"items"`
		}
		if json.Unmarshal(b, &disk) == nil {
			for key, item := range disk.Items {
				s.Items[key] = item
			}
		}
	}
	s.Items[c.Key] = c
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(struct {
		Items map[string]Capability `json:"items"`
	}{s.Items}, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

func (s *Store) List(cfg *config.Config) []Capability {
	if cfg == nil {
		return nil
	}
	var out []Capability
	for i := range cfg.Providers {
		for _, model := range cfg.Providers[i].ModelList() {
			e := cfg.Providers[i]
			e.Model = model
			out = append(out, s.Get(&e))
		}
	}
	return out
}

// Probe sends a generated high-contrast digit image. It is intentionally
// independent of desktop/Wails so fake providers can test every outcome.
func Probe(ctx context.Context, p provider.Provider, e *config.ProviderEntry, iconPNG []byte) Capability {
	code, data, err := probeImage(iconPNG)
	c := Capability{ModelRef: ModelRef(e), Key: Key(e), Status: Unknown, CheckedAt: time.Now().UnixMilli(), Source: SourceProbe, ProbeVersion: CurrentProbeVersion}
	if err != nil {
		c.Reason = err.Error()
		return c
	}
	return probeWithImage(ctx, p, e, code, data)
}

func probeWithImage(ctx context.Context, p provider.Provider, e *config.ProviderEntry, code, data string) Capability {
	c := Capability{ModelRef: ModelRef(e), Key: Key(e), Status: Unknown, CheckedAt: time.Now().UnixMilli(), Source: SourceProbe, Override: OverrideAuto, ProbeVersion: CurrentProbeVersion}
	req := provider.Request{Temperature: 0, MaxTokens: 1024, Messages: []provider.Message{{Role: provider.RoleUser, Content: "Inspect the attached DeepSeek-Orca icon test image. Read the four-digit code printed below the icon and reply with the digits only.", Images: []provider.ImageContent{{Name: "orca-vision-probe.png", MediaType: "image/png", Data: data}}}}}
	ch, err := p.Stream(ctx, req)
	if err != nil {
		c.Status = statusForProbeError(err)
		c.Reason = summarizeError(err)
		return c
	}
	var b strings.Builder
	for chunk := range ch {
		if chunk.Type == provider.ChunkText {
			b.WriteString(chunk.Text)
		}
		if chunk.Type == provider.ChunkError && chunk.Err != nil {
			c.Status = statusForProbeError(chunk.Err)
			c.Reason = summarizeError(chunk.Err)
			return c
		}
	}
	answer := strings.TrimSpace(b.String())
	if probeAnswerMatches(answer, code) {
		c.Status = Supported
		return c
	}
	if explicitlyRejectsImageAnswer(answer) {
		c.Status = Unsupported
		c.Reason = "model explicitly stated that it cannot inspect images"
		return c
	}
	c.Status = Unknown
	if answer == "" {
		c.Reason = "probe completed without a verifiable final answer"
	} else {
		c.Reason = "probe response did not contain the test code"
	}
	return c
}

func explicitlyRejectsImageAnswer(answer string) bool {
	message := strings.ToLower(strings.TrimSpace(answer))
	if message == "" {
		return false
	}
	mentionsImage := strings.Contains(message, "image") || strings.Contains(message, "vision") || strings.Contains(message, "图片") || strings.Contains(message, "图像")
	rejects := strings.Contains(message, "cannot") || strings.Contains(message, "can't") || strings.Contains(message, "unable") || strings.Contains(message, "not support") || strings.Contains(message, "无法") || strings.Contains(message, "不支持")
	return mentionsImage && rejects
}

func probeAnswerMatches(answer, code string) bool {
	answer = strings.TrimSpace(answer)
	if answer == code {
		return true
	}
	// Providers occasionally wrap a correct short answer in punctuation or a
	// sentence despite the digits-only instruction. Require a standalone match
	// so unrelated numbers in a refusal cannot produce a false positive.
	for i := 0; i+len(code) <= len(answer); i++ {
		if answer[i:i+len(code)] != code {
			continue
		}
		leftDigit := i > 0 && answer[i-1] >= '0' && answer[i-1] <= '9'
		right := i + len(code)
		rightDigit := right < len(answer) && answer[right] >= '0' && answer[right] <= '9'
		if !leftDigit && !rightDigit {
			return true
		}
	}
	return false
}

func statusForProbeError(err error) string {
	if err == nil {
		return Unknown
	}
	message := strings.ToLower(err.Error())
	transientOrAccess := []string{"401", "403", "408", "409", "429", "5xx", "timeout", "timed out", "rate limit", "quota", "network", "connection", "authentication", "unauthorized", "forbidden"}
	for _, marker := range transientOrAccess {
		if strings.Contains(message, marker) {
			return Unknown
		}
	}
	mentionsImage := strings.Contains(message, "image") || strings.Contains(message, "vision") || strings.Contains(message, "multimodal")
	explicitRejection := strings.Contains(message, "not support") || strings.Contains(message, "unsupported") || strings.Contains(message, "doesn't support") || strings.Contains(message, "does not accept") || strings.Contains(message, "invalid content type") || strings.Contains(message, "text only") || strings.Contains(message, "no endpoints found")
	if mentionsImage && explicitRejection {
		return Unsupported
	}
	return Unknown
}

func summarizeError(err error) string {
	if err == nil {
		return "probe failed"
	}
	s := strings.TrimSpace(err.Error())
	if len(s) > 240 {
		s = s[:240]
	}
	return s
}

var digitRows = [10][7]string{
	{"111", "101", "101", "101", "101", "101", "111"}, {"010", "110", "010", "010", "010", "010", "111"},
	{"111", "001", "001", "111", "100", "100", "111"}, {"111", "001", "001", "111", "001", "001", "111"},
	{"101", "101", "101", "111", "001", "001", "001"}, {"111", "100", "100", "111", "001", "001", "111"},
	{"111", "100", "100", "111", "101", "101", "111"}, {"111", "001", "001", "010", "010", "100", "100"},
	{"111", "101", "101", "111", "101", "101", "111"}, {"111", "101", "101", "111", "001", "001", "111"},
}

func probeImage(iconPNG []byte) (string, string, error) {
	raw := make([]byte, 4)
	if _, err := rand.Read(raw); err != nil {
		return "", "", err
	}
	code := fmt.Sprintf("%d%d%d%d", raw[0]%10, raw[1]%10, raw[2]%10, raw[3]%10)
	img := image.NewRGBA(image.Rect(0, 0, 320, 230))
	white := color.RGBA{255, 255, 255, 255}
	black := color.RGBA{15, 23, 42, 255}
	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dx(); x++ {
			img.Set(x, y, white)
		}
	}
	if len(iconPNG) > 0 {
		icon, _, decodeErr := image.Decode(bytes.NewReader(iconPNG))
		if decodeErr != nil {
			return "", "", fmt.Errorf("decode probe icon: %w", decodeErr)
		}
		src := icon.Bounds()
		const iconSize = 128
		const iconX = (320 - iconSize) / 2
		const iconY = 10
		for y := 0; y < iconSize; y++ {
			for x := 0; x < iconSize; x++ {
				sx := src.Min.X + x*src.Dx()/iconSize
				sy := src.Min.Y + y*src.Dy()/iconSize
				img.Set(iconX+x, iconY+y, icon.At(sx, sy))
			}
		}
	}
	for i, ch := range code {
		d := int(ch - '0')
		ox := 46 + i*58
		for row, bits := range digitRows[d] {
			for col, bit := range bits {
				if bit != '1' {
					continue
				}
				for yy := 0; yy < 12; yy++ {
					for xx := 0; xx < 12; xx++ {
						img.Set(ox+col*12+xx, 145+row*12+yy, black)
					}
				}
			}
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", "", err
	}
	return code, base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}
