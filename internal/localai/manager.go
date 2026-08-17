package localai

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/nanbo0ne/O.R.C.A-for-Windows/internal/fileutil"
)

type TaskKind string
type TaskState string

const (
	TaskModel   TaskKind = "model"
	TaskRuntime TaskKind = "runtime"

	TaskQueued      TaskState = "queued"
	TaskDownloading TaskState = "downloading"
	TaskPaused      TaskState = "paused"
	TaskVerifying   TaskState = "verifying"
	TaskInstalling  TaskState = "installing"
	TaskCompleted   TaskState = "completed"
	TaskFailed      TaskState = "failed"
	TaskCancelled   TaskState = "cancelled"
)

type DownloadTask struct {
	ID              string    `json:"id"`
	Kind            TaskKind  `json:"kind"`
	TargetID        string    `json:"targetId"`
	Label           string    `json:"label"`
	State           TaskState `json:"state"`
	Artifact        string    `json:"artifact,omitempty"`
	DownloadedBytes int64     `json:"downloadedBytes"`
	TotalBytes      int64     `json:"totalBytes"`
	BytesPerSecond  int64     `json:"bytesPerSecond"`
	ETASeconds      int64     `json:"etaSeconds"`
	Source          string    `json:"source,omitempty"`
	Error           string    `json:"error,omitempty"`
	CreatedAt       int64     `json:"createdAt"`
	UpdatedAt       int64     `json:"updatedAt"`
}

type ModelInstallation struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Path        string    `json:"path"`
	ModelPath   string    `json:"modelPath"`
	MMProjPath  string    `json:"mmprojPath,omitempty"`
	Size        int64     `json:"size"`
	InstalledAt time.Time `json:"installedAt"`
	Vision      bool      `json:"vision"`
	ToolUse     bool      `json:"toolUse"`
}

type RuntimeInstallation struct {
	ID          string    `json:"id"`
	Backend     string    `json:"backend"`
	Version     string    `json:"version"`
	Path        string    `json:"path"`
	ServerPath  string    `json:"serverPath"`
	InstalledAt time.Time `json:"installedAt"`
}

type persistedState struct {
	Tasks []DownloadTask `json:"tasks"`
}

type Manager struct {
	mu        sync.Mutex
	root      string
	models    string
	runtimes  string
	downloads string
	tasks     map[string]*DownloadTask
	cancels   map[string]context.CancelFunc
	emit      func(DownloadTask)
	client    *http.Client
}

func DefaultRoot() string {
	dir, err := os.UserCacheDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "O.R.C.A", "local-ai")
}

func DefaultModelsDir() string {
	dir, err := os.UserCacheDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "O.R.C.A", "models")
}

func NewManager(root string, emit func(DownloadTask)) *Manager {
	return NewManagerWithModels(root, "", emit)
}

func NewManagerWithModels(root, models string, emit func(DownloadTask)) *Manager {
	if strings.TrimSpace(root) == "" {
		root = DefaultRoot()
	}
	if strings.TrimSpace(models) == "" {
		models = DefaultModelsDir()
	}
	m := &Manager{
		root: root, models: models, runtimes: filepath.Join(root, "runtimes"), downloads: filepath.Join(root, "downloads"),
		tasks: map[string]*DownloadTask{}, cancels: map[string]context.CancelFunc{}, emit: emit,
		client: &http.Client{Transport: &http.Transport{Proxy: http.ProxyFromEnvironment, MaxIdleConns: 8, IdleConnTimeout: 60 * time.Second}},
	}
	_ = os.MkdirAll(m.downloads, 0o755)
	m.loadState()
	return m
}

func (m *Manager) Root() string { return m.root }

func (m *Manager) Tasks() []DownloadTask {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]DownloadTask, 0, len(m.tasks))
	for _, task := range m.tasks {
		out = append(out, *task)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt > out[j].CreatedAt })
	return out
}

func (m *Manager) StartModelDownload(id string) (DownloadTask, error) {
	spec, ok := ModelByID(id)
	if !ok {
		return DownloadTask{}, fmt.Errorf("unknown local model %q", id)
	}
	return m.start(TaskModel, id, spec.Name, spec.Artifacts)
}

