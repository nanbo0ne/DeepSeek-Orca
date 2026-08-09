package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync"
	"testing"
	"time"

	"deepseek-orca/internal/config"
	"deepseek-orca/internal/provider"
	"deepseek-orca/internal/visioncap"
)

func TestConcurrentConfigWritesPreserveBothChanges(t *testing.T) {
	isolateDesktopUserDirs(t)
	if err := config.Default().SaveTo(config.UserConfigPath()); err != nil {
		t.Fatal(err)
	}

	app := NewApp()
	firstEntered := make(chan struct{})
	releaseFirst := make(chan struct{})
	errs := make(chan error, 2)
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		errs <- app.applyConfigOnly(func(c *config.Config) error {
			close(firstEntered)
			<-releaseFirst
			return c.SetDesktopUIScale(90)
		})
	}()
	<-firstEntered
	go func() {
		defer wg.Done()
		errs <- app.applyConfigOnly(func(c *config.Config) error {
			return c.SetDesktopCheckUpdates(false)
		})
	}()

	// Give the second writer time to contend while the first writer owns the
	// read-modify-write transaction. Without serialization it reads stale data.
	time.Sleep(30 * time.Millisecond)
	close(releaseFirst)
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}

	got := config.LoadForEdit(config.UserConfigPath())
	if got.DesktopUIScale() != 90 || got.DesktopCheckUpdates() {
		t.Fatalf("concurrent settings update lost a change: scale=%d checkUpdates=%v", got.DesktopUIScale(), got.DesktopCheckUpdates())
	}
}

func TestWithFreshSystemPromptReplacesExistingSystemMessage(t *testing.T) {
	msgs := []provider.Message{
		{Role: provider.RoleSystem, Content: "old", ReasoningContent: "stale", ReasoningSignature: "sig", ToolCalls: []provider.ToolCall{{ID: "call", Name: "noop"}}, ToolCallID: "tool", Name: "name"},
		{Role: provider.RoleUser, Content: "hello"},
	}

	got := withFreshSystemPrompt(msgs, "new")
	if got[0].Content != "new" {
		t.Fatalf("system prompt = %q, want new", got[0].Content)
	}
	if got[0].ReasoningContent != "" || got[0].ReasoningSignature != "" || len(got[0].ToolCalls) != 0 || got[0].ToolCallID != "" || got[0].Name != "" {
		t.Fatalf("system metadata should be cleared, got %+v", got[0])
	}
	if got[1].Content != "hello" {
		t.Fatalf("non-system message changed: %+v", got[1])
	}
	if msgs[0].Content != "old" {
		t.Fatalf("input slice was mutated: %+v", msgs[0])
	}
}

func TestWithFreshSystemPromptPrependsMissingSystemMessage(t *testing.T) {
	msgs := []provider.Message{{Role: provider.RoleUser, Content: "hello"}}

	got := withFreshSystemPrompt(msgs, "new")
	if len(got) != 2 || got[0].Role != provider.RoleSystem || got[0].Content != "new" {
		t.Fatalf("expected prepended system prompt, got %+v", got)
	}
	if got[1].Content != "hello" {
		t.Fatalf("existing user message changed: %+v", got[1])
	}
}

func TestProviderViewFromEntry_FiltersNonChatModels(t *testing.T) {
	p := config.ProviderEntry{
		Name: "mimo-api",
		Models: []string{
			"mimo-v2", "mimo-v2-pro",
			"mimo-v2-asr", "mimo-v2-tts",
			"mimo-v2-tts-voiceclone", "mimo-v2-tts-voicedesign",
		},
	}
	view := providerViewFromEntry(p, true, false)
	want := []string{"mimo-v2", "mimo-v2-pro"}
	if !reflect.DeepEqual(view.Models, want) {
		t.Errorf("ProviderView.Models = %v, want %v", view.Models, want)
	}
}

func TestMimoAPITemplateEnablesRegularAndProModels(t *testing.T) {
	entries, _, err := officialProviderTemplate("mimo-api")
	if err != nil || len(entries) != 1 {
		t.Fatalf("officialProviderTemplate(mimo-api) = %+v, %v", entries, err)
	}
	want := []string{"mimo-v2.5", "mimo-v2.5-pro"}
	if got := entries[0].ModelList(); !reflect.DeepEqual(got, want) {
		t.Fatalf("Mimo API models = %v, want %v", got, want)
	}
	if got := entries[0].DefaultModel(); got != "mimo-v2.5-pro" {
		t.Fatalf("Mimo API default = %q, want mimo-v2.5-pro", got)
	}
}

