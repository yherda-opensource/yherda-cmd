package cmd

import (
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

func TestArcUse_MissingArg(t *testing.T) {
	rootCmd.SetArgs([]string{"arc", "use"})
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
