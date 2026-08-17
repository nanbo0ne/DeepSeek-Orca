package weixin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/nanbo0ne/O.R.C.A-for-Windows/internal/config"
)

type savedAccount struct {
	Token   string `json:"token"`
	BaseURL string `json:"base_url"`
	UserID  string `json:"user_id"`
	SavedAt string `json:"saved_at"`
}

type LoginResult struct {
	AccountID        string
	Token            string
	BaseURL          string
	UserID           string
	AlreadyConnected bool
}

type LoginSession struct {
	SessionKey        string
	QRCode            string
	QRCodeURL         string
	BaseURL           string
	StartedAt         time.Time
	PendingVerifyCode string
}

func weixinAccountDir(root string) string {
	return filepath.Join(root, "weixin", "accounts")
}

func savedAccountPath(accountID string) string {
	root := config.MemoryUserDir()
	if root == "" || accountID == "" {
		return ""
	}
	return filepath.Join(weixinAccountDir(root), accountID+".json")
}

func loadSavedAccount(accountID string) (savedAccount, error) {
	path := savedAccountPath(accountID)
	if path == "" {
		return savedAccount{}, fmt.Errorf("O.R.C.A user config dir is unavailable")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return savedAccount{}, err
	}
	var account savedAccount
	if err := json.Unmarshal(data, &account); err != nil {
		return savedAccount{}, err
	}
	return account, nil
}

func loadAnySavedAccount() (savedAccount, error) {
	_, account, err := loadAnySavedAccountWithID()
	return account, err
}

func loadAnySavedAccountWithID() (string, savedAccount, error) {
	root := config.MemoryUserDir()
	if root == "" {
		return "", savedAccount{}, fmt.Errorf("O.R.C.A user config dir is unavailable")
	}
	entries, err := os.ReadDir(weixinAccountDir(root))
	if err != nil {
		return "", savedAccount{}, err
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") || strings.Contains(entry.Name(), "context-tokens") {
			continue
		}
		accountID := strings.TrimSuffix(entry.Name(), ".json")
		account, err := loadSavedAccount(accountID)
		if err == nil && strings.TrimSpace(account.Token) != "" {
			return accountID, account, nil
		}
	}
	return "", savedAccount{}, fmt.Errorf("no saved weixin account")
}

func HasSavedAccount(accountID string) bool {
	if accountID != "" {
		account, err := loadSavedAccount(accountID)
		return err == nil && strings.TrimSpace(account.Token) != ""
	}
	account, err := loadSavedAccount("default")
	if err == nil && strings.TrimSpace(account.Token) != "" {
		return true
	}
	_, account, err = loadAnySavedAccountWithID()
	return err == nil && strings.TrimSpace(account.Token) != ""
}

