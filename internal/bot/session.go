package bot

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"
)

// BuildSessionKey returns a stable key for the remote IM conversation.
//
//   - DM: one shared session per direct chat.
//   - Group/guild: isolate by user inside the group.
//   - Thread: share one session inside the thread.
func BuildSessionKey(src SessionSource) string {
	var scope string
	switch src.ChatType {
	case ChatDM:
		scope = fmt.Sprintf("%s:dm:%s", src.Platform, src.ChatID)
	case ChatGroup:
		scope = fmt.Sprintf("%s:group:%s:%s", src.Platform, src.ChatID, src.UserID)
	case ChatGuild:
		scope = fmt.Sprintf("%s:guild:%s:%s", src.Platform, src.ChatID, src.UserID)
	case ChatDirect:
		scope = fmt.Sprintf("%s:direct:%s", src.Platform, src.ChatID)
	case ChatThread:
		threadID := src.ThreadID
		if threadID == "" {
			threadID = src.ChatID
		}
		scope = fmt.Sprintf("%s:thread:%s", src.Platform, threadID)
	default:
		scope = fmt.Sprintf("%s:%s:%s:%s", src.Platform, src.ChatType, src.ChatID, src.UserID)
	}
	h := sha256.Sum256([]byte(scope))
	return hex.EncodeToString(h[:])[:16]
}

var slashCommands = map[string]bool{
	"/start":    true,
	"/stop":     true,
	"/new":      true,
	"/continue": true,
	"/hi":       true,
	"/reset":    true,
	"/approve":  true,
	"/deny":     true,
	"/answer":   true,
	"/status":   true,
	"/help":     true,
}

// IsSlashBypass reports whether text is a management command that must bypass
// the per-session turn queue.
func IsSlashBypass(text string) bool {
	if len(text) == 0 {
		return false
	}
	cmd := text
	for i, r := range text {
		if r == ' ' {
			cmd = text[:i]
			break
		}
	}
	return slashCommands[cmd]
}

type pendingTurn struct {
	msg       InboundMessage
	timestamp time.Time
}

// SessionManager keeps only one running task per selected bot session.
type SessionManager struct {
	mu       sync.Mutex
	active   map[string]bool
	pending  map[string][]pendingTurn
	debounce time.Duration
}

func NewSessionManager(debounce time.Duration) *SessionManager {
	if debounce <= 0 {
		debounce = 1500 * time.Millisecond
	}
	return &SessionManager{
		active:   make(map[string]bool),
		pending:  make(map[string][]pendingTurn),
		debounce: debounce,
	}
}

func (sm *SessionManager) TryAcquire(key string, msg InboundMessage) (acquired bool, merged bool) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.active[key] {
		if IsSlashBypass(msg.Text) {
			return true, false
		}
		queue := sm.pending[key]
		if len(queue) > 0 {
			last := &queue[len(queue)-1]
			if msg.Text != "" && time.Since(last.timestamp) < sm.debounce {
				if last.msg.Text != "" {
					last.msg.Text = last.msg.Text + "\n" + msg.Text
				} else {
					last.msg.Text = msg.Text
				}
				last.timestamp = time.Now()
				return false, true
			}
		}
		queue = append(queue, pendingTurn{msg: msg, timestamp: time.Now()})
		sm.pending[key] = queue
		return false, true
	}

	sm.active[key] = true
	return true, false
}

func (sm *SessionManager) Release(key string) *InboundMessage {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	queue := sm.pending[key]
	if len(queue) == 0 {
		delete(sm.active, key)
		delete(sm.pending, key)
		return nil
	}

	var merged *InboundMessage
	for i := range queue {
		if merged == nil {
			m := queue[i].msg
			merged = &m
		} else if queue[i].msg.Text != "" {
			merged.Text = strings.TrimSpace(merged.Text + "\n" + queue[i].msg.Text)
		}
	}
	delete(sm.pending, key)
	return merged
}

func (sm *SessionManager) IsActive(key string) bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	return sm.active[key]
}

func (sm *SessionManager) ActiveCount() int {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	return len(sm.active)
}

func (sm *SessionManager) ForceRelease(key string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.active, key)
	delete(sm.pending, key)
}
