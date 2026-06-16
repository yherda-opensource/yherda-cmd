package config

import (
	"testing"
)

func withTempDir(t *testing.T) {
	t.Helper()
	t.Chdir(t.TempDir())
}

func TestLoadContext_FileNotExist(t *testing.T) {
	withTempDir(t)

	ctx, err := LoadContext()
	if err != nil {
		t.Fatalf("LoadContext on missing file: %v", err)
	}
	if ctx == nil {
		t.Fatal("LoadContext returned nil for missing file")
	}
	if ctx.Workspace != "" || ctx.Idea != "" {
		t.Errorf("expected empty context, got %+v", ctx)
	}
}

func TestSaveAndLoadContext(t *testing.T) {
	withTempDir(t)

	ctx := &Context{
		Workspace: "ws",
		APIServer: "https://ws.example.com",
		Idea:      "idea-1",
		Person:    "person-1",
		Arc:       "arc-1",
		Place:     "place-1",
		Thing:     "thing-1",
	}
	if err := SaveContext(ctx); err != nil {
		t.Fatalf("SaveContext: %v", err)
	}
	loaded, err := LoadContext()
	if err != nil {
		t.Fatalf("LoadContext: %v", err)
	}
	if loaded.Workspace != ctx.Workspace {
		t.Errorf("workspace: got %q, want %q", loaded.Workspace, ctx.Workspace)
	}
	if loaded.APIServer != ctx.APIServer {
		t.Errorf("api_server: got %q, want %q", loaded.APIServer, ctx.APIServer)
	}
	if loaded.Idea != ctx.Idea {
		t.Errorf("active_idea: got %q", loaded.Idea)
	}
	if loaded.Person != ctx.Person {
		t.Errorf("active_person: got %q", loaded.Person)
	}
	if loaded.Arc != ctx.Arc {
		t.Errorf("active_arc: got %q", loaded.Arc)
	}
	if loaded.Place != ctx.Place {
		t.Errorf("active_place: got %q", loaded.Place)
	}
	if loaded.Thing != ctx.Thing {
		t.Errorf("active_thing: got %q", loaded.Thing)
	}
}

func TestLoadContext_PartialFile(t *testing.T) {
	withTempDir(t)

	if err := SaveContext(&Context{Workspace: "ws"}); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadContext()
	if err != nil {
		t.Fatalf("LoadContext: %v", err)
	}
	if loaded.Workspace != "ws" {
		t.Errorf("workspace: got %q", loaded.Workspace)
	}
	if loaded.Idea != "" || loaded.Person != "" {
		t.Errorf("expected other fields empty, got %+v", loaded)
	}
}

func TestSaveContext_CascadeReset_IdeaUse(t *testing.T) {
	withTempDir(t)

	ctx := &Context{
		Workspace: "ws",
		Idea:      "old-idea",
		Person:    "old-person",
		Arc:       "old-arc",
		Place:     "old-place",
		Thing:     "old-thing",
	}
	if err := SaveContext(ctx); err != nil {
		t.Fatal(err)
	}

	ctx.Idea = "new-idea"
	ctx.Person = ""
	ctx.Arc = ""
	ctx.Place = ""
	ctx.Thing = ""
	if err := SaveContext(ctx); err != nil {
		t.Fatal(err)
	}

	loaded, _ := LoadContext()
	if loaded.Idea != "new-idea" {
		t.Errorf("active_idea: got %q", loaded.Idea)
	}
	if loaded.Person != "" || loaded.Arc != "" || loaded.Place != "" || loaded.Thing != "" {
		t.Error("downstream context not cleared after idea use")
	}
}
