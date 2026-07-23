package cmd

import (
	"testing"

	"github.com/yherda-opensource/yherda-cmd/internal/config"
)

// --- perspective get ---

func TestModelPerspectiveGet_MissingID_Error(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	rootCmd.SetArgs([]string{"model", "perspective", "get"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error when subject id arg is missing")
	}
}

func TestModelPerspectiveGet_ReachesAPI(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	rootCmd.SetArgs([]string{"model", "perspective", "get", "42"})
	err := rootCmd.Execute()
	if err != nil && err.Error() == "not logged in — run 'yherda login' first" {
		t.Errorf("credentials should have satisfied login requirement: %v", err)
	}
}

// --- perspective contexts add/update/remove ---

func TestModelPerspectiveContextsAdd_MissingContext_Error(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	rootCmd.SetArgs([]string{"model", "perspective", "contexts", "add", "9"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error when --context is missing")
	}
}

func TestModelPerspectiveContextsAdd_NoPriority_ReachesAPI(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	rootCmd.SetArgs([]string{"model", "perspective", "contexts", "add", "9", "--context", "3"})
	err := rootCmd.Execute()
	if err != nil && err.Error() == "--context is required" {
		t.Errorf("--context was provided, should not error: %v", err)
	}
}

func TestModelPerspectiveContextsUpdate_MissingPriority_Error(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	rootCmd.SetArgs([]string{"model", "perspective", "contexts", "update", "9", "--context", "3"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error when --priority is missing on update")
	}
}

func TestModelPerspectiveContextsUpdate_WithPriority_ReachesAPI(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	rootCmd.SetArgs([]string{"model", "perspective", "contexts", "update", "9", "--context", "3", "--priority", "2"})
	err := rootCmd.Execute()
	if err != nil && err.Error() == "--priority is required" {
		t.Errorf("--priority was provided, should not error: %v", err)
	}
}

func TestModelPerspectiveContextsRemove_MissingContext_Error(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	rootCmd.SetArgs([]string{"model", "perspective", "contexts", "remove", "9"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error when --context is missing")
	}
}
