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
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000", SubjectStack: []string{"42"}})

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

// --- resolveSubjectID fallback (YOS-82) ---
// model use should be the default fallback for every bare-Subject-id command.
// resolveSubjectID: explicit arg wins, else falls back to ctx.Subject, else a
// clear error — same shape as resolveGoalID/resolveStateID.

func TestResolveSubjectID_ExplicitArg_WinsOverContext(t *testing.T) {
	withTempHome(t)
	saveContext(t, &config.Context{Workspace: "ws", SubjectStack: []string{"99"}})

	got, err := resolveSubjectID([]string{"42"})
	if err != nil {
		t.Fatalf("resolveSubjectID: %v", err)
	}
	if got != "42" {
		t.Errorf("resolveSubjectID with explicit arg = %q, want %q (should not fall back to ctx.Subject)", got, "42")
	}
}

func TestResolveSubjectID_ExplicitArg_ResetsExistingStack(t *testing.T) {
	withTempHome(t)
	saveContext(t, &config.Context{Workspace: "ws", SubjectStack: []string{"person-1", "disposition-7"}})

	if _, err := resolveSubjectID([]string{"99"}); err != nil {
		t.Fatalf("resolveSubjectID: %v", err)
	}

	loaded, _ := config.LoadContext()
	if loaded.Subject() != "99" {
		t.Errorf("stack top after explicit override = %q, want %q", loaded.Subject(), "99")
	}
	if len(loaded.SubjectStack) != 1 {
		t.Errorf("stack should be reset to a single item, got %v", loaded.SubjectStack)
	}
}

func TestResolveSubjectID_NoArg_FallsBackToContext(t *testing.T) {
	withTempHome(t)
	saveContext(t, &config.Context{Workspace: "ws", SubjectStack: []string{"99"}})

	got, err := resolveSubjectID(nil)
	if err != nil {
		t.Fatalf("resolveSubjectID: %v", err)
	}
	if got != "99" {
		t.Errorf("resolveSubjectID with no arg = %q, want ctx.Subject %q", got, "99")
	}
}

func TestResolveSubjectID_NoArgNoContext_ReturnsClearError(t *testing.T) {
	withTempHome(t)
	saveContext(t, &config.Context{Workspace: "ws"})

	_, err := resolveSubjectID(nil)
	if err == nil {
		t.Fatal("expected an error when no arg and no active subject")
	}
	if !bytes.Contains([]byte(err.Error()), []byte("model use")) {
		t.Errorf("error should point the user at 'model use <subject-id>', got: %q", err.Error())
	}
}

// Command-level smoke tests: each formerly-required-id command still accepts
// no positional arg and reaches the fallback (rather than failing cobra's
// Args validation), for every command YOS-82 converted from ExactArgs(1) to
// MaximumNArgs(1). No live server in tests, so each errors on the API call —
// the point is it gets past arg parsing and into resolveSubjectID/the request.

func TestModelShow_NoArg_FallsBackToContext(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000", SubjectStack: []string{"42"}})

	rootCmd.SetArgs([]string{"model", "show"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected an error against the fake test server")
	}
	if bytes.Contains([]byte(err.Error()), []byte("no active subject")) {
		t.Errorf("should have used ctx.Subject fallback, not errored on missing subject: %v", err)
	}
}

func TestModelShow_NoArgNoContext_ReturnsFallbackError(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	rootCmd.SetArgs([]string{"model", "show"})
	err := rootCmd.Execute()
	if err == nil || !bytes.Contains([]byte(err.Error()), []byte("no active subject")) {
		t.Errorf("expected 'no active subject' error, got: %v", err)
	}
}

func TestModelDispositionsList_NoArg_FallsBackToContext(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000", SubjectStack: []string{"42"}})

	rootCmd.SetArgs([]string{"model", "dispositions", "list"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected an error against the fake test server")
	}
	if bytes.Contains([]byte(err.Error()), []byte("no active subject")) {
		t.Errorf("should have used ctx.Subject fallback, not errored on missing subject: %v", err)
	}
}

func TestModelStatesList_NoArg_FallsBackToContext(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000", SubjectStack: []string{"42"}})

	rootCmd.SetArgs([]string{"model", "states", "list"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected an error against the fake test server")
	}
	if bytes.Contains([]byte(err.Error()), []byte("no active subject")) {
		t.Errorf("should have used ctx.Subject fallback, not errored on missing subject: %v", err)
	}
}

func TestModelPerspectiveGet_NoArg_FallsBackToContext(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000", SubjectStack: []string{"42"}})

	rootCmd.SetArgs([]string{"model", "perspective", "get"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected an error against the fake test server")
	}
	if bytes.Contains([]byte(err.Error()), []byte("no active subject")) {
		t.Errorf("should have used ctx.Subject fallback, not errored on missing subject: %v", err)
	}
}

func TestModelGoalsList_NoArg_FallsBackToContext(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000", SubjectStack: []string{"42"}})

	rootCmd.SetArgs([]string{"model", "goals", "list"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected an error against the fake test server")
	}
	if bytes.Contains([]byte(err.Error()), []byte("no active subject")) {
		t.Errorf("should have used ctx.Subject fallback, not errored on missing subject: %v", err)
	}
}
