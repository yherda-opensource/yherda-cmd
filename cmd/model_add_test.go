package cmd

import (
	"testing"

	"github.com/yherda-opensource/yherda-cmd/internal/config"
)

// --- model add perspective ---

func TestModelAddPerspective_NoArgNoContext_Error(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	rootCmd.SetArgs([]string{"model", "add", "perspective"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error when no subject id arg and no active subject in context")
	}
}

func TestModelAddPerspective_ExplicitSubjectID_ReachesAPI(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	rootCmd.SetArgs([]string{"model", "add", "perspective", "42"})
	err := rootCmd.Execute()
	if err != nil && err.Error() == "no active subject — pass a subject id or run 'yherda model use <subject-id>'" {
		t.Errorf("explicit subject id arg should have satisfied the requirement: %v", err)
	}
}

func TestModelAddPerspective_ContextSubject_ReachesAPI(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000", Subject: "subject-1"})

	rootCmd.SetArgs([]string{"model", "add", "perspective"})
	err := rootCmd.Execute()
	if err != nil && err.Error() == "no active subject — pass a subject id or run 'yherda model use <subject-id>'" {
		t.Errorf("active subject in context should have satisfied the requirement: %v", err)
	}
}

// --- model add goal ---

func TestModelAddGoal_NoArgNoContext_Error(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	rootCmd.SetArgs([]string{"model", "add", "goal", "--want", "To find her father", "--yes"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error when no subject id arg and no active subject in context")
	}
}

func TestModelAddGoal_ExplicitSubjectID_SkipConfirm_ReachesAPIDirectly(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	rootCmd.SetArgs([]string{"model", "add", "goal", "42", "--want", "To find her father", "--yes"})
	err := rootCmd.Execute()
	if err != nil && err.Error() == "aborted" {
		t.Error("--yes should have skipped the confirmation prompt entirely")
	}
	if err != nil && err.Error() == "no active subject — pass a subject id or run 'yherda model use <subject-id>'" {
		t.Errorf("explicit subject id arg should have satisfied the requirement: %v", err)
	}
}

func TestModelAddGoal_ContextSubject_SkipConfirm_ReachesAPIDirectly(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000", Subject: "subject-1"})

	rootCmd.SetArgs([]string{"model", "add", "goal", "--want", "To find her father", "--yes"})
	err := rootCmd.Execute()
	if err != nil && err.Error() == "no active subject — pass a subject id or run 'yherda model use <subject-id>'" {
		t.Errorf("active subject in context should have satisfied the requirement: %v", err)
	}
}

func TestModelAddGoal_EmptyWant_WarnsButProceeds(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000", Subject: "subject-1"})

	rootCmd.SetArgs([]string{"model", "add", "goal", "--yes"})
	err := rootCmd.Execute()
	if err != nil && err.Error() == "--want is required" {
		t.Errorf("empty --want should warn, not hard-block: %v", err)
	}
}