func TestFetchProviderModelsFiltersNonChatModels(t *testing.T) {
	t.Setenv("TEST_PROVIDER_KEY", "test-key")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/models" {
			http.NotFound(w, r)
			return
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"object": "list",
			"data": []map[string]string{
				{"id": "mimo-v2.5-pro", "object": "model"},
				{"id": "mimo-v2.5-asr", "object": "model"},
				{"id": "mimo-v2.5-tts", "object": "model"},
			},
		})
	}))
	defer srv.Close()

	got, err := NewApp().FetchProviderModels(ProviderView{
		Name:      "mimo-api",
		BaseURL:   srv.URL,
		APIKeyEnv: "TEST_PROVIDER_KEY",
	})
	if err != nil {
		t.Fatalf("FetchProviderModels: %v", err)
	}
	want := []string{"mimo-v2.5-pro"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("FetchProviderModels = %v, want %v", got, want)
	}
}

func TestSaveProviderFiltersNonChatModels(t *testing.T) {
	isolateDesktopUserDirs(t)

	app := NewApp()
	if err := app.SaveProvider(ProviderView{
		Name:      "mimo-api",
		Kind:      "openai",
		BaseURL:   "https://api.xiaomimimo.com/v1",
		Models:    []string{"mimo-v2.5-asr", "mimo-v2.5-pro", "mimo-v2.5-tts"},
		Default:   "mimo-v2.5-asr",
		APIKeyEnv: "MIMO_API_KEY",
	}); err != nil {
		t.Fatalf("SaveProvider: %v", err)
	}

	cfg := config.LoadForEdit(config.UserConfigPath())
	got, ok := cfg.Provider("mimo-api")
	if !ok {
		t.Fatal("saved provider not found")
	}
	want := []string{"mimo-v2.5-pro"}
	if !reflect.DeepEqual(got.ModelList(), want) {
		t.Errorf("saved provider models = %v, want %v", got.ModelList(), want)
	}
	if got.DefaultModel() != "mimo-v2.5-pro" {
		t.Errorf("saved provider default = %q, want mimo-v2.5-pro", got.DefaultModel())
	}
}

func TestSetAgentParamsPersistsStepLimitsToUserConfig(t *testing.T) {
	isolateDesktopUserDirs(t)

	app := NewApp()
	if err := app.SetAgentParams(0.35, 37, 9, 0.45, 0.75, 0.88, "custom system"); err != nil {
		t.Fatalf("SetAgentParams: %v", err)
	}

	view := app.Settings()
	if view.Agent.MaxSteps != 37 || view.Agent.PlannerMaxSteps != 9 {
		t.Fatalf("Settings().Agent = %+v, want maxSteps=37 plannerMaxSteps=9", view.Agent)
	}
	if view.Agent.Temperature != 0.35 || view.Agent.SystemPrompt != "custom system" {
		t.Fatalf("Settings().Agent did not preserve other agent params: %+v", view.Agent)
	}
	if view.Agent.SoftCompactRatio != 0.45 || view.Agent.CompactRatio != 0.75 || view.Agent.CompactForceRatio != 0.88 {
		t.Fatalf("Settings().Agent compact ratios = %+v, want 0.45/0.75/0.88", view.Agent)
	}

	cfg := config.LoadForEdit(config.UserConfigPath())
	if cfg.Agent.MaxSteps != 37 || cfg.Agent.PlannerMaxSteps != 9 {
		t.Fatalf("saved config agent steps = max:%d planner:%d, want 37/9", cfg.Agent.MaxSteps, cfg.Agent.PlannerMaxSteps)
	}
	if cfg.Agent.Temperature != 0.35 || cfg.Agent.SystemPrompt != "custom system" {
		t.Fatalf("saved config did not preserve other agent params: %+v", cfg.Agent)
	}
	if cfg.Agent.SoftCompactRatio != 0.45 || cfg.Agent.CompactRatio != 0.75 || cfg.Agent.CompactForceRatio != 0.88 {
		t.Fatalf("saved config compact ratios = %.2f/%.2f/%.2f, want 0.45/0.75/0.88", cfg.Agent.SoftCompactRatio, cfg.Agent.CompactRatio, cfg.Agent.CompactForceRatio)
	}
}

