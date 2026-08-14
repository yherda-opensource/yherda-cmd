package cmd

import (
	"testing"

	"github.com/yherda-opensource/yherda-cmd/internal/config"
)

// --- context list ---

func TestContextList_NoIdeaNoContext_Error(t *testing.T) {
	withTempHome(t)
	saveContext(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	rootCmd.SetArgs([]string{"context", "list"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error when no --idea and no active context")
	}
}

func TestContextList_ContextIdeaUsed_NoContextError(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000", Idea: "idea-1"})

	rootCmd.SetArgs([]string{"context", "list"})
	err := rootCmd.Execute()
	if err != nil && err.Error() == "--idea is required (or set active idea with 'yherda ideas use <id>')" {
		t.Error("active idea in context should have satisfied the requirement")
	}
}

// --- context use ---

func TestContextUse_SetsContext(t *testing.T) {
	withTempHome(t)
	saveContext(t, &config.Context{Workspace: "ws", Idea: "idea-1"})

	rootCmd.SetArgs([]string{"context", "use", "new-context"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	loaded, _ := config.LoadContext()
	if loaded.Context != "new-context" {
		t.Errorf("active_context: got %q, want %q", loaded.Context, "new-context")
	}
	if loaded.Idea != "idea-1" {
		t.Errorf("active_idea should be unchanged: got %q", loaded.Idea)
	}
}

// --- ideas use clears context ---

func TestIdeasUse_ClearsContext(t *testing.T) {
	withTempHome(t)
	saveContext(t, &config.Context{
		Workspace: "ws",
		Idea:      "old-idea",
		Context:   "context-1",
	})

	rootCmd.SetArgs([]string{"ideas", "use", "new-idea"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	loaded, _ := config.LoadContext()
	if loaded.Context != "" {
		t.Errorf("active_context should be cleared, got %q", loaded.Context)
	}
}

// --- context belief list ---

func TestContextBeliefList_NoContextNoActive_Error(t *testing.T) {
	withTempHome(t)
	saveContext(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000"})
	contextBeliefContextID = ""

	rootCmd.SetArgs([]string{"context", "belief", "list"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error when no --context and no active context")
	}
}
