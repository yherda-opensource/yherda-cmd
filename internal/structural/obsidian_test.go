package structural

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func makeGraph(opts ...func(*IdeaGraph)) IdeaGraph {
	g := IdeaGraph{
		Idea:       map[string]any{"id": "1", "name": "Test Idea"},
		PluginData: map[string]any{},
	}
	for _, o := range opts {
		o(&g)
	}
	return g
}

func withIdentity(id, name string) func(*IdeaGraph) {
	return func(g *IdeaGraph) {
		g.Identities = append(g.Identities, map[string]any{"id": id, "name": name, "created": "2026-01-01"})
	}
}

func withDoc(id, entityType, entityID, content string) func(*IdeaGraph) {
	return func(g *IdeaGraph) {
		g.Docs = append(g.Docs, map[string]any{
			"id":          id,
			"entity_type": entityType,
			"entity_id":   entityID,
			"body":     content,
			"created":     "2026-01-01",
		})
	}
}

func withOrphanDoc(id, title, content string) func(*IdeaGraph) {
	return func(g *IdeaGraph) {
		g.Docs = append(g.Docs, map[string]any{
			"id":      id,
			"title":   title,
			"body": content,
			"created": "2026-01-01",
		})
	}
}

func exportTo(t *testing.T, graph IdeaGraph) string {
	t.Helper()
	dir := t.TempDir()
	output := filepath.Join(dir, "vault")
	e := &ObsidianExporter{}
	if err := e.Export(graph, output); err != nil {
		t.Fatalf("Export failed: %v", err)
	}
	return output
}

func TestObsidian_OutputDirCreated(t *testing.T) {
	dir := t.TempDir()
	output := filepath.Join(dir, "new-vault")
	e := &ObsidianExporter{}
	if err := e.Export(makeGraph(), output); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(output); err != nil {
		t.Errorf("output dir not created: %v", err)
	}
}


func TestObsidian_CreatesDirectoryLayout(t *testing.T) {
	graph := makeGraph(withIdentity("1", "Jo Ann Hayes"))
	graph.Arcs = []map[string]any{{"id": "10", "name": "The Reluctant Hero", "identity": "1", "created": "2026-01-01"}}
	graph.Beats = []map[string]any{{"id": "20", "name": "Act 1 Beat 1", "arc": "10", "identity": "1", "created": "2026-01-01"}}
	graph.Places = []map[string]any{{"id": "30", "name": "The Warehouse", "created": "2026-01-01"}}
	graph.Things = []map[string]any{{"id": "40", "name": "The Briefcase", "created": "2026-01-01"}}

	output := exportTo(t, graph)

	for _, sub := range []string{"identities", "arcs", "beats", "places", "things", "docs"} {
		if _, err := os.Stat(filepath.Join(output, sub)); err != nil {
			t.Errorf("missing subdir %s: %v", sub, err)
		}
	}
}

func TestObsidian_EntityFile(t *testing.T) {
	graph := makeGraph(
		withIdentity("1", "Jo Ann Hayes"),
		withDoc("d1", "identities", "1", "Her story begins in silence."),
	)
	output := exportTo(t, graph)

	data, err := os.ReadFile(filepath.Join(output, "identities", "jo-ann-hayes.md"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, "id: 1") {
		t.Error("expected front matter id")
	}
	if !strings.Contains(content, "Her story begins in silence.") {
		t.Error("expected doc body in file")
	}
}

func TestObsidian_EntityNoDoc(t *testing.T) {
	graph := makeGraph(withIdentity("1", "Jo Ann Hayes"))
	output := exportTo(t, graph)

	data, err := os.ReadFile(filepath.Join(output, "identities", "jo-ann-hayes.md"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	// Should have front matter but no body content after the closing ---
	parts := strings.SplitN(content, "---\n", 3)
	if len(parts) < 3 {
		t.Fatalf("unexpected front matter format: %q", content)
	}
	if strings.TrimSpace(parts[2]) != "" {
		t.Errorf("expected empty body for entity with no doc, got: %q", parts[2])
	}
}

func TestObsidian_TitleSanitization(t *testing.T) {
	graph := makeGraph(withIdentity("1", "Jo Ann Hayes!!! (The Hero)"))
	output := exportTo(t, graph)

	expected := filepath.Join(output, "identities", "jo-ann-hayes-the-hero.md")
	if _, err := os.Stat(expected); err != nil {
		t.Errorf("expected sanitized filename %s: %v", expected, err)
	}
}

func TestObsidian_TitleCollision(t *testing.T) {
	graph := makeGraph(
		withIdentity("1", "Same Name"),
		withIdentity("2", "Same Name"),
	)
	output := exportTo(t, graph)

	for _, name := range []string{"same-name.md", "same-name-2.md"} {
		if _, err := os.Stat(filepath.Join(output, "identities", name)); err != nil {
			t.Errorf("expected file %s: %v", name, err)
		}
	}
}

func TestObsidian_OrphanDocs(t *testing.T) {
	graph := makeGraph(withOrphanDoc("d1", "Worldbuilding Notes", "The world is vast."))
	output := exportTo(t, graph)

	data, err := os.ReadFile(filepath.Join(output, "docs", "worldbuilding-notes.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "The world is vast.") {
		t.Error("expected orphan doc body in docs/ file")
	}
}
