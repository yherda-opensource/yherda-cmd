package cmd

import (
	"testing"

	"github.com/yherda-opensource/yherda-cmd/internal/config"
)

// --- belief create ---

func TestModelBeliefCreate_MissingStatement_Error(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000", Idea: "idea-1"})

	rootCmd.SetArgs([]string{"model", "belief", "create"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error when --statement is missing")
	}
}

func TestModelBeliefCreate_NoIdeaNoContext_Error(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	rootCmd.SetArgs([]string{"model", "belief", "create", "--statement", "The king is not to be trusted"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error when no --idea and no active context")
	}
}

func TestModelBeliefCreate_ContextIdea_ReachesAPI(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000", Idea: "idea-1"})

	rootCmd.SetArgs([]string{"model", "belief", "create", "--statement", "She loved him once"})
	err := rootCmd.Execute()
	if err != nil && (err.Error() == "--idea is required (or set active idea with 'yherda ideas use <id>')" ||
		err.Error() == "--statement is required") {
		t.Errorf("context should have satisfied requirements, got: %v", err)
	}
}

func TestModelBeliefCreate_NoSubjectFlag_DoesNotErrorOnMissingSubject(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000", Idea: "idea-1"})

	rootCmd.SetArgs([]string{"model", "belief", "create", "--statement", "A belief with no subject"})
	err := rootCmd.Execute()
	if err != nil && err.Error() == "--statement is required" {
		t.Errorf("--subject should be optional: %v", err)
	}
}

// --- belief contexts add/update/remove ---

func TestModelBeliefContextsAdd_MissingContext_Error(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	rootCmd.SetArgs([]string{"model", "belief", "contexts", "add", "--belief", "12"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error when --context is missing")
	}
}

func TestModelBeliefContextsAdd_MissingBelief_Error(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	rootCmd.SetArgs([]string{"model", "belief", "contexts", "add", "--context", "3"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error when --belief is missing")
	}
}

func TestModelBeliefContextsAdd_NoStatusMode_ReachesAPI(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	rootCmd.SetArgs([]string{"model", "belief", "contexts", "add", "--context", "3", "--belief", "12"})
	err := rootCmd.Execute()
	if err != nil && (err.Error() == "--context is required" || err.Error() == "--belief is required") {
		t.Errorf("both flags were provided, should not error: %v", err)
	}
}

func TestModelBeliefContextsUpdate_MissingBelief_Error(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	rootCmd.SetArgs([]string{"model", "belief", "contexts", "update", "--context", "3", "--status", "Strained"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error when --belief is missing")
	}
}

func TestModelBeliefContextsRemove_MissingBelief_Error(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	rootCmd.SetArgs([]string{"model", "belief", "contexts", "remove", "--context", "3"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error when --belief is missing")
	}
}
