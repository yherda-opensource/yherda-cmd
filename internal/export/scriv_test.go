package export

import (
	"encoding/xml"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func fixtureTree() []SegmentNode {
	return []SegmentNode{
		{
			TypeName: "Act",
			Number:   "1",
			Content:  "",
			Children: []SegmentNode{
				{TypeName: "Scene", Number: "1", Content: "Scene one content"},
				{TypeName: "Scene", Number: "2", Content: "Scene two content"},
			},
		},
		{
			TypeName: "Act",
			Number:   "2",
			Content:  "Act with content",
			Children: []SegmentNode{},
		},
	}
}

func TestScrivExport_Structure(t *testing.T) {
	dir := t.TempDir()
	output := filepath.Join(dir, "Test.scriv")

	e := &ScrivExporter{}
	if err := e.Export("Test", fixtureTree(), output); err != nil {
		t.Fatalf("Export error: %v", err)
	}

	// Package directory must exist.
	if _, err := os.Stat(output); err != nil {
		t.Fatalf("output dir missing: %v", err)
	}

	// Files/Data/ must exist.
	dataDir := filepath.Join(output, "Files", "Data")
	if _, err := os.Stat(dataDir); err != nil {
		t.Fatalf("Files/Data/ missing: %v", err)
	}

	// Manifest must exist and be valid XML.
	scrivx := filepath.Join(output, "Test.scrivx")
	data, err := os.ReadFile(scrivx)
	if err != nil {
		t.Fatalf("manifest missing: %v", err)
	}
	var doc interface{}
	if err := xml.Unmarshal(data, &doc); err != nil {
		t.Errorf("manifest is not valid XML: %v", err)
	}

	// Each UUID subdirectory must have a content.rtf.
	entries, err := os.ReadDir(dataDir)
	if err != nil {
		t.Fatalf("could not read Files/Data/: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("no UUID directories in Files/Data/")
	}
	for _, entry := range entries {
		rtf := filepath.Join(dataDir, entry.Name(), "content.rtf")
		if _, err := os.Stat(rtf); err != nil {
			t.Errorf("missing content.rtf for UUID %s", entry.Name())
		}
	}
}

func TestScrivExport_TypeCalculation(t *testing.T) {
	cases := []struct {
		node     SegmentNode
		wantType string
	}{
		{
			node:     SegmentNode{Content: "", Children: []SegmentNode{{TypeName: "Scene"}}},
			wantType: "Folder",
		},
		{
			node:     SegmentNode{Content: "has content", Children: []SegmentNode{{TypeName: "Scene"}}},
			wantType: "Text",
		},
		{
			node:     SegmentNode{Content: "has content", Children: nil},
			wantType: "Text",
		},
		{
			node:     SegmentNode{Content: "", Children: nil},
			wantType: "Text",
		},
	}
	for _, tc := range cases {
		got := scrivType(tc.node)
		if got != tc.wantType {
			t.Errorf("scrivType(%+v) = %q, want %q", tc.node, got, tc.wantType)
		}
	}
}

func TestScrivExport_OutputExists(t *testing.T) {
	dir := t.TempDir()
	output := filepath.Join(dir, "existing.scriv")
	if err := os.Mkdir(output, 0755); err != nil {
		t.Fatal(err)
	}

	e := &ScrivExporter{}
	err := e.Export("Test", fixtureTree(), output)
	if err == nil {
		t.Fatal("expected error when output path already exists")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("expected 'already exists' in error, got: %v", err)
	}

	// Output must be unmodified (still just an empty dir with no .scrivx).
	entries, _ := os.ReadDir(output)
	if len(entries) != 0 {
		t.Errorf("output was modified despite error")
	}
}

func TestScrivExport_EmptyTree(t *testing.T) {
	dir := t.TempDir()
	output := filepath.Join(dir, "Empty.scriv")

	e := &ScrivExporter{}
	if err := e.Export("Empty", nil, output); err != nil {
		t.Fatalf("Export error: %v", err)
	}

	scrivx := filepath.Join(output, "Empty.scrivx")
	data, err := os.ReadFile(scrivx)
	if err != nil {
		t.Fatalf("manifest missing: %v", err)
	}
	var doc interface{}
	if err := xml.Unmarshal(data, &doc); err != nil {
		t.Errorf("manifest is not valid XML: %v", err)
	}
}

func TestScrivExport_DefaultExt(t *testing.T) {
	e := &ScrivExporter{}
	if ext := e.DefaultExt(); ext != "scriv" {
		t.Errorf("expected DefaultExt=scriv, got %q", ext)
	}
}
