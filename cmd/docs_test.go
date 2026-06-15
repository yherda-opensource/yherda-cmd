package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yherda-opensource/yherda-cmd/internal/config"
)

// TestDocsShow_MissingArg verifies that docs show requires a positional argument.
func TestDocsShow_MissingArg(t *testing.T) {
	rootCmd.SetArgs([]string{"doc", "show"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error when doc-id argument is missing")
	}
}

// TestDocsUpdate_MissingArg verifies that docs update requires a positional argument.
func TestDocsUpdate_MissingArg(t *testing.T) {
	rootCmd.SetArgs([]string{"doc", "update"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error when doc-id argument is missing")
	}
}

// TestDocsList_MissingIdeaFlag verifies that docs list fails without --idea.
func TestDocsList_MissingIdeaFlag(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Config{ActiveWorkspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	rootCmd.SetArgs([]string{"doc", "list"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error when --idea flag is missing")
	}
	if err.Error() != "--idea is required" {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestDocsCreate_MissingIdeaFlag verifies that docs create fails without --idea.
func TestDocsCreate_MissingIdeaFlag(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Config{ActiveWorkspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	rootCmd.SetArgs([]string{"doc", "create", "--title", "My Doc"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error when --idea flag is missing")
	}
	if err.Error() != "--idea is required" {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestDocsCreate_MissingTitleFlag verifies that docs create fails without --title.
// Note: --idea must also be provided so we reach the --title check.
func TestDocsCreate_MissingTitleFlag(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Config{ActiveWorkspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	// Reset flag state from previous tests in this package.
	docsCreateCmd.Flags().Set("idea", "idea-1")
	docsCreateCmd.Flags().Set("title", "")
	t.Cleanup(func() {
		docsCreateCmd.Flags().Set("idea", "")
		docsCreateCmd.Flags().Set("title", "")
	})

	rootCmd.SetArgs([]string{"doc", "create", "--idea", "idea-1"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error when --title flag is missing")
	}
	if err.Error() != "--title is required" {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestDocsCreate_MissingFile verifies that docs create fails when --file path doesn't exist.
func TestDocsCreate_MissingFile(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Config{ActiveWorkspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	rootCmd.SetArgs([]string{"doc", "create", "--idea", "idea-1", "--title", "My Doc", "--file", "/nonexistent/path.md"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error when --file path does not exist")
	}
}

// TestDocsUpdate_MissingFile verifies that docs update fails when --file path doesn't exist.
func TestDocsUpdate_MissingFile(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Config{ActiveWorkspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	rootCmd.SetArgs([]string{"doc", "update", "doc-1", "--file", "/nonexistent/path.md"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error when --file path does not exist")
	}
}

// TestReadContent_File verifies that readContent reads from a file when --file is provided.
func TestReadContent_File(t *testing.T) {
	content := "# Hello\n\nThis is a test document.\n"
	tmp := filepath.Join(t.TempDir(), "doc.md")
	if err := os.WriteFile(tmp, []byte(content), 0600); err != nil {
		t.Fatalf("setup: %v", err)
	}

	docsCreateCmd.Flags().Set("file", tmp)
	t.Cleanup(func() { docsCreateCmd.Flags().Set("file", "") })

	got, err := readContent(docsCreateCmd)
	if err != nil {
		t.Fatalf("readContent: %v", err)
	}
	if got != content {
		t.Errorf("got %q, want %q", got, content)
	}
}