func saveAccount(accountID string, account savedAccount) error {
	path := savedAccountPath(accountID)
	if path == "" {
		return fmt.Errorf("O.R.C.A user config dir is unavailable")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(account, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func localBotTokenList(limit int) []string {
	root := config.MemoryUserDir()
	if root == "" || limit <= 0 {
		return []string{}
	}
	entries, err := os.ReadDir(weixinAccountDir(root))
	if err != nil {
		return []string{}
	}
	type item struct {
		name string
		mod  time.Time
	}
	items := make([]item, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") || strings.Contains(entry.Name(), "context-tokens") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		items = append(items, item{name: strings.TrimSuffix(entry.Name(), ".json"), mod: info.ModTime()})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].mod.After(items[j].mod) })
	out := make([]string, 0, limit)
	for _, item := range items {
		account, err := loadSavedAccount(item.name)
		if err == nil && strings.TrimSpace(account.Token) != "" {
			out = append(out, strings.TrimSpace(account.Token))
			if len(out) >= limit {
				break
			}
		}
	}
	return out
}

func Login(ctx context.Context, out io.Writer, timeout time.Duration) (*LoginResult, error) {
	if timeout <= 0 {
		timeout = 8 * time.Minute
	}
	session, err := StartLogin(ctx)
	if err != nil {
		return nil, err
	}
	if out != nil {
		fmt.Fprintln(out, "请使用微信扫描下面的二维码链接：")
		if session.QRCodeURL != "" {
			fmt.Fprintln(out, session.QRCodeURL)
		} else {
			fmt.Fprintln(out, session.QRCode)
		}
	}

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(time.Second):
		}
		result, status, err := PollLogin(ctx, session)
		if err != nil {
			if out != nil {
				fmt.Fprintf(out, "二维码状态查询失败：%v\n", err)
			}
			continue
		}
		if result != nil {
			return result, nil
		}
		if out != nil {
			switch status {
			case "wait", "", "<nil>":
				fmt.Fprint(out, ".")
			case "scaned":
				fmt.Fprintln(out, "\n已扫码，请在微信里确认。")
			case "need_verifycode":
				fmt.Fprintln(out, "\n微信要求输入配对码，请在桌面连接界面继续。")
			default:
				fmt.Fprintf(out, "\n二维码状态：%s\n", status)
			}
		}
	}
	return nil, fmt.Errorf("weixin login timed out")
}

func StartLogin(ctx context.Context) (*LoginSession, error) {
	qrResp, err := ilinkPostMap(ctx, defaultWeixinAPI, getBotQRPath+"?bot_type=3", map[string]any{
		"local_token_list": localBotTokenList(10),
	})
	if err != nil {
		return nil, fmt.Errorf("fetch qr code: %w", err)
	}
	qrcode := fmt.Sprint(qrResp["qrcode"])
	qrcodeURL := fmt.Sprint(qrResp["qrcode_img_content"])
	if qrcode == "" || qrcode == "<nil>" {
		return nil, fmt.Errorf("weixin qr response missing qrcode")
	}
	if qrcodeURL == "<nil>" {
		qrcodeURL = ""
	}
	return &LoginSession{
		SessionKey: qrcode,
		QRCode:     qrcode,
		QRCodeURL:  qrcodeURL,
		BaseURL:    defaultWeixinAPI,
		StartedAt:  time.Now(),
	}, nil
}

func PollLogin(ctx context.Context, session *LoginSession) (*LoginResult, string, error) {
	if session == nil || session.QRCode == "" {
		return nil, "", fmt.Errorf("weixin login session is missing")
	}
	baseURL := session.BaseURL
	if baseURL == "" {
		baseURL = defaultWeixinAPI
	}
	endpoint := getQRStatusPath + "?qrcode=" + url.QueryEscape(session.QRCode)
	if session.PendingVerifyCode != "" {
		endpoint += "&verify_code=" + url.QueryEscape(session.PendingVerifyCode)
	}
	statusResp, err := ilinkGET(ctx, baseURL, endpoint)
	if err != nil {
		if ctx.Err() != nil || isTemporaryLoginPollError(err) {
			return nil, "wait", nil
		}
		return nil, "", err
	}
	status := fmt.Sprint(statusResp["status"])
	switch status {
	case "wait", "", "<nil>", "scaned":
		return nil, status, nil
	case "need_verifycode", "verify_code_blocked":
		return nil, status, nil
	case "binded_redirect":
		accountID, account, err := loadAnySavedAccountWithID()
		if err == nil {
			return &LoginResult{AccountID: accountID, Token: account.Token, BaseURL: firstNonEmptyString(account.BaseURL, baseURL), UserID: account.UserID, AlreadyConnected: true}, status, nil
		}
		return &LoginResult{AccountID: "default", BaseURL: baseURL, AlreadyConnected: true}, status, nil
	case "scaned_but_redirect":
		if host := fmt.Sprint(statusResp["redirect_host"]); host != "" && host != "<nil>" {
			session.BaseURL = "https://" + host
		}
		return nil, status, nil
	case "confirmed":
		accountID := fmt.Sprint(statusResp["ilink_bot_id"])
		token := fmt.Sprint(statusResp["bot_token"])
		userID := fmt.Sprint(statusResp["ilink_user_id"])
		respBaseURL := fmt.Sprint(statusResp["baseurl"])
		if respBaseURL == "" || respBaseURL == "<nil>" {
			respBaseURL = baseURL
		}
		if accountID == "" || accountID == "<nil>" || token == "" || token == "<nil>" {
			return nil, status, fmt.Errorf("weixin qr confirmed but credential payload is incomplete")
		}
		account := savedAccount{
			Token:   token,
			BaseURL: respBaseURL,
			UserID:  userID,
			SavedAt: time.Now().UTC().Format(time.RFC3339),
		}
		if err := saveAccount(accountID, account); err != nil {
			return nil, status, err
		}
		if err := saveAccount("default", account); err != nil {
			return nil, status, err
		}
		return &LoginResult{AccountID: accountID, Token: token, BaseURL: respBaseURL, UserID: userID}, status, nil
	case "expired":
		return nil, status, fmt.Errorf("weixin qr code expired; rerun login")
	default:
		return nil, status, nil
	}
}

func ilinkPostMap(ctx context.Context, baseURL, endpoint string, payload map[string]any) (map[string]any, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", strings.TrimRight(baseURL, "/")+"/"+strings.TrimLeft(endpoint, "/"), bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("iLink-App-Id", ilinkAppID)
	req.Header.Set("iLink-App-ClientVersion", fmt.Sprintf("%d", ilinkClientVersion))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		if len(body) > 300 {
			body = body[:300]
		}
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}
	var out map[string]any
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func isTemporaryLoginPollError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "timeout") ||
		strings.Contains(msg, "deadline exceeded") ||
		strings.Contains(msg, "connection reset") ||
		strings.Contains(msg, "eof") ||
		strings.Contains(msg, "524")
}
