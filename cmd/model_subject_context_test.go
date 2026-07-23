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
// text this file's tests are about (the IDEA ID label, the two-line
// Subject/context footer) — YOS-81 is a bug about output text being
// misleading, so the fix has to be verified by reading that text, not just
// by checking that no error occurred.
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

// --- Subject-context footer (YOS-81 fix 2, revised to a two-line footer) ---
// printContextWithSubject prints the "record used" line (Subject id/name/type)
// then the usual context row, both AFTER a command's own output, replacing
// the earlier pre-output printSubjectContext confirmation line.

func TestModelDispositionsList_DoesNotError(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	rootCmd.SetArgs([]string{"model", "dispositions", "list", "42"})
	err := rootCmd.Execute()
	// No live server in tests, so this always errors on the primary API call —
	// the point is it doesn't panic when the trailing footer lookup also fails.
	if err == nil {
		t.Fatal("expected an error against the fake test server")
	}
}

func TestModelStatesList_DoesNotError(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	rootCmd.SetArgs([]string{"model", "states", "list", "42"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected an error against the fake test server")
	}
}

func TestModelPerspectiveGet_JSONMode_SkipsFooter(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	out := captureStdout(t, func() {
		rootCmd.SetArgs([]string{"model", "perspective", "get", "42", "--json"})
		rootCmd.Execute()
	})

	if bytes.Contains([]byte(out), []byte("Subject:")) {
		t.Errorf("--json mode should not print the human-readable Subject record line, got: %q", out)
	}
}

func TestPrintContextWithSubject_NoContext_SkipsFooter(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	out := captureStdout(t, func() {
		rootCmd.SetArgs([]string{"model", "perspective", "get", "42", "--no-context"})
		rootCmd.Execute()
	})

	if bytes.Contains([]byte(out), []byte("Subject:")) {
		t.Errorf("--no-context should suppress the Subject record footer line, got: %q", out)
	}
}

// --- context footer shows the active Subject's full row ---

func TestSubjectContextLabel_Empty_ReturnsEmpty(t *testing.T) {
	if got := subjectContextLabel(""); got != "" {
		t.Errorf("subjectContextLabel(\"\") = %q, want empty", got)
	}
}

func TestSubjectContextLabel_NoCreds_FallsBackToBareID(t *testing.T) {
	withTempHome(t)
	saveContext(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	got := subjectContextLabel("42")
	if got != "42" {
		t.Errorf("subjectContextLabel with no credentials should fall back to the bare id, got %q", got)
	}
}

func TestSubjectContextLabel_FetchFails_FallsBackToBareID(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	got := subjectContextLabel("42")
	if got != "42" {
		t.Errorf("subjectContextLabel should fall back to the bare id when the fetch fails, got %q", got)
	}
}

func TestPrintContext_SubjectSet_DoesNotPanic(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000", Subject: "42"})

	captureStdout(t, func() {
		printContext()
	})
}

func TestPrintContextWithSubject_DoesNotPanic(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	captureStdout(t, func() {
		printContextWithSubject(mustClient(), "42")
	})
}
