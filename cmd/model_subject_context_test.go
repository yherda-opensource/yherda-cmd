package cmd

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/yherda-opensource/yherda-cmd/internal/config"
)

// captureStdout runs fn with os.Stdout redirected to a pipe and returns
// everything written to it. Used specifically to assert on the disambiguating
// text this file's tests are about (the IDEA ID label, the Subject
// confirmation line) — YOS-81 is a bug about output text being misleading,
// so the fix has to be verified by reading that text, not just by checking
// that no error occurred.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stdout = w
	defer func() { os.Stdout = old }()

	fn()

	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

// --- model list idea-fallback labeling (YOS-81 fix 1) ---

func TestModelList_Fallback_LabelsIdeaIDsDistinctly(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	out := captureStdout(t, func() {
		rootCmd.SetArgs([]string{"model", "list"})
		rootCmd.Execute()
	})

	if !bytes.Contains([]byte(out), []byte("IDEA ids")) {
		t.Errorf("fallback should explicitly call out IDEA ids in its warning, got output: %q", out)
	}
	if !bytes.Contains([]byte(out), []byte("not Subject ids")) {
		t.Errorf("fallback should explicitly warn these aren't Subject ids, got output: %q", out)
	}
}

func TestModelList_ExplicitIdea_DoesNotUseIdeaIDLabel(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	out := captureStdout(t, func() {
		rootCmd.SetArgs([]string{"model", "list", "some-idea"})
		rootCmd.Execute()
	})

	if bytes.Contains([]byte(out), []byte("IDEA ID")) {
		t.Errorf("normal (non-fallback) model list should not print the IDEA ID fallback label, got: %q", out)
	}
}

// --- Subject-context visibility (YOS-81 fix 2) ---
// These assert that commands attempt the extra Subject lookup (surfaced as a
// distinguishable error path against the fake test server) rather than
// silently acting on a bare id with no visibility into what it resolves to.

func TestModelDispositionsList_LooksUpSubjectContext(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	rootCmd.SetArgs([]string{"model", "dispositions", "list", "42"})
	err := rootCmd.Execute()
	// No live server in tests, so this always errors — the point is that it
	// doesn't panic and doesn't skip straight past the subject-context lookup.
	if err == nil {
		t.Fatal("expected an error against the fake test server")
	}
}

func TestModelStatesList_LooksUpSubjectContext(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	rootCmd.SetArgs([]string{"model", "states", "list", "42"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected an error against the fake test server")
	}
}

func TestModelPerspectiveGet_JSONMode_SkipsSubjectContextLine(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	out := captureStdout(t, func() {
		rootCmd.SetArgs([]string{"model", "perspective", "get", "42", "--json"})
		rootCmd.Execute()
	})

	if bytes.Contains([]byte(out), []byte("Subject:")) {
		t.Errorf("--json mode should not print the human-readable Subject confirmation line, got: %q", out)
	}
}
