package computeruse

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrNotSupported     = errors.New("computer use is not supported on this platform")
	ErrNotRunning       = errors.New("computer use session is not running")
	ErrStaleObservation = errors.New("the observation is stale; observe again before acting")
	ErrProtectedSurface = errors.New("the target is protected and cannot be automated")
	ErrUserInputPaused  = errors.New("computer use paused because the user took control")
)

type Backend interface {
	Capabilities() Capabilities
	Observe(ctx context.Context, sessionID string, generation uint64) (Observation, error)
	Execute(ctx context.Context, observation Observation, action Action) error
	StartSafetyHooks(onEmergencyStop func(), onUserInput func()) error
	StopSafetyHooks()
	ReleaseInjectedInput()
	ShowOverlay(OverlayState) error
	UpdateOverlay(OverlayState)
	HideOverlay()
}

type Event struct {
	Kind        string        `json:"kind"`
	Session     Session       `json:"session"`
	Observation *Observation  `json:"observation,omitempty"`
	Action      *ActionResult `json:"action,omitempty"`
}

type Service struct {
	mu          sync.Mutex
	backend     Backend
	session     Session
	observation Observation
	generation  uint64
	cancel      context.CancelFunc
	emit        func(Event)
}

func NewService(backend Backend, emit func(Event)) *Service {
	return &Service{backend: backend, emit: emit, session: Session{State: StateIdle, Logs: []ActionLog{}}}
}

func (s *Service) Capabilities() Capabilities {
	if s.backend == nil {
		return Capabilities{Supported: false, UnavailableReason: ErrNotSupported.Error()}
	}
	return s.backend.Capabilities()
}

func (s *Service) Current() Session {
	s.mu.Lock()
	defer s.mu.Unlock()
	return cloneSession(s.session)
}

func (s *Service) Start(ctx context.Context, request StartRequest) (Session, error) {
	if s.backend == nil || !s.backend.Capabilities().Supported {
		return Session{}, ErrNotSupported
	}
	if request.Goal == "" {
		return Session{}, fmt.Errorf("computer task goal is required")
	}
	_ = s.Stop("superseded")
	id, err := sessionID()
	if err != nil {
		return Session{}, err
	}
	runCtx, cancel := context.WithCancel(ctx)
	now := time.Now().UTC()
	s.mu.Lock()
	s.cancel = cancel
	s.generation = 0
	s.observation = Observation{}
	s.session = Session{ID: id, TabID: request.TabID, Goal: request.Goal, SuccessCriteria: request.SuccessCriteria, Restrictions: request.Restrictions, ModelRef: request.ModelRef, State: StateRunning, StartedAt: now, UpdatedAt: now, Logs: []ActionLog{}}
	session := cloneSession(s.session)
	s.mu.Unlock()
	if err := s.backend.StartSafetyHooks(func() { _ = s.Stop("escape") }, s.pauseForUserInput); err != nil {
		cancel()
		return s.finish(StateFailed, err.Error()), err
	}
	if err := s.backend.ShowOverlay(OverlayState{State: StateRunning, Action: "正在观察屏幕"}); err != nil {
		s.backend.StopSafetyHooks()
		cancel()
		return s.finish(StateFailed, err.Error()), err
	}
	s.publish(Event{Kind: "started", Session: session})
	if _, err := s.Observe(runCtx); err != nil {
		_ = s.Stop("observation failed")
		return s.Current(), err
	}
	return s.Current(), nil
}

func (s *Service) Observe(ctx context.Context) (Observation, error) {
	s.mu.Lock()
	if s.session.State != StateRunning && s.session.State != StatePaused {
		s.mu.Unlock()
		return Observation{}, ErrNotRunning
	}
	s.generation++
	generation, sessionID := s.generation, s.session.ID
	s.mu.Unlock()
	observation, err := s.backend.Observe(ctx, sessionID, generation)
	if err != nil {
		return Observation{}, err
	}
	if observation.SecureDesktop {
		return Observation{}, ErrProtectedSurface
	}
	observation.SessionID = sessionID
	observation.Generation = generation
	if observation.ObservedAt.IsZero() {
		observation.ObservedAt = time.Now().UTC()
	}
	s.mu.Lock()
	if s.session.ID != sessionID || (s.session.State != StateRunning && s.session.State != StatePaused) {
		s.mu.Unlock()
		return Observation{}, ErrNotRunning
	}
	s.observation = observation
	s.session.CurrentApp = observation.Foreground.Title
	s.session.UpdatedAt = time.Now().UTC()
	session := cloneSession(s.session)
	s.mu.Unlock()
	public := observation
	public.Screenshot = ""
	s.publish(Event{Kind: "observation", Session: session, Observation: &public})
	return observation, nil
}

