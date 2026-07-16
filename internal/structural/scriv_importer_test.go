package structural

import (
	"os"
	"path/filepath"
	"testing"
)

// makeScrivFixture creates a minimal .scriv directory structure in a temp dir.
func makeScrivFixture(t *testing.T, scrivxContent string, dataFiles map[string]string) string {
	t.Helper()
	dir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, "test.scrivx"), []byte(scrivxContent), 0644); err != nil {
		t.Fatal(err)
	}

	for relPath, content := range dataFiles {
		full := filepath.Join(dir, relPath)
		if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	return dir
}

const minimalScrivx = `<?xml version="1.0" encoding="UTF-8"?>
<ScrivenerProject>
  <Binder>
    <BinderItem UUID="root" Type="Folder">
      <Title>root</Title>
      <Children>
        <BinderItem UUID="char-folder" Type="Folder">
          <Title>Characters</Title>
          <Children>
            <BinderItem UUID="char-001" Type="Text">
              <Title>Alice</Title>
              <Children></Children>
            </BinderItem>
            <BinderItem UUID="char-002" Type="Text">
              <Title>Bob</Title>
              <Children></Children>
            </BinderItem>
          </Children>
        </BinderItem>
        <BinderItem UUID="ms-folder" Type="Folder">
          <Title>Manuscript</Title>
          <Children>
            <BinderItem UUID="arc-001" Type="Folder">
              <Title>Act One</Title>
              <Children>
                <BinderItem UUID="beat-001" Type="Text">
                  <Title>Opening scene</Title>
                  <Children></Children>
                </BinderItem>
                <BinderItem UUID="beat-002" Type="Text">
                  <Title>Inciting incident</Title>
                  <Children></Children>
                </BinderItem>
              </Children>
            </BinderItem>
          </Children>
        </BinderItem>
        <BinderItem UUID="res-folder" Type="Folder">
          <Title>Research</Title>
          <Children>
            <BinderItem UUID="res-001" Type="Text">
              <Title>World notes</Title>
              <Children></Children>
            </BinderItem>
          </Children>
        </BinderItem>
        <BinderItem UUID="other-folder" Type="Folder">
          <Title>Front Matter</Title>
          <Children>
            <BinderItem UUID="other-001" Type="Text">
              <Title>Dedication</Title>
              <Children></Children>
            </BinderItem>
          </Children>
        </BinderItem>
      </Children>
    </BinderItem>
  </Binder>
</ScrivenerProject>`

func TestScrivImporter_CharacterFolder(t *testing.T) {
	dir := makeScrivFixture(t, minimalScrivx, nil)
	imp := &ScrivImporter{}
	graph, err := imp.Import(dir)
	if err != nil {
		t.Fatalf("Import error: %v", err)
	}

	if got := len(graph.Identities); got != 2 {
		t.Errorf("identities: got %d, want 2", got)
	}
	names := map[string]bool{}
	for _, id := range graph.Identities {
		names[strField(id, "name")] = true
	}
	if !names["Alice"] || !names["Bob"] {
		t.Errorf("expected Alice and Bob in identities, got %v", names)
	}
}

func TestScrivImporter_ManuscriptFolder(t *testing.T) {
	// GEN-555: Arc/Beat removed. Manuscript scenes are preserved as unmapped
	// idea documents pending a future structural rebuild.
	dir := makeScrivFixture(t, minimalScrivx, map[string]string{
		"Files/Data/beat-001/content.rtf": `{\rtf1 Opening content}`,
	})
	imp := &ScrivImporter{}
	graph, err := imp.Import(dir)
	if err != nil {
		t.Fatalf("Import error: %v", err)
	}

	var sceneTitles []string
	for _, doc := range graph.Docs {
		if strField(doc, "_unmapped") == "true" {
			sceneTitles = append(sceneTitles, strField(doc, "title"))
		}
	}
	if len(sceneTitles) != 3 {
		t.Fatalf("unmapped docs: got %d, want 3 (2 scenes + Dedication)", len(sceneTitles))
	}
	found := map[string]bool{}
	for _, title := range sceneTitles {
		found[title] = true
	}
	if !found["Opening scene"] || !found["Inciting incident"] {
		t.Errorf("expected manuscript scenes among unmapped docs, got %v", sceneTitles)
	}
}

func TestScrivImporter_ResearchFolder(t *testing.T) {
	dir := makeScrivFixture(t, minimalScrivx, nil)
	imp := &ScrivImporter{}
	graph, err := imp.Import(dir)
	if err != nil {
		t.Fatalf("Import error: %v", err)
	}

	var researchDocs []map[string]any
	for _, doc := range graph.Docs {
		if strField(doc, "_unmapped") != "true" {
			researchDocs = append(researchDocs, doc)
		}
	}
	if got := len(researchDocs); got != 1 {
		t.Errorf("research docs: got %d, want 1", got)
	}
	if got := strField(researchDocs[0], "title"); got != "World notes" {
		t.Errorf("research doc title: got %q, want %q", got, "World notes")
	}
}

func TestScrivImporter_UnmappedItemsPreserved(t *testing.T) {
	dir := makeScrivFixture(t, minimalScrivx, nil)
	imp := &ScrivImporter{}
	graph, err := imp.Import(dir)
	if err != nil {
		t.Fatalf("Import error: %v", err)
	}

	var unmapped []map[string]any
	for _, doc := range graph.Docs {
		if strField(doc, "_unmapped") == "true" {
			unmapped = append(unmapped, doc)
		}
	}
	// Manuscript scenes (2) are also unmapped now that Arc/Beat is gone (GEN-555),
	// plus Front Matter/Dedication.
	if got := len(unmapped); got != 3 {
		t.Errorf("unmapped docs: got %d, want 3", got)
	}
	found := map[string]bool{}
	for _, doc := range unmapped {
		found[strField(doc, "title")] = true
	}
	if !found["Dedication"] {
		t.Errorf("expected Dedication among unmapped docs, got %v", found)
	}
}

func TestScrivImporter_RTFContent(t *testing.T) {
	dir := makeScrivFixture(t, minimalScrivx, map[string]string{
		"Files/Data/beat-001/content.rtf": `{\rtf1\ansi Hello world}`,
	})
	imp := &ScrivImporter{}
	graph, err := imp.Import(dir)
	if err != nil {
		t.Fatalf("Import error: %v", err)
	}

	found := false
	for _, doc := range graph.Docs {
		if strField(doc, "title") == "Opening scene" {
			body := strField(doc, "body")
			if body == "" {
				t.Error("expected non-empty body for scene with RTF content")
			}
			found = true
		}
	}
	if !found {
		t.Error("scene 'Opening scene' not found in graph")
	}
}

func TestStripRTF(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{`{\rtf1\ansi Hello world}`, "Hello world"},
		{`{\rtf1 Plain text here}`, "Plain text here"},
		{`{\rtf1\b Bold\b0  normal}`, "Bold normal"},
	}
	for _, tc := range cases {
		got := stripRTF(tc.input)
		if got != tc.want {
			t.Errorf("stripRTF(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestScrivImporter_MissingScrivx(t *testing.T) {
	dir := t.TempDir()
	imp := &ScrivImporter{}
	_, err := imp.Import(dir)
	if err == nil {
		t.Error("expected error for directory with no .scrivx file")
	}
}
