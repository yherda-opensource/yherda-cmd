package cmd

import (
	"testing"

	"github.com/yherda-opensource/yherda-cmd/internal/config"
)

func withTempHome(t *testing.T) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	t.Chdir(t.TempDir())
}

func saveContext(t *testing.T, ctx *config.Context) {
	t.Helper()
	if err := config.SaveContext(ctx); err != nil {
		t.Fatalf("saveContext: %v", err)
	}
}

// saveContextWithCreds saves context and stub credentials so mustClient() doesn't os.Exit.
func saveContextWithCreds(t *testing.T, ctx *config.Context) {
	t.Helper()
	saveContext(t, ctx)
	if err := config.SaveCredentials(&config.Credentials{
		AccessToken: "test-token", TokenType: "Bearer",
	}); err != nil {
		t.Fatalf("SaveCredentials: %v", err)
	}
}

// --- person list ---

func TestPersonList_NoIdeaContext_FallsBackToIdeaList(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	rootCmd.SetArgs([]string{"person", "list"})
	err := rootCmd.Execute()
	if err != nil && err.Error() == "no active idea — run: yherda ideas use <id>" {
		t.Error("should fall back to listing ideas rather than returning a hard error")
	}
}

func TestPersonList_FlagProvided_NoContextError(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	rootCmd.SetArgs([]string{"person", "list", "--idea", "some-id"})
	err := rootCmd.Execute()
	if err != nil && err.Error() == "no active idea — run: yherda ideas use <id>" {
		t.Error("flag should have satisfied the idea requirement")
	}
}

func TestPersonList_ContextIdeaUsed_NoContextError(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000", Idea: "idea-1"})

	rootCmd.SetArgs([]string{"person", "list"})
	err := rootCmd.Execute()
	if err != nil && err.Error() == "no active idea — run: yherda ideas use <id>" {
		t.Error("active idea in context should have satisfied the requirement")
	}
}

// --- identity list ---

func TestIdentityList_NoPersonContext_FallsBack(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	rootCmd.SetArgs([]string{"identity", "list"})
	err := rootCmd.Execute()
	if err != nil && err.Error() == "no active person — run: yherda person use <id>" {
		t.Error("should fall back to listing rather than returning a hard error")
	}
}

func TestIdentityList_ContextPersonUsed_NoContextError(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000", Person: "person-1"})

	rootCmd.SetArgs([]string{"identity", "list"})
	err := rootCmd.Execute()
	if err != nil && err.Error() == "no active person — run: yherda person use <id>" {
		t.Error("active person in context should have satisfied the requirement")
	}
}

// --- place list ---

func TestPlaceList_NoIdeaContext_FallsBack(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	rootCmd.SetArgs([]string{"place", "list"})
	err := rootCmd.Execute()
	if err != nil && err.Error() == "no active idea — run: yherda ideas use <id>" {
		t.Error("should fall back to listing rather than returning a hard error")
	}
}

func TestPlaceList_ContextIdeaUsed_NoContextError(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000", Idea: "idea-1"})

	rootCmd.SetArgs([]string{"place", "list"})
	err := rootCmd.Execute()
	if err != nil && err.Error() == "no active idea — run: yherda ideas use <id>" {
		t.Error("active idea in context should have satisfied the requirement")
	}
}

// --- setting list ---

func TestSettingList_NoPlaceContext_FallsBack(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	rootCmd.SetArgs([]string{"setting", "list"})
	err := rootCmd.Execute()
	if err != nil && err.Error() == "no active place — run: yherda place use <id>" {
		t.Error("should fall back to listing rather than returning a hard error")
	}
}

func TestSettingList_ContextPlaceUsed_NoContextError(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000", Place: "place-1"})

	rootCmd.SetArgs([]string{"setting", "list"})
	err := rootCmd.Execute()
	if err != nil && err.Error() == "no active place — run: yherda place use <id>" {
		t.Error("active place in context should have satisfied the requirement")
	}
}

// --- thing list ---

func TestThingList_NoIdeaContext_FallsBack(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	rootCmd.SetArgs([]string{"thing", "list"})
	err := rootCmd.Execute()
	if err != nil && err.Error() == "no active idea — run: yherda ideas use <id>" {
		t.Error("should fall back to listing rather than returning a hard error")
	}
}

func TestThingList_ContextIdeaUsed_NoContextError(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000", Idea: "idea-1"})

	rootCmd.SetArgs([]string{"thing", "list"})
	err := rootCmd.Execute()
	if err != nil && err.Error() == "no active idea — run: yherda ideas use <id>" {
		t.Error("active idea in context should have satisfied the requirement")
	}
}