func (s *Service) Execute(ctx context.Context, action Action) (ActionResult, error) {
	s.mu.Lock()
	if s.session.State == StatePaused {
		s.mu.Unlock()
		return ActionResult{}, ErrUserInputPaused
	}
	if s.session.State != StateRunning {
		s.mu.Unlock()
		return ActionResult{}, ErrNotRunning
	}
	if action.Generation == 0 || action.Generation != s.observation.Generation {
		s.mu.Unlock()
		return ActionResult{}, ErrStaleObservation
	}
	observation := s.observation
	s.session.CurrentAction = action.Description
	if s.session.CurrentAction == "" {
		s.session.CurrentAction = action.Type
	}
	s.session.UpdatedAt = time.Now().UTC()
	session := cloneSession(s.session)
	s.mu.Unlock()
	s.backend.UpdateOverlay(OverlayState{State: StateRunning, App: observation.Foreground.Title, Action: session.CurrentAction})

	started := time.Now()
	err := s.backend.Execute(ctx, observation, action)
	result := ActionResult{ActionID: fmt.Sprintf("action-%d", started.UnixNano()), Action: action, Success: err == nil, DurationMS: time.Since(started).Milliseconds()}
	if err != nil {
		result.Error = err.Error()
	}
	log := ActionLog{ID: result.ActionID, Type: action.Type, Window: observation.Foreground.Title, StartedAt: started.UTC(), DurationMS: result.DurationMS, Success: result.Success, Error: result.Error}
	s.mu.Lock()
	if s.session.State == StateRunning {
		s.session.ActionCount++
		s.session.Logs = append(s.session.Logs, log)
		s.session.UpdatedAt = time.Now().UTC()
	}
	session = cloneSession(s.session)
	s.mu.Unlock()
	if err != nil {
		s.publish(Event{Kind: "action", Session: session, Action: &result})
		return result, err
	}
	// Every successful action invalidates the previous element IDs and coordinates.
	next, observeErr := s.Observe(ctx)
	result.Observation = next
	if observeErr != nil {
		result.Success, result.Error = false, observeErr.Error()
		err = observeErr
	}
	s.publish(Event{Kind: "action", Session: s.Current(), Action: &result})
	return result, err
}

func (s *Service) Pause() (Session, error) {
	s.mu.Lock()
	if s.session.State != StateRunning {
		s.mu.Unlock()
		return Session{}, ErrNotRunning
	}
	s.session.State, s.session.UpdatedAt = StatePaused, time.Now().UTC()
	session := cloneSession(s.session)
	s.mu.Unlock()
	s.backend.ReleaseInjectedInput()
	s.backend.UpdateOverlay(OverlayState{State: StatePaused, App: session.CurrentApp, Action: "已暂停"})
	s.publish(Event{Kind: "paused", Session: session})
	return session, nil
}

func (s *Service) Resume(ctx context.Context) (Session, error) {
	s.mu.Lock()
	if s.session.State != StatePaused {
		s.mu.Unlock()
		return Session{}, ErrNotRunning
	}
	s.session.State, s.session.UpdatedAt = StateRunning, time.Now().UTC()
	session := cloneSession(s.session)
	s.mu.Unlock()
	s.backend.UpdateOverlay(OverlayState{State: StateRunning, App: session.CurrentApp, Action: "正在重新观察"})
	if _, err := s.Observe(ctx); err != nil {
		return s.finish(StateFailed, err.Error()), err
	}
	s.publish(Event{Kind: "resumed", Session: s.Current()})
	return s.Current(), nil
}

func (s *Service) Complete(success bool, message string) Session {
	if success {
		return s.finish(StateSucceeded, message)
	}
	return s.finish(StateFailed, message)
}

func (s *Service) Stop(reason string) error {
	s.mu.Lock()
	if s.session.State == StateIdle || s.session.State == StateCancelled || s.session.State == StateSucceeded || s.session.State == StateFailed {
		s.mu.Unlock()
		return nil
	}
	s.session.State, s.session.CurrentAction, s.session.UpdatedAt = StateStopping, "正在停止", time.Now().UTC()
	cancel := s.cancel
	session := cloneSession(s.session)
	s.mu.Unlock()
	s.publish(Event{Kind: "stopping", Session: session})
	if cancel != nil {
		cancel()
	}
	s.backend.ReleaseInjectedInput()
	s.backend.StopSafetyHooks()
	s.backend.HideOverlay()
	s.finish(StateCancelled, reason)
	return nil
}

func (s *Service) finish(state SessionState, message string) Session {
	now := time.Now().UTC()
	s.mu.Lock()
	if s.cancel != nil {
		s.cancel()
		s.cancel = nil
	}
	s.session.State, s.session.UpdatedAt, s.session.CompletedAt = state, now, &now
	if state == StateFailed {
		s.session.LastError = message
	}
	session := cloneSession(s.session)
	s.mu.Unlock()
	if s.backend != nil {
		s.backend.ReleaseInjectedInput()
		s.backend.StopSafetyHooks()
		s.backend.HideOverlay()
	}
	s.publish(Event{Kind: string(state), Session: session})
	return session
}

func (s *Service) pauseForUserInput() {
	_, _ = s.Pause()
}

func (s *Service) publish(event Event) {
	if s.emit != nil {
		s.emit(event)
	}
}

func cloneSession(in Session) Session {
	in.Logs = append([]ActionLog(nil), in.Logs...)
	return in
}

func sessionID() (string, error) {
	body := make([]byte, 12)
	if _, err := rand.Read(body); err != nil {
		return "", err
	}
	return "computer-" + hex.EncodeToString(body), nil
}