func (m *Manager) StartRuntimeInstall(id string) (DownloadTask, error) {
	spec, ok := RuntimeByID(id)
	if !ok {
		return DownloadTask{}, fmt.Errorf("unknown local runtime %q", id)
	}
	return m.start(TaskRuntime, id, "llama.cpp "+spec.Version+" · "+spec.Backend, spec.Artifacts)
}

func (m *Manager) start(kind TaskKind, targetID, label string, artifacts []Artifact) (DownloadTask, error) {
	if len(artifacts) == 0 {
		return DownloadTask{}, fmt.Errorf("%s has no downloadable artifacts", targetID)
	}
	var total int64
	for _, artifact := range artifacts {
		if artifact.Size <= 0 || len(artifact.SHA256) != 64 || len(artifact.Sources) == 0 {
			return DownloadTask{}, fmt.Errorf("%s has an incomplete signed manifest", artifact.Name)
		}
		total += artifact.Size
	}
	if err := m.ensureDiskSpace(total + 2<<30); err != nil {
		return DownloadTask{}, err
	}
	now := time.Now().UnixMilli()
	task := &DownloadTask{ID: fmt.Sprintf("%s-%s-%d", kind, targetID, now), Kind: kind, TargetID: targetID, Label: label, State: TaskQueued, TotalBytes: total, CreatedAt: now, UpdatedAt: now}
	m.mu.Lock()
	for _, existing := range m.tasks {
		if existing.Kind == kind && existing.TargetID == targetID && (existing.State == TaskQueued || existing.State == TaskDownloading || existing.State == TaskPaused || existing.State == TaskVerifying || existing.State == TaskInstalling) {
			copy := *existing
			m.mu.Unlock()
			return copy, nil
		}
	}
	m.tasks[task.ID] = task
	m.saveStateLocked()
	m.mu.Unlock()
	m.launch(task.ID)
	return *task, nil
}

func (m *Manager) Pause(id string) error {
	m.mu.Lock()
	task := m.tasks[id]
	if task == nil {
		m.mu.Unlock()
		return fmt.Errorf("download task %q not found", id)
	}
	if task.State != TaskDownloading && task.State != TaskQueued {
		m.mu.Unlock()
		return nil
	}
	task.State, task.UpdatedAt = TaskPaused, time.Now().UnixMilli()
	cancel := m.cancels[id]
	m.saveStateLocked()
	copy := *task
	m.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	m.publish(copy)
	return nil
}

func (m *Manager) Resume(id string) error {
	m.mu.Lock()
	task := m.tasks[id]
	if task == nil {
		m.mu.Unlock()
		return fmt.Errorf("download task %q not found", id)
	}
	if task.State != TaskPaused && task.State != TaskFailed {
		m.mu.Unlock()
		return nil
	}
	task.State, task.Error, task.UpdatedAt = TaskQueued, "", time.Now().UnixMilli()
	m.saveStateLocked()
	m.mu.Unlock()
	m.launch(id)
	return nil
}

func (m *Manager) Cancel(id string) error {
	m.mu.Lock()
	task := m.tasks[id]
	if task == nil {
		m.mu.Unlock()
		return fmt.Errorf("download task %q not found", id)
	}
	task.State, task.UpdatedAt = TaskCancelled, time.Now().UnixMilli()
	cancel := m.cancels[id]
	m.saveStateLocked()
	copy := *task
	m.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	m.publish(copy)
	return nil
}

func (m *Manager) launch(id string) {
	ctx, cancel := context.WithCancel(context.Background())
	m.mu.Lock()
	m.cancels[id] = cancel
	m.mu.Unlock()
	go func() {
		err := m.run(ctx, id)
		m.mu.Lock()
		delete(m.cancels, id)
		task := m.tasks[id]
		if task != nil && err != nil && task.State != TaskPaused && task.State != TaskCancelled {
			task.State, task.Error, task.UpdatedAt = TaskFailed, err.Error(), time.Now().UnixMilli()
		}
		m.saveStateLocked()
		var copy DownloadTask
		if task != nil {
			copy = *task
		}
		m.mu.Unlock()
		if task != nil {
			m.publish(copy)
		}
	}()
}