// --- disposition list ---

func TestDispositionList_NoThingContext_FallsBack(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	rootCmd.SetArgs([]string{"disposition", "list"})
	err := rootCmd.Execute()
	if err != nil && err.Error() == "no active thing — run: yherda thing use <id>" {
		t.Error("should fall back to listing rather than returning a hard error")
	}
}

func TestDispositionList_ContextThingUsed_NoContextError(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000", Thing: "thing-1"})

	rootCmd.SetArgs([]string{"disposition", "list"})
	err := rootCmd.Execute()
	if err != nil && err.Error() == "no active thing — run: yherda thing use <id>" {
		t.Error("active thing in context should have satisfied the requirement")
	}
}

// --- use cascade ---

func TestPersonUse_SetsPersonClearsDownstream(t *testing.T) {
	withTempHome(t)
	saveContext(t, &config.Context{
		Workspace: "ws",
		Idea:      "idea-1",
		Person:    "old-person",
		Place:     "place-1",
		Thing:     "thing-1",
	})

	rootCmd.SetArgs([]string{"person", "use", "new-person"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	loaded, _ := config.LoadContext()
	if loaded.Person != "new-person" {
		t.Errorf("active_person: got %q, want %q", loaded.Person, "new-person")
	}
	if loaded.Idea != "idea-1" {
		t.Errorf("active_idea should be unchanged: got %q", loaded.Idea)
	}
	if loaded.Place != "" {
		t.Errorf("active_place should be cleared, got %q", loaded.Place)
	}
	if loaded.Thing != "" {
		t.Errorf("active_thing should be cleared, got %q", loaded.Thing)
	}
}

func TestPlaceUse_SetsPlaceClearsDownstream(t *testing.T) {
	withTempHome(t)
	saveContext(t, &config.Context{
		Workspace: "ws",
		Idea:      "idea-1",
		Person:    "person-1",
		Place:     "old-place",
		Thing:     "thing-1",
	})

	rootCmd.SetArgs([]string{"place", "use", "new-place"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	loaded, _ := config.LoadContext()
	if loaded.Place != "new-place" {
		t.Errorf("active_place: got %q, want %q", loaded.Place, "new-place")
	}
	if loaded.Idea != "idea-1" {
		t.Errorf("active_idea should be unchanged: got %q", loaded.Idea)
	}
	if loaded.Person != "" {
		t.Errorf("active_person should be cleared, got %q", loaded.Person)
	}
	if loaded.Thing != "" {
		t.Errorf("active_thing should be cleared, got %q", loaded.Thing)
	}
}

func TestThingUse_SetsThingClearsDownstream(t *testing.T) {
	withTempHome(t)
	saveContext(t, &config.Context{
		Workspace: "ws",
		Idea:      "idea-1",
		Person:    "person-1",
		Place:     "place-1",
		Thing:     "old-thing",
	})

	rootCmd.SetArgs([]string{"thing", "use", "new-thing"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	loaded, _ := config.LoadContext()
	if loaded.Thing != "new-thing" {
		t.Errorf("active_thing: got %q, want %q", loaded.Thing, "new-thing")
	}
	if loaded.Idea != "idea-1" {
		t.Errorf("active_idea should be unchanged: got %q", loaded.Idea)
	}
	if loaded.Person != "" {
		t.Errorf("active_person should be cleared, got %q", loaded.Person)
	}
	if loaded.Place != "" {
		t.Errorf("active_place should be cleared, got %q", loaded.Place)
	}
}

func TestIdeasUse_SetsIdeaClearsAll(t *testing.T) {
	withTempHome(t)
	saveContext(t, &config.Context{
		Workspace: "ws",
		Idea:      "old-idea",
		Person:    "person-1",
		Place:     "place-1",
		Thing:     "thing-1",
	})

	rootCmd.SetArgs([]string{"ideas", "use", "new-idea"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	loaded, _ := config.LoadContext()
	if loaded.Idea != "new-idea" {
		t.Errorf("active_idea: got %q, want %q", loaded.Idea, "new-idea")
	}
	if loaded.Person != "" {
		t.Errorf("active_person should be cleared, got %q", loaded.Person)
	}
	if loaded.Place != "" {
		t.Errorf("active_place should be cleared, got %q", loaded.Place)
	}
	if loaded.Thing != "" {
		t.Errorf("active_thing should be cleared, got %q", loaded.Thing)
	}
}
