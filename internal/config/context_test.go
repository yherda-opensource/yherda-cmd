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
		Place:     "place-1",
		Thing:     "thing-1",
		State:     "state-1",
		Goal:      "goal-1",
		Subject:   "subject-1",
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
	if loaded.Place != ctx.Place {
		t.Errorf("active_place: got %q", loaded.Place)
	}
	if loaded.Thing != ctx.Thing {
		t.Errorf("active_thing: got %q", loaded.Thing)
	}
	if loaded.State != ctx.State {
		t.Errorf("active_state: got %q, want %q", loaded.State, ctx.State)
	}
	if loaded.Goal != ctx.Goal {
		t.Errorf("active_goal: got %q, want %q", loaded.Goal, ctx.Goal)
	}
	if loaded.Subject != ctx.Subject {
		t.Errorf("active_subject: got %q, want %q", loaded.Subject, ctx.Subject)
	}
}

func TestSaveContext_StateOmittedWhenUnset(t *testing.T) {
	withTempDir(t)

	if err := SaveContext(&Context{Workspace: "ws"}); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadContext()
	if err != nil {
		t.Fatalf("LoadContext: %v", err)
	}
	if loaded.State != "" {
		t.Errorf("active_state should be empty when unset, got %q", loaded.State)
	}
	if loaded.Goal != "" {
		t.Errorf("active_goal should be empty when unset, got %q", loaded.Goal)
	}
	if loaded.Subject != "" {
		t.Errorf("active_subject should be empty when unset, got %q", loaded.Subject)
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
		Place:     "old-place",
		Thing:     "old-thing",
	}
	if err := SaveContext(ctx); err != nil {
		t.Fatal(err)
	}

	ctx.Idea = "new-idea"
	ctx.Person = ""
	ctx.Place = ""
	ctx.Thing = ""
	if err := SaveContext(ctx); err != nil {
		t.Fatal(err)
	}

	loaded, _ := LoadContext()
	if loaded.Idea != "new-idea" {
		t.Errorf("active_idea: got %q", loaded.Idea)
	}
	if loaded.Person != "" || loaded.Place != "" || loaded.Thing != "" {
		t.Error("downstream context not cleared after idea use")
	}
}
