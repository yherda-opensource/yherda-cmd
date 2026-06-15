package cmd

import (
	"testing"

	"github.com/yherda-opensource/yherda-cmd/internal/config"
)

func withTempHome(t *testing.T) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
}

func saveContext(t *testing.T, cfg *config.Config) {
	t.Helper()
	if err := config.SaveConfig(cfg); err != nil {
		t.Fatalf("saveContext: %v", err)
	}
}

// saveContextWithCreds saves config and stub credentials so mustClient() doesn't os.Exit.
// Use this in tests where context is satisfied and execution reaches the API call.
// The network call will fail, but we only care that the context logic behaved correctly.
func saveContextWithCreds(t *testing.T, cfg *config.Config) {
	t.Helper()
	saveContext(t, cfg)
	if err := config.SaveCredentials(&config.Credentials{
		AccessToken: "test-token", TokenType: "Bearer",
	}); err != nil {
		t.Fatalf("SaveCredentials: %v", err)
	}
}

// --- person list ---

func TestPersonList_NoIdeaContext_FallsBackToIdeaList(t *testing.T) {
	withTempHome(t)
	// Stub creds so mustClient() doesn't os.Exit; network call will fail but that's fine.
	saveContextWithCreds(t, &config.Config{ActiveWorkspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	rootCmd.SetArgs([]string{"person", "list"})
	// Should not return the old "no active idea" hard error — falls back to listing ideas.
	err := rootCmd.Execute()
	if err != nil && err.Error() == "no active idea — run: yherda ideas use <id>" {
		t.Error("should fall back to listing ideas rather than returning a hard error")
	}
}

func TestPersonList_FlagProvided_NoContextError(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Config{ActiveWorkspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	rootCmd.SetArgs([]string{"person", "list", "--idea", "some-id"})
	err := rootCmd.Execute()
	if err != nil && err.Error() == "no active idea — run: yherda ideas use <id>" {
		t.Error("flag should have satisfied the idea requirement")
	}
}

func TestPersonList_ContextIdeaUsed_NoContextError(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Config{ActiveWorkspace: "ws", APIServer: "https://ws.yherda.test:8000", ActiveIdea: "idea-1"})

	rootCmd.SetArgs([]string{"person", "list"})
	err := rootCmd.Execute()
	if err != nil && err.Error() == "no active idea — run: yherda ideas use <id>" {
		t.Error("active_idea in config should have satisfied the requirement")
	}
}

// --- identity list ---

func TestIdentityList_NoPersonContext_FallsBack(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Config{ActiveWorkspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	rootCmd.SetArgs([]string{"identity", "list"})
	err := rootCmd.Execute()
	if err != nil && err.Error() == "no active person — run: yherda person use <id>" {
		t.Error("should fall back to listing rather than returning a hard error")
	}
}

func TestIdentityList_ContextPersonUsed_NoContextError(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Config{ActiveWorkspace: "ws", APIServer: "https://ws.yherda.test:8000", ActivePerson: "person-1"})

	rootCmd.SetArgs([]string{"identity", "list"})
	err := rootCmd.Execute()
	if err != nil && err.Error() == "no active person — run: yherda person use <id>" {
		t.Error("active_person in config should have satisfied the requirement")
	}
}

// --- arc list ---

func TestArcList_NoPersonOrIdeaContext_FallsBack(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Config{ActiveWorkspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	rootCmd.SetArgs([]string{"arc", "list"})
	err := rootCmd.Execute()
	if err != nil && err.Error() == "no active person — run: yherda person use <id>" {
		t.Error("should fall back to listing rather than returning a hard error")
	}
}

func TestArcList_ContextPersonUsed_NoContextError(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Config{ActiveWorkspace: "ws", APIServer: "https://ws.yherda.test:8000", ActivePerson: "person-1"})

	rootCmd.SetArgs([]string{"arc", "list"})
	err := rootCmd.Execute()
	if err != nil && err.Error() == "no active person — run: yherda person use <id>" {
		t.Error("active_person in config should have satisfied the requirement")
	}
}

func TestArcList_IdeaFlagBypassesPersonContext(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Config{ActiveWorkspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	rootCmd.SetArgs([]string{"arc", "list", "--idea", "idea-1"})
	err := rootCmd.Execute()
	if err != nil && err.Error() == "no active person — run: yherda person use <id>" {
		t.Error("--idea flag should bypass person context requirement")
	}
}

// --- beat list ---

func TestBeatList_NoArcContext_FallsBack(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Config{ActiveWorkspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	rootCmd.SetArgs([]string{"beat", "list"})
	err := rootCmd.Execute()
	if err != nil && err.Error() == "no active arc — run: yherda arc use <id>" {
		t.Error("should fall back to listing rather than returning a hard error")
	}
}

func TestBeatList_ContextArcUsed_NoContextError(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Config{ActiveWorkspace: "ws", APIServer: "https://ws.yherda.test:8000", ActiveArc: "arc-1"})

	rootCmd.SetArgs([]string{"beat", "list"})
	err := rootCmd.Execute()
	if err != nil && err.Error() == "no active arc — run: yherda arc use <id>" {
		t.Error("active_arc in config should have satisfied the requirement")
	}
}

// --- place list ---

func TestPlaceList_NoIdeaContext_FallsBack(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Config{ActiveWorkspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	rootCmd.SetArgs([]string{"place", "list"})
	err := rootCmd.Execute()
	if err != nil && err.Error() == "no active idea — run: yherda ideas use <id>" {
		t.Error("should fall back to listing rather than returning a hard error")
	}
}

func TestPlaceList_ContextIdeaUsed_NoContextError(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Config{ActiveWorkspace: "ws", APIServer: "https://ws.yherda.test:8000", ActiveIdea: "idea-1"})

	rootCmd.SetArgs([]string{"place", "list"})
	err := rootCmd.Execute()
	if err != nil && err.Error() == "no active idea — run: yherda ideas use <id>" {
		t.Error("active_idea in config should have satisfied the requirement")
	}
}

// --- setting list ---

func TestSettingList_NoPlaceContext_FallsBack(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Config{ActiveWorkspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	rootCmd.SetArgs([]string{"setting", "list"})
	err := rootCmd.Execute()
	if err != nil && err.Error() == "no active place — run: yherda place use <id>" {
		t.Error("should fall back to listing rather than returning a hard error")
	}
}

func TestSettingList_ContextPlaceUsed_NoContextError(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Config{ActiveWorkspace: "ws", APIServer: "https://ws.yherda.test:8000", ActivePlace: "place-1"})

	rootCmd.SetArgs([]string{"setting", "list"})
	err := rootCmd.Execute()
	if err != nil && err.Error() == "no active place — run: yherda place use <id>" {
		t.Error("active_place in config should have satisfied the requirement")
	}
}

// --- thing list ---

func TestThingList_NoIdeaContext_FallsBack(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Config{ActiveWorkspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	rootCmd.SetArgs([]string{"thing", "list"})
	err := rootCmd.Execute()
	if err != nil && err.Error() == "no active idea — run: yherda ideas use <id>" {
		t.Error("should fall back to listing rather than returning a hard error")
	}
}

func TestThingList_ContextIdeaUsed_NoContextError(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Config{ActiveWorkspace: "ws", APIServer: "https://ws.yherda.test:8000", ActiveIdea: "idea-1"})

	rootCmd.SetArgs([]string{"thing", "list"})
	err := rootCmd.Execute()
	if err != nil && err.Error() == "no active idea — run: yherda ideas use <id>" {
		t.Error("active_idea in config should have satisfied the requirement")
	}
}

// --- disposition list ---

func TestDispositionList_NoThingContext_FallsBack(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Config{ActiveWorkspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	rootCmd.SetArgs([]string{"disposition", "list"})
	err := rootCmd.Execute()
	if err != nil && err.Error() == "no active thing — run: yherda thing use <id>" {
		t.Error("should fall back to listing rather than returning a hard error")
	}
}

func TestDispositionList_ContextThingUsed_NoContextError(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Config{ActiveWorkspace: "ws", APIServer: "https://ws.yherda.test:8000", ActiveThing: "thing-1"})

	rootCmd.SetArgs([]string{"disposition", "list"})
	err := rootCmd.Execute()
	if err != nil && err.Error() == "no active thing — run: yherda thing use <id>" {
		t.Error("active_thing in config should have satisfied the requirement")
	}
}

// --- use cascade ---

func TestPersonUse_SetsPersonClearsDownstream(t *testing.T) {
	withTempHome(t)
	saveContext(t, &config.Config{
		ActiveWorkspace: "ws",
		ActiveIdea:      "idea-1",
		ActivePerson:    "old-person",
		ActiveArc:       "arc-1",
		ActivePlace:     "place-1",
		ActiveThing:     "thing-1",
	})

	rootCmd.SetArgs([]string{"person", "use", "new-person"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	loaded, _ := config.LoadConfig()
	if loaded.ActivePerson != "new-person" {
		t.Errorf("active_person: got %q, want %q", loaded.ActivePerson, "new-person")
	}
	if loaded.ActiveIdea != "idea-1" {
		t.Errorf("active_idea should be unchanged: got %q", loaded.ActiveIdea)
	}
	if loaded.ActiveArc != "" {
		t.Errorf("active_arc should be cleared, got %q", loaded.ActiveArc)
	}
	if loaded.ActivePlace != "" {
		t.Errorf("active_place should be cleared, got %q", loaded.ActivePlace)
	}
	if loaded.ActiveThing != "" {
		t.Errorf("active_thing should be cleared, got %q", loaded.ActiveThing)
	}
}

func TestPlaceUse_SetsPlaceClearsDownstream(t *testing.T) {
	withTempHome(t)
	saveContext(t, &config.Config{
		ActiveWorkspace: "ws",
		ActiveIdea:      "idea-1",
		ActivePerson:    "person-1",
		ActiveArc:       "arc-1",
		ActivePlace:     "old-place",
		ActiveThing:     "thing-1",
	})

	rootCmd.SetArgs([]string{"place", "use", "new-place"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	loaded, _ := config.LoadConfig()
	if loaded.ActivePlace != "new-place" {
		t.Errorf("active_place: got %q, want %q", loaded.ActivePlace, "new-place")
	}
	if loaded.ActiveIdea != "idea-1" {
		t.Errorf("active_idea should be unchanged: got %q", loaded.ActiveIdea)
	}
	if loaded.ActivePerson != "" {
		t.Errorf("active_person should be cleared, got %q", loaded.ActivePerson)
	}
	if loaded.ActiveArc != "" {
		t.Errorf("active_arc should be cleared, got %q", loaded.ActiveArc)
	}
	if loaded.ActiveThing != "" {
		t.Errorf("active_thing should be cleared, got %q", loaded.ActiveThing)
	}
}

func TestThingUse_SetsThingClearsDownstream(t *testing.T) {
	withTempHome(t)
	saveContext(t, &config.Config{
		ActiveWorkspace: "ws",
		ActiveIdea:      "idea-1",
		ActivePerson:    "person-1",
		ActiveArc:       "arc-1",
		ActivePlace:     "place-1",
		ActiveThing:     "old-thing",
	})

	rootCmd.SetArgs([]string{"thing", "use", "new-thing"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	loaded, _ := config.LoadConfig()
	if loaded.ActiveThing != "new-thing" {
		t.Errorf("active_thing: got %q, want %q", loaded.ActiveThing, "new-thing")
	}
	if loaded.ActiveIdea != "idea-1" {
		t.Errorf("active_idea should be unchanged: got %q", loaded.ActiveIdea)
	}
	if loaded.ActivePerson != "" {
		t.Errorf("active_person should be cleared, got %q", loaded.ActivePerson)
	}
	if loaded.ActiveArc != "" {
		t.Errorf("active_arc should be cleared, got %q", loaded.ActiveArc)
	}
	if loaded.ActivePlace != "" {
		t.Errorf("active_place should be cleared, got %q", loaded.ActivePlace)
	}
}

func TestIdeasUse_SetsIdeaClearsAll(t *testing.T) {
	withTempHome(t)
	saveContext(t, &config.Config{
		ActiveWorkspace: "ws",
		ActiveIdea:      "old-idea",
		ActivePerson:    "person-1",
		ActiveArc:       "arc-1",
		ActivePlace:     "place-1",
		ActiveThing:     "thing-1",
	})

	rootCmd.SetArgs([]string{"ideas", "use", "new-idea"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	loaded, _ := config.LoadConfig()
	if loaded.ActiveIdea != "new-idea" {
		t.Errorf("active_idea: got %q, want %q", loaded.ActiveIdea, "new-idea")
	}
	if loaded.ActivePerson != "" {
		t.Errorf("active_person should be cleared, got %q", loaded.ActivePerson)
	}
	if loaded.ActiveArc != "" {
		t.Errorf("active_arc should be cleared, got %q", loaded.ActiveArc)
	}
	if loaded.ActivePlace != "" {
		t.Errorf("active_place should be cleared, got %q", loaded.ActivePlace)
	}
	if loaded.ActiveThing != "" {
		t.Errorf("active_thing should be cleared, got %q", loaded.ActiveThing)
	}
}