func (m *Manager) run(ctx context.Context, id string) error {
	m.mu.Lock()
	task := m.tasks[id]
	if task == nil {
		m.mu.Unlock()
		return fmt.Errorf("task disappeared")
	}
	task.State, task.UpdatedAt = TaskDownloading, time.Now().UnixMilli()
	kind, targetID := task.Kind, task.TargetID
	m.saveStateLocked()
	m.mu.Unlock()

	var artifacts []Artifact
	if kind == TaskModel {
		spec, _ := ModelByID(targetID)
		artifacts = spec.Artifacts
	} else {
		spec, _ := RuntimeByID(targetID)
		artifacts = spec.Artifacts
	}
	taskDir := filepath.Join(m.downloads, id)
	if err := os.MkdirAll(taskDir, 0o755); err != nil {
		return err
	}
	var completed int64
	for _, artifact := range artifacts {
		part := filepath.Join(taskDir, artifact.Name+".part")
		if ok, _ := fileMatches(part, artifact); ok {
			completed += artifact.Size
			continue
		}
		if err := m.downloadArtifact(ctx, id, artifact, part, completed); err != nil {
			return err
		}
		completed += artifact.Size
	}
	m.setTaskState(id, TaskVerifying, "")
	for _, artifact := range artifacts {
		if ok, err := fileMatches(filepath.Join(taskDir, artifact.Name+".part"), artifact); !ok {
			if err == nil {
				err = fmt.Errorf("checksum mismatch")
			}
			return fmt.Errorf("verify %s: %w", artifact.Name, err)
		}
	}
	m.setTaskState(id, TaskInstalling, "")
	if kind == TaskModel {
		if err := m.installModel(targetID, taskDir, artifacts); err != nil {
			return err
		}
	} else if err := m.installRuntime(targetID, taskDir, artifacts); err != nil {
		return err
	}
	m.mu.Lock()
	if task = m.tasks[id]; task != nil {
		task.State, task.Error, task.DownloadedBytes, task.BytesPerSecond, task.ETASeconds, task.UpdatedAt = TaskCompleted, "", task.TotalBytes, 0, 0, time.Now().UnixMilli()
	}
	m.saveStateLocked()
	var copy DownloadTask
	if task != nil {
		copy = *task
	}
	m.mu.Unlock()
	if task != nil {
		m.publish(copy)
	}
	return nil
}

func (m *Manager) downloadArtifact(ctx context.Context, id string, artifact Artifact, part string, completed int64) error {
	var lastErr error
	for _, source := range artifact.Sources {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := m.downloadFrom(ctx, id, source, artifact, part, completed); err == nil {
			return nil
		} else {
			lastErr = err
		}
	}
	return fmt.Errorf("all mirrors failed for %s: %w", artifact.Name, lastErr)
}

func (m *Manager) downloadFrom(ctx context.Context, id, source string, artifact Artifact, part string, completed int64) error {
	if err := os.MkdirAll(filepath.Dir(part), 0o755); err != nil {
		return err
	}
	offset := int64(0)
	if info, err := os.Stat(part); err == nil {
		offset = info.Size()
		if offset >= artifact.Size {
			_ = os.Remove(part)
			offset = 0
		}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, source, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "Orca/3.0 local-model-manager")
	if offset > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", offset))
	}
	resp, err := m.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	flags := os.O_CREATE | os.O_WRONLY
	if offset > 0 && resp.StatusCode == http.StatusPartialContent {
		flags |= os.O_APPEND
	} else {
		flags |= os.O_TRUNC
		offset = 0
	}
	f, err := os.OpenFile(part, flags, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()
	buf := make([]byte, 256*1024)
	lastTick, lastBytes := time.Now(), offset
	written := offset
	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			if _, err := f.Write(buf[:n]); err != nil {
				return err
			}
			written += int64(n)
		}
		now := time.Now()
		if now.Sub(lastTick) >= 250*time.Millisecond || readErr == io.EOF {
			deltaSec := now.Sub(lastTick).Seconds()
			speed := int64(0)
			if deltaSec > 0 {
				speed = int64(float64(written-lastBytes) / deltaSec)
			}
			m.updateProgress(id, artifact.Name, source, completed+written, speed)
			lastTick, lastBytes = now, written
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return readErr
		}
		if err := ctx.Err(); err != nil {
			return err
		}
	}
	if written != artifact.Size {
		return fmt.Errorf("downloaded %d bytes, expected %d", written, artifact.Size)
	}
	return f.Sync()
}

