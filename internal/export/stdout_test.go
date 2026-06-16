package export

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func captureStdout(fn func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	fn()
	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestStdoutExport_ContentOnly(t *testing.T) {
	roots := []SegmentNode{
		{TypeName: "Act", Number: "1", Content: "", Children: []SegmentNode{
			{TypeName: "Scene", Number: "1", Content: "First scene content"},
			{TypeName: "Scene", Number: "2", Content: "Second scene content"},
		}},
	}
	e := &StdoutExporter{}
	out := captureStdout(func() {
		if err := e.Export("", roots, ""); err != nil {
			t.Fatalf("Export error: %v", err)
		}
	})
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d: %q", len(lines), out)
	}
	if lines[0] != "First scene content" {
		t.Errorf("line 0: got %q", lines[0])
	}
	if lines[1] != "Second scene content" {
		t.Errorf("line 1: got %q", lines[1])
	}
}

func TestStdoutExport_SkipsEmptyContent(t *testing.T) {
	roots := []SegmentNode{
		{TypeName: "Act", Number: "1", Content: ""},
		{TypeName: "Scene", Number: "1", Content: "Has content"},
		{TypeName: "Act", Number: "2", Content: ""},
	}
	e := &StdoutExporter{}
	out := captureStdout(func() {
		_ = e.Export("", roots, "")
	})
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 1 || lines[0] != "Has content" {
		t.Errorf("expected only content line, got: %q", out)
	}
}

func TestStdoutExport_DefaultExt(t *testing.T) {
	e := &StdoutExporter{}
	if ext := e.DefaultExt(); ext != "" {
		t.Errorf("expected empty DefaultExt, got %q", ext)
	}
}

func TestStdoutExport_EmptyTree(t *testing.T) {
	e := &StdoutExporter{}
	out := captureStdout(func() {
		if err := e.Export("", nil, ""); err != nil {
			t.Fatalf("Export error: %v", err)
		}
	})
	if out != "" {
		t.Errorf("expected no output for empty tree, got %q", out)
	}
}