func TestSetBotSettingsKeepsAssistantModeAndAutomationModel(t *testing.T) {
	isolateDesktopUserDirs(t)

	app := NewApp()
	cfg := config.LoadForEdit(config.UserConfigPath())
	cfg.Bot.Model = "deepseek/deepseek-v4-pro"
	if err := cfg.SaveTo(config.UserConfigPath()); err != nil {
		t.Fatal(err)
	}
	view := app.Settings()
	bot := view.Bot
	bot.Model = "stale-provider/stale-model"
	bot.PromptMode = promptModeEnhanced
	bot.WorkspaceRoot = t.TempDir()
	if err := app.SetBotSettings(bot); err != nil {
		t.Fatalf("SetBotSettings: %v", err)
	}

	got := app.Settings().Bot
	if got.PromptMode != promptModeAssistant {
		t.Fatalf("Settings().Bot.PromptMode = %q, want %q", got.PromptMode, promptModeAssistant)
	}
	cfg = config.LoadForEdit(config.UserConfigPath())
	if cfg.Bot.PromptMode != promptModeAssistant {
		t.Fatalf("saved bot.prompt_mode = %q, want %q", cfg.Bot.PromptMode, promptModeAssistant)
	}
	if cfg.Bot.Model != "deepseek/deepseek-v4-pro" {
		t.Fatalf("saved bot.model = %q, want dedicated automation model preserved", cfg.Bot.Model)
	}
	if cfg.Bot.WorkspaceRoot != "" {
		t.Fatalf("saved bot.workspace_root = %q, want Automation Workspace root", cfg.Bot.WorkspaceRoot)
	}
}

func TestAutomationBotModeMigratesToAssistant(t *testing.T) {
	isolateDesktopUserDirs(t)

	cfg := config.LoadForEdit(config.UserConfigPath())
	cfg.Bot.PromptMode = promptModeEnhanced
	cfg.Bot.WorkspaceRoot = t.TempDir()
	if err := cfg.SaveTo(config.UserConfigPath()); err != nil {
		t.Fatal(err)
	}

	app := NewApp()
	app.migrateDesktopBotPromptMode()
	got := config.LoadForEdit(config.UserConfigPath())
	if got.Bot.PromptMode != promptModeAssistant {
		t.Fatalf("migrated bot.prompt_mode = %q, want %q", got.Bot.PromptMode, promptModeAssistant)
	}
	if got.Bot.WorkspaceRoot != "" {
		t.Fatalf("migrated bot.workspace_root = %q, want empty legacy override", got.Bot.WorkspaceRoot)
	}
}

func TestSetDesktopCheckUpdatesPersistsToUserConfig(t *testing.T) {
	isolateDesktopUserDirs(t)

	app := NewApp()
	if !app.Settings().CheckUpdates {
		t.Fatal("Settings().CheckUpdates default = false, want true")
	}
	if err := app.SetDesktopCheckUpdates(false); err != nil {
		t.Fatalf("SetDesktopCheckUpdates: %v", err)
	}
	view := app.Settings()
	if view.CheckUpdates {
		t.Fatal("Settings().CheckUpdates = true, want false")
	}
	cfg := config.LoadForEdit(config.UserConfigPath())
	if cfg.Desktop.CheckUpdates == nil || *cfg.Desktop.CheckUpdates {
		t.Fatalf("desktop.check_updates = %+v, want false", cfg.Desktop.CheckUpdates)
	}
	if cfg.DesktopCheckUpdates() {
		t.Fatal("DesktopCheckUpdates() = true, want false")
	}
}