func (m *Manager) updateProgress(id, artifact, source string, downloaded, speed int64) {
	m.mu.Lock()
	task := m.tasks[id]
	if task == nil {
		m.mu.Unlock()
		return
	}
	task.Artifact, task.Source, task.DownloadedBytes, task.BytesPerSecond, task.UpdatedAt = artifact, source, downloaded, speed, time.Now().UnixMilli()
	remaining := task.TotalBytes - downloaded
	if speed > 0 && remaining > 0 {
		task.ETASeconds = remaining / speed
	} else {
		task.ETASeconds = 0
	}
	copy := *task
	m.mu.Unlock()
	m.publish(copy)
}

func (m *Manager) setTaskState(id string, state TaskState, errText string) {
	m.mu.Lock()
	if task := m.tasks[id]; task != nil {
		task.State, task.Error, task.UpdatedAt = state, errText, time.Now().UnixMilli()
		copy := *task
		m.saveStateLocked()
		m.mu.Unlock()
		m.publish(copy)
		return
	}
	m.mu.Unlock()
}

func (m *Manager) installModel(id, taskDir string, artifacts []Artifact) error {
	target := filepath.Join(m.models, id)
	if err := os.MkdirAll(target, 0o755); err != nil {
		return err
	}
	for _, artifact := range artifacts {
		source := filepath.Join(taskDir, artifact.Name+".part")
		dest := filepath.Join(target, artifact.Name)
		if err := promoteFile(source, dest); err != nil {
			return err
		}
	}
	spec, _ := ModelByID(id)
	return writeJSONAtomic(filepath.Join(target, "installation.json"), ModelInstallation{ID: id, Name: spec.Name, Path: target, InstalledAt: time.Now().UTC(), Vision: spec.Vision, ToolUse: spec.ToolUse})
}

func (m *Manager) installRuntime(id, taskDir string, artifacts []Artifact) error {
	target := filepath.Join(m.runtimes, id)
	stage := target + ".installing"
	_ = os.RemoveAll(stage)
	if err := os.MkdirAll(stage, 0o755); err != nil {
		return err
	}
	for _, artifact := range artifacts {
		if err := extractZip(filepath.Join(taskDir, artifact.Name+".part"), stage); err != nil {
			_ = os.RemoveAll(stage)
			return err
		}
	}
	server, err := findFile(stage, "llama-server.exe")
	if err != nil {
		_ = os.RemoveAll(stage)
		return err
	}
	rel, _ := filepath.Rel(stage, server)
	backup := target + ".previous"
	_ = os.RemoveAll(backup)
	if _, err := os.Stat(target); err == nil {
		if err := os.Rename(target, backup); err != nil {
			return err
		}
	}
	if err := os.Rename(stage, target); err != nil {
		_ = os.Rename(backup, target)
		return err
	}
	_ = os.RemoveAll(backup)
	spec, _ := RuntimeByID(id)
	return writeJSONAtomic(filepath.Join(target, "installation.json"), RuntimeInstallation{ID: id, Backend: spec.Backend, Version: spec.Version, Path: target, ServerPath: filepath.Join(target, rel), InstalledAt: time.Now().UTC()})
}

