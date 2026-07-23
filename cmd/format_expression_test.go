package cmd

// format and expression commands are disabled — being replaced. See YOS-80.
// These tests drive both via rootCmd, which now returns "unknown command"
// since neither is registered; commented out rather than deleted so they can
// be restored once the replacement lands.

/*
import (
	"testing"

	"github.com/yherda-opensource/yherda-cmd/internal/config"
)

// format show requires an id argument
func TestFormatShow_MissingArg(t *testing.T) {
	rootCmd.SetArgs([]string{"format", "show"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error when id argument is missing")
	}
}

// expression show requires an id argument
func TestExpressionShow_MissingArg(t *testing.T) {
	rootCmd.SetArgs([]string{"expression", "show"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error when id argument is missing")
	}
}

// expression print requires an id argument
func TestExpressionPrint_MissingArg(t *testing.T) {
	rootCmd.SetArgs([]string{"expression", "print"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error when id argument is missing")
	}
}

// expression list falls back to idea list when no context is set
func TestExpressionList_NoIdeaContext_FallsBack(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	rootCmd.SetArgs([]string{"expression", "list"})
	err := rootCmd.Execute()
	if err != nil && err.Error() == "no active idea" {
		t.Error("should fall back to listing ideas rather than returning a hard error")
	}
}

// expression list uses --idea flag when no active idea is set
func TestExpressionList_IdeaFlagProvided_NoContextError(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	rootCmd.SetArgs([]string{"expression", "list", "--idea", "some-id"})
	err := rootCmd.Execute()
	if err != nil && err.Error() == "no active idea" {
		t.Error("--idea flag should have satisfied the idea requirement")
	}
}

// expression list uses active idea from context
func TestExpressionList_ContextIdeaUsed_NoContextError(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000", Idea: "idea-1"})

	rootCmd.SetArgs([]string{"expression", "list"})
	err := rootCmd.Execute()
	if err != nil && err.Error() == "no active idea" {
		t.Error("active_idea in config should have satisfied the requirement")
	}
}
*/
