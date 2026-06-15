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

func TestArcsShow_MissingArg(t *testing.T) {
	rootCmd.SetArgs([]string{"arcs", "show"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error when id argument is missing")
	}
}

func TestBeatsShow_MissingArg(t *testing.T) {
	rootCmd.SetArgs([]string{"beats", "show"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error when id argument is missing")
	}
}

func TestIdentitiesShow_MissingArg(t *testing.T) {
	rootCmd.SetArgs([]string{"identities", "show"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error when id argument is missing")
	}
}
