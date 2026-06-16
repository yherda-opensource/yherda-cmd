package export

import (
	"sort"
	"testing"
)

func TestGet_KnownFormats(t *testing.T) {
	for _, name := range []string{"stdout", "scriv"} {
		e, ok := Get(name)
		if !ok {
			t.Errorf("Get(%q) returned false", name)
		}
		if e == nil {
			t.Errorf("Get(%q) returned nil exporter", name)
		}
	}
}

func TestGet_Unknown(t *testing.T) {
	_, ok := Get("pdf")
	if ok {
		t.Error("Get(\"pdf\") should return false for unregistered format")
	}
}

func TestFormats_List(t *testing.T) {
	formats := Formats()
	if len(formats) == 0 {
		t.Fatal("Formats() returned empty list")
	}
	if !sort.StringsAreSorted(formats) {
		t.Errorf("Formats() result is not sorted: %v", formats)
	}
	seen := map[string]bool{}
	for _, f := range formats {
		if seen[f] {
			t.Errorf("duplicate format in Formats(): %s", f)
		}
		seen[f] = true
	}
}

func TestBuildTree_Nested(t *testing.T) {
	raw := []map[string]any{
		{
			"type_name": "Act",
			"number":    "1",
			"data":      []any{},
			"children": []any{
				map[string]any{
					"type_name": "Scene",
					"number":    "1",
					"data": []any{
						map[string]any{"plugin": "writer", "segment_data": "Hello world"},
					},
					"children": []any{},
				},
			},
		},
	}
	tree := BuildTree(raw)
	if len(tree) != 1 {
		t.Fatalf("expected 1 root, got %d", len(tree))
	}
	if tree[0].TypeName != "Act" {
		t.Errorf("expected TypeName=Act, got %s", tree[0].TypeName)
	}
	if len(tree[0].Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(tree[0].Children))
	}
	if tree[0].Children[0].Content != "Hello world" {
		t.Errorf("expected child Content=Hello world, got %s", tree[0].Children[0].Content)
	}
}

func TestBuildTree_Empty(t *testing.T) {
	tree := BuildTree(nil)
	if len(tree) != 0 {
		t.Errorf("expected empty tree, got %d nodes", len(tree))
	}
}

func TestBuildTree_PicksWriterPluginAmongMultiple(t *testing.T) {
	raw := []map[string]any{
		{
			"type_name": "Scene",
			"number":    "1",
			"data": []any{
				map[string]any{"plugin": "editor", "segment_data": "Editor notes"},
				map[string]any{"plugin": "writer", "segment_data": "Writer prose"},
			},
		},
	}
	tree := BuildTree(raw)
	if len(tree) != 1 {
		t.Fatalf("expected 1 node, got %d", len(tree))
	}
	if tree[0].Content != "Writer prose" {
		t.Errorf("expected writer plugin's content, got %q", tree[0].Content)
	}
}

func TestBuildTree_NoMatchingPlugin(t *testing.T) {
	raw := []map[string]any{
		{
			"type_name": "Scene",
			"number":    "1",
			"data": []any{
				map[string]any{"plugin": "editor", "segment_data": "Editor notes"},
			},
		},
	}
	tree := BuildTree(raw)
	if len(tree) != 1 {
		t.Fatalf("expected 1 node, got %d", len(tree))
	}
	if tree[0].Content != "" {
		t.Errorf("expected empty content when writer plugin absent, got %q", tree[0].Content)
	}
}