func (m *Manager) InstalledModels() []ModelInstallation {
	entries, _ := os.ReadDir(m.models)
	var out []ModelInstallation
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		var item ModelInstallation
		if readJSON(filepath.Join(m.models, entry.Name(), "installation.json"), &item) != nil {
			continue
		}
		spec, ok := ModelByID(item.ID)
		if !ok {
			continue
		}
		valid := true
		for _, artifact := range spec.Artifacts {
			if ok, _ := fileMatches(filepath.Join(item.Path, artifact.Name), artifact); !ok {
				valid = false
				break
			}
			item.Size += artifact.Size
			if strings.Contains(strings.ToLower(artifact.Name), "mmproj") {
				item.MMProjPath = filepath.Join(item.Path, artifact.Name)
			} else {
				item.ModelPath = filepath.Join(item.Path, artifact.Name)
			}
		}
		if valid {
			out = append(out, item)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func (m *Manager) InstalledRuntime() (RuntimeInstallation, bool) {
	entries, _ := os.ReadDir(m.runtimes)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		var item RuntimeInstallation
		if readJSON(filepath.Join(m.runtimes, entry.Name(), "installation.json"), &item) == nil {
			if info, err := os.Stat(item.ServerPath); err == nil && !info.IsDir() {
				return item, true
			}
		}
	}
	return RuntimeInstallation{}, false
}

func (m *Manager) DeleteModel(id string) error {
	if _, ok := ModelByID(id); !ok {
		return fmt.Errorf("unknown local model %q", id)
	}
	return removeOwnedDir(m.models, filepath.Join(m.models, id))
}

func (m *Manager) UninstallRuntime() error {
	if err := removeOwnedDir(m.root, m.runtimes); err != nil {
		return err
	}
	return os.MkdirAll(m.runtimes, 0o755)
}

func (m *Manager) ensureDiskSpace(required int64) error {
	profile := DetectHardware(m.root)
	if profile.DiskFreeBytes > 0 && profile.DiskFreeBytes < required {
		return fmt.Errorf("insufficient disk space: need %.1f GB including safety margin, have %.1f GB", float64(required)/(1<<30), float64(profile.DiskFreeBytes)/(1<<30))
	}
	return nil
}

func (m *Manager) loadState() {
	var state persistedState
	if readJSON(filepath.Join(m.root, "downloads.json"), &state) != nil {
		return
	}
	for i := range state.Tasks {
		task := state.Tasks[i]
		if task.State == TaskDownloading || task.State == TaskQueued || task.State == TaskVerifying || task.State == TaskInstalling {
			task.State = TaskPaused
			task.Error = "应用已重启，可继续下载"
		}
		copy := task
		m.tasks[copy.ID] = &copy
	}
}

func (m *Manager) saveStateLocked() {
	state := persistedState{Tasks: make([]DownloadTask, 0, len(m.tasks))}
	for _, task := range m.tasks {
		state.Tasks = append(state.Tasks, *task)
	}
	_ = writeJSONAtomic(filepath.Join(m.root, "downloads.json"), state)
}

func (m *Manager) publish(task DownloadTask) {
	if m.emit != nil {
		m.emit(task)
	}
}

func fileMatches(path string, artifact Artifact) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	if info.Size() != artifact.Size {
		return false, nil
	}
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return false, err
	}
	return strings.EqualFold(hex.EncodeToString(h.Sum(nil)), artifact.SHA256), nil
}

func promoteFile(source, target string) error {
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	_ = os.Remove(target)
	if err := os.Rename(source, target); err == nil {
		return nil
	}
	in, err := os.Open(source)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(target + ".tmp")
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}
	if err := os.Rename(target+".tmp", target); err != nil {
		return err
	}
	return os.Remove(source)
}

func extractZip(path, target string) error {
	r, err := zip.OpenReader(path)
	if err != nil {
		return err
	}
	defer r.Close()
	root, _ := filepath.Abs(target)
	for _, file := range r.File {
		dest := filepath.Join(target, filepath.FromSlash(file.Name))
		abs, _ := filepath.Abs(dest)
		if abs != root && !strings.HasPrefix(abs, root+string(os.PathSeparator)) {
			return fmt.Errorf("zip entry escapes install root: %s", file.Name)
		}
		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(dest, 0o755); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return err
		}
		in, err := file.Open()
		if err != nil {
			return err
		}
		out, err := os.OpenFile(dest, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, file.Mode())
		if err != nil {
			_ = in.Close()
			return err
		}
		_, copyErr := io.Copy(out, in)
		_ = in.Close()
		_ = out.Close()
		if copyErr != nil {
			return copyErr
		}
	}
	return nil
}

func findFile(root, name string) (string, error) {
	var found string
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() && strings.EqualFold(entry.Name(), name) {
			found = path
			return io.EOF
		}
		return nil
	})
	if errors.Is(err, io.EOF) {
		err = nil
	}
	if err != nil {
		return "", err
	}
	if found == "" {
		return "", fmt.Errorf("%s not found in runtime package", name)
	}
	return found, nil
}

func removeOwnedDir(root, target string) error {
	rootAbs, _ := filepath.Abs(root)
	targetAbs, _ := filepath.Abs(target)
	if targetAbs == rootAbs || !strings.HasPrefix(targetAbs, rootAbs+string(os.PathSeparator)) {
		return fmt.Errorf("refusing to remove path outside local AI root")
	}
	return os.RemoveAll(targetAbs)
}

func writeJSONAtomic(path string, value any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	body, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".state-*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	if _, err := tmp.Write(append(body, '\n')); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return fileutil.ReplaceFile(tmpPath, path)
}

func readJSON(path string, target any) error {
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, target)
}