func TestFetchProviderModelsStoresVisionMetadata(t *testing.T) {
	isolateDesktopUserDirs(t)
	t.Setenv("TEST_PROVIDER_KEY", "test-key")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{{
			"id": "vision-model", "input_modalities": []string{"text", "image"},
		}}})
	}))
	defer srv.Close()

	models, err := NewApp().FetchProviderModels(ProviderView{Name: "proxy", Kind: "openai", BaseURL: srv.URL, APIKeyEnv: "TEST_PROVIDER_KEY"})
	if err != nil || !reflect.DeepEqual(models, []string{"vision-model"}) {
		t.Fatalf("models=%v err=%v", models, err)
	}
	e := &config.ProviderEntry{Name: "proxy", Kind: "openai", BaseURL: srv.URL, Model: "vision-model"}
	got := visioncap.Load("").Get(e)
	if got.Status != visioncap.Supported || got.Source != visioncap.SourceMetadata {
		t.Fatalf("vision metadata capability = %+v", got)
	}
}

func TestUpdateProviderModelsPreservesLatestProviderFields(t *testing.T) {
	isolateDesktopUserDirs(t)
	cfg := config.Default()
	cfg.Providers = []config.ProviderEntry{{
		Name: "proxy", Kind: "openai", BaseURL: "https://new.example/v1",
		Model: "old", APIKeyEnv: "PROXY_KEY", BalanceURL: "https://new.example/balance",
	}}
	cfg.Desktop.ProviderAccess = []string{"proxy"}
	if err := cfg.SaveTo(config.UserConfigPath()); err != nil {
		t.Fatal(err)
	}

	if err := NewApp().UpdateProviderModels("proxy", []string{"new-a", "new-b"}, "new-b"); err != nil {
		t.Fatal(err)
	}
	got := config.LoadForEdit(config.UserConfigPath())
	p, ok := got.Provider("proxy")
	if !ok {
		t.Fatal("provider disappeared after model refresh")
	}
	if p.BaseURL != "https://new.example/v1" || p.BalanceURL != "https://new.example/balance" || p.APIKeyEnv != "PROXY_KEY" {
		t.Fatalf("provider fields changed during model refresh: %+v", p)
	}
	if !reflect.DeepEqual(p.Models, []string{"new-a", "new-b"}) || p.Default != "new-b" {
		t.Fatalf("provider models = %+v default=%q", p.Models, p.Default)
	}
}

func TestUpdateProviderModelsRejectsEmptyDiscovery(t *testing.T) {
	isolateDesktopUserDirs(t)
	cfg := config.Default()
	cfg.Providers = []config.ProviderEntry{{Name: "proxy", Kind: "openai", BaseURL: "https://example.test", Model: "keep"}}
	if err := cfg.SaveTo(config.UserConfigPath()); err != nil {
		t.Fatal(err)
	}
	if err := NewApp().UpdateProviderModels("proxy", nil, ""); err == nil {
		t.Fatal("empty model refresh should fail")
	}
	p, _ := config.LoadForEdit(config.UserConfigPath()).Provider("proxy")
	if p == nil || p.Model != "keep" {
		t.Fatalf("empty refresh changed provider: %+v", p)
	}
}

func TestVisionCapabilityOverrideCanReturnToAutomaticResult(t *testing.T) {
	isolateDesktopUserDirs(t)
	cfg := config.Default()
	cfg.Providers = []config.ProviderEntry{{Name: "proxy", Kind: "openai", BaseURL: "https://example.test/v1", Model: "vision"}}
	if err := cfg.SaveTo(config.UserConfigPath()); err != nil {
		t.Fatal(err)
	}
	e := &cfg.Providers[0]
	if err := visioncap.Load("").Put(visioncap.Capability{
		ModelRef: visioncap.ModelRef(e), Key: visioncap.Key(e), Status: visioncap.Unsupported,
		Source: visioncap.SourceProbe, Override: visioncap.OverrideAuto, ProbeVersion: visioncap.CurrentProbeVersion,
	}); err != nil {
		t.Fatal(err)
	}
	app := NewApp()
	manual, err := app.SetVisionCapabilityOverride("proxy/vision", visioncap.Supported)
	if err != nil || manual.Status != visioncap.Supported || manual.Source != visioncap.SourceManual {
		t.Fatalf("manual override = %+v err=%v", manual, err)
	}
	automatic, err := app.SetVisionCapabilityOverride("proxy/vision", visioncap.OverrideAuto)
	if err != nil || automatic.Status != visioncap.Unsupported || automatic.Source != visioncap.SourceProbe {
		t.Fatalf("automatic result = %+v err=%v", automatic, err)
	}
}
