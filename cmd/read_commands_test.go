package cmd

import (
	"strings"
	"testing"
)

// API path routing and baseURL logic is tested in internal/api/client_test.go.
// These tests cover argument validation at the command layer.

func TestIdeasShow_MissingArg(t *testing.T) {
	rootCmd.SetArgs([]string{"ideas", "show"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error when id argument is missing")
	}
}

func TestPersonUse_MissingArg(t *testing.T) {
	rootCmd.SetArgs([]string{"person", "use"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error when id argument is missing")
	}
}

func TestPlaceUse_MissingArg(t *testing.T) {
	rootCmd.SetArgs([]string{"place", "use"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error when id argument is missing")
	}
}

func TestThingUse_MissingArg(t *testing.T) {
	rootCmd.SetArgs([]string{"thing", "use"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error when id argument is missing")
	}
}

func TestWorkspaceList_RejectsArgs(t *testing.T) {
	rootCmd.SetArgs([]string{"workspacelist", "extra-arg"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error when unexpected argument is passed")
	}
}

// YOS-88: the expression group follows the CLI's conventions.
func TestExpressionUse_MissingArg(t *testing.T) {
	rootCmd.SetArgs([]string{"expression", "use"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error when id argument is missing")
	}
}

func TestExpressionList_RejectsArgs(t *testing.T) {
	rootCmd.SetArgs([]string{"expression", "list", "extra"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error when an unexpected argument is passed")
	}
}

func TestActiveExpression_FlagWins(t *testing.T) {
	got, err := activeExpression("42")
	if err != nil || got != "42" {
		t.Fatalf("got %q, %v", got, err)
	}
}

func TestActiveExpression_NoneSetSaysWhatToDo(t *testing.T) {
	t.Chdir(t.TempDir())
	_, err := activeExpression("")
	if err == nil || !strings.Contains(err.Error(), "expression use") {
		t.Fatalf("got %v", err)
	}
}
