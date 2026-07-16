package structural

import (
	"os"
	"path/filepath"
	"testing"
)

// stubDriver is a no-op driver used to test controller behaviour without
// exercising any serialisation logic.
type stubDriver struct {
	called bool
}

func (s *stubDriver) Export(graph IdeaGraph, output string) error {
	s.called = true
	return nil
}

func (s *stubDriver) DefaultExt() string { return "" }

func TestExport_OutputDirExistsWithFiles(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "existing.md"), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	f := &mockFetcher{}
	driver := &stubDriver{}
	err := Export(f, driver, Manifest{}, "42", dir)
	if err == nil {
		t.Error("expected error for non-empty output dir, got nil")
	}
	if driver.called {
		t.Error("driver should not be called when output dir check fails")
	}
}

func TestExport_EmptyOutputSkipsCheck(t *testing.T) {
	f := &mockFetcher{}
	driver := &stubDriver{}
	// Empty output path — controller skips the dir check and lets the driver decide.
	if err := Export(f, driver, Manifest{}, "42", ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !driver.called {
		t.Error("expected driver to be called")
	}
}

func TestExport_IdentitiesFetchesPersons(t *testing.T) {
	f := &mockFetcher{personIDs: []string{"p1"}}
	driver := &stubDriver{}
	if err := Export(f, driver, Manifest{Identities: true}, "42", ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasCall(f.calls, "/storyline/42/persons/") {
		t.Error("expected persons to be fetched")
	}
}
