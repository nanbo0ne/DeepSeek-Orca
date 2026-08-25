package permission

import (
	"context"
	"encoding/json"
	"regexp"
	"sort"
	"strings"
)

type RiskLevel string

const (
	RiskLow    RiskLevel = "low"
	RiskMedium RiskLevel = "medium"
	RiskHigh   RiskLevel = "high"
)

// RiskInput is the deliberately small, redacted payload sent to an independent
// risk classifier. It never carries conversation history, tool schemas, images,
// or provider credentials.
type RiskInput struct {
	Tool            string          `json:"tool"`
	Operation       string          `json:"operation,omitempty"`
	Arguments       json.RawMessage `json:"arguments,omitempty"`
	ReadOnly        bool            `json:"read_only"`
	SecurityContext string          `json:"security_context"`
}

type RiskAssessment struct {
	Level  RiskLevel `json:"level"`
	Reason string    `json:"reason"`
}

type RiskClassifier interface {
	Assess(context.Context, RiskInput) (RiskAssessment, error)
}

// AutoReviewer is the Gate-facing adapter. false means the call should continue
// to the ordinary interactive Approver; an error means preserve the historical
// auto-approval fallback and allow the call.
type AutoReviewer interface {
	Review(context.Context, string, string, json.RawMessage, bool) (bool, error)
}

const (
	maxRiskStringBytes = 512
	maxRiskArgsBytes   = 8 * 1024
)

var (
	riskSecretAssignment = regexp.MustCompile(`(?i)\b(api[_-]?key|access[_-]?token|auth(?:orization)?|bearer|password|passwd|secret|credential|cookie)\s*[:=]\s*([^\s,;]+)`)
	riskBearerToken      = regexp.MustCompile(`(?i)\bbearer\s+[a-z0-9._~+/-]{8,}`)
	riskAPIKeyToken      = regexp.MustCompile(`(?i)\bsk-[a-z0-9_-]{8,}`)
)

// RedactedRiskInput converts raw tool arguments into the only payload the risk
// model may see. Sensitive fields are replaced by markers and large/binary image
// content is omitted before the request is built.
func RedactedRiskInput(toolName, subject string, args json.RawMessage, readOnly bool) RiskInput {
	out := RiskInput{
		Tool:            truncateRiskString(strings.TrimSpace(toolName)),
		Operation:       redactRiskString(subject),
		ReadOnly:        readOnly,
		SecurityContext: "Host deny and explicit ask rules were evaluated before this automatic review.",
	}
	if len(args) == 0 {
		return out
	}
	var raw any
	if err := json.Unmarshal(args, &raw); err != nil {
		raw = map[string]any{"raw": redactRiskString(string(args))}
	}
	redacted := redactRiskValue(raw, "")
	b, err := json.Marshal(redacted)
	if err != nil {
		return out
	}
	if len(b) > maxRiskArgsBytes {
		b, _ = json.Marshal(map[string]any{
			"summary": "arguments omitted after redaction because the payload was too large",
			"keys":    riskTopLevelKeys(redacted),
		})
	}
	out.Arguments = b
	return out
}

func redactRiskValue(value any, key string) any {
	if riskImageKey(key) {
		return "[OMITTED_IMAGE]"
	}
	if riskSensitiveKey(key) {
		return "[REDACTED]"
	}
	switch typed := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(typed))
		for childKey, child := range typed {
			out[childKey] = redactRiskValue(child, childKey)
		}
		return out
	case []any:
		limit := len(typed)
		if limit > 32 {
			limit = 32
		}
		out := make([]any, 0, limit+1)
		for i := 0; i < limit; i++ {
			out = append(out, redactRiskValue(typed[i], key))
		}
		if len(typed) > limit {
			out = append(out, "[TRUNCATED]")
		}
		return out
	case string:
		if strings.HasPrefix(strings.ToLower(strings.TrimSpace(typed)), "data:image/") {
			return "[OMITTED_IMAGE]"
		}
		return redactRiskString(typed)
	default:
		return value
	}
}

func riskSensitiveKey(key string) bool {
	key = normalizeRiskKey(key)
	for _, marker := range []string{"apikey", "accesstoken", "authtoken", "authorization", "bearer", "password", "passwd", "secret", "credential", "cookie"} {
		if strings.Contains(key, marker) {
			return true
		}
	}
	return false
}

func riskImageKey(key string) bool {
	key = normalizeRiskKey(key)
	for _, marker := range []string{"image", "screenshot", "base64", "datauri"} {
		if strings.Contains(key, marker) {
			return true
		}
	}
	return false
}

func normalizeRiskKey(key string) string {
	key = strings.ToLower(strings.TrimSpace(key))
	return strings.NewReplacer("_", "", "-", "", ".", "", " ", "").Replace(key)
}

func redactRiskString(value string) string {
	value = strings.TrimSpace(value)
	value = riskSecretAssignment.ReplaceAllString(value, "$1=[REDACTED]")
	value = riskBearerToken.ReplaceAllString(value, "Bearer [REDACTED]")
	value = riskAPIKeyToken.ReplaceAllString(value, "[REDACTED_KEY]")
	if strings.Contains(strings.ToLower(value), "data:image/") {
		return "[OMITTED_IMAGE]"
	}
	return truncateRiskString(value)
}

func truncateRiskString(value string) string {
	if len(value) <= maxRiskStringBytes {
		return value
	}
	return value[:maxRiskStringBytes] + "...[TRUNCATED]"
}

func riskTopLevelKeys(value any) []string {
	m, ok := value.(map[string]any)
	if !ok {
		return nil
	}
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	if len(keys) > 32 {
		keys = keys[:32]
	}
	return keys
}
