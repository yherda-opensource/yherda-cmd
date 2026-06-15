package config

import "testing"

func TestSaveAndLoadContext(t *testing.T) {
	withTempHome(t)

	cfg := &Config{
		ActiveWorkspace: "ws",
		ActiveIdea:      "idea-1",
		ActivePerson:    "person-1",
		ActiveArc:       "arc-1",
		ActivePlace:     "place-1",
		ActiveThing:     "thing-1",
	}
	if err := SaveConfig(cfg); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}
	loaded, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if loaded.ActiveIdea != "idea-1" {
		t.Errorf("active_idea: got %q", loaded.ActiveIdea)
	}
	if loaded.ActivePerson != "person-1" {
		t.Errorf("active_person: got %q", loaded.ActivePerson)
	}
	if loaded.ActiveArc != "arc-1" {
		t.Errorf("active_arc: got %q", loaded.ActiveArc)
	}
	if loaded.ActivePlace != "place-1" {
		t.Errorf("active_place: got %q", loaded.ActivePlace)
	}
	if loaded.ActiveThing != "thing-1" {
		t.Errorf("active_thing: got %q", loaded.ActiveThing)
	}
}

func TestCascadeReset_IdeaUse(t *testing.T) {
	withTempHome(t)

	cfg := &Config{
		ActiveWorkspace: "ws",
		ActiveIdea:      "old-idea",
		ActivePerson:    "old-person",
		ActiveArc:       "old-arc",
		ActivePlace:     "old-place",
		ActiveThing:     "old-thing",
	}
	if err := SaveConfig(cfg); err != nil {
		t.Fatal(err)
	}

	cfg.ActiveIdea = "new-idea"
	cfg.ActivePerson = ""
	cfg.ActiveArc = ""
	cfg.ActivePlace = ""
	cfg.ActiveThing = ""
	if err := SaveConfig(cfg); err != nil {
		t.Fatal(err)
	}

	loaded, _ := LoadConfig()
	if loaded.ActiveIdea != "new-idea" {
		t.Errorf("active_idea: got %q", loaded.ActiveIdea)
	}
	if loaded.ActivePerson != "" || loaded.ActiveArc != "" || loaded.ActivePlace != "" || loaded.ActiveThing != "" {
		t.Error("downstream context not cleared after idea use")
	}
}

func TestCascadeReset_PersonUse(t *testing.T) {
	withTempHome(t)

	cfg := &Config{
		ActiveWorkspace: "ws",
		ActiveIdea:      "idea-1",
		ActivePerson:    "old-person",
		ActiveArc:       "old-arc",
		ActivePlace:     "old-place",
		ActiveThing:     "old-thing",
	}
	if err := SaveConfig(cfg); err != nil {
		t.Fatal(err)
	}

	cfg.ActivePerson = "new-person"
	cfg.ActiveArc = ""
	cfg.ActivePlace = ""
	cfg.ActiveThing = ""
	if err := SaveConfig(cfg); err != nil {
		t.Fatal(err)
	}

	loaded, _ := LoadConfig()
	if loaded.ActiveIdea != "idea-1" {
		t.Errorf("active_idea should be unchanged: got %q", loaded.ActiveIdea)
	}
	if loaded.ActivePerson != "new-person" {
		t.Errorf("active_person: got %q", loaded.ActivePerson)
	}
	if loaded.ActiveArc != "" || loaded.ActivePlace != "" || loaded.ActiveThing != "" {
		t.Error("downstream context not cleared after person use")
	}
}

func TestCascadeReset_ArcUse(t *testing.T) {
	withTempHome(t)

	cfg := &Config{
		ActiveWorkspace: "ws",
		ActiveIdea:      "idea-1",
		ActivePerson:    "person-1",
		ActiveArc:       "old-arc",
		ActivePlace:     "old-place",
		ActiveThing:     "old-thing",
	}
	if err := SaveConfig(cfg); err != nil {
		t.Fatal(err)
	}

	cfg.ActiveArc = "new-arc"
	cfg.ActivePerson = "person-from-arc"
	cfg.ActivePlace = ""
	cfg.ActiveThing = ""
	if err := SaveConfig(cfg); err != nil {
		t.Fatal(err)
	}

	loaded, _ := LoadConfig()
	if loaded.ActiveArc != "new-arc" {
		t.Errorf("active_arc: got %q", loaded.ActiveArc)
	}
	if loaded.ActivePerson != "person-from-arc" {
		t.Errorf("active_person: got %q", loaded.ActivePerson)
	}
	if loaded.ActivePlace != "" || loaded.ActiveThing != "" {
		t.Error("downstream context not cleared after arc use")
	}
}

func TestCascadeReset_PlaceUse(t *testing.T) {
	withTempHome(t)

	cfg := &Config{
		ActiveWorkspace: "ws",
		ActiveIdea:      "idea-1",
		ActivePerson:    "person-1",
		ActiveArc:       "arc-1",
		ActivePlace:     "old-place",
		ActiveThing:     "old-thing",
	}
	if err := SaveConfig(cfg); err != nil {
		t.Fatal(err)
	}

	cfg.ActivePlace = "new-place"
	cfg.ActivePerson = ""
	cfg.ActiveArc = ""
	cfg.ActiveThing = ""
	if err := SaveConfig(cfg); err != nil {
		t.Fatal(err)
	}

	loaded, _ := LoadConfig()
	if loaded.ActivePlace != "new-place" {
		t.Errorf("active_place: got %q", loaded.ActivePlace)
	}
	if loaded.ActivePerson != "" || loaded.ActiveArc != "" || loaded.ActiveThing != "" {
		t.Error("downstream context not cleared after place use")
	}
}

func TestCascadeReset_ThingUse(t *testing.T) {
	withTempHome(t)

	cfg := &Config{
		ActiveWorkspace: "ws",
		ActiveIdea:      "idea-1",
		ActivePerson:    "person-1",
		ActiveArc:       "arc-1",
		ActivePlace:     "place-1",
		ActiveThing:     "old-thing",
	}
	if err := SaveConfig(cfg); err != nil {
		t.Fatal(err)
	}

	cfg.ActiveThing = "new-thing"
	cfg.ActivePerson = ""
	cfg.ActiveArc = ""
	cfg.ActivePlace = ""
	if err := SaveConfig(cfg); err != nil {
		t.Fatal(err)
	}

	loaded, _ := LoadConfig()
	if loaded.ActiveThing != "new-thing" {
		t.Errorf("active_thing: got %q", loaded.ActiveThing)
	}
	if loaded.ActivePerson != "" || loaded.ActiveArc != "" || loaded.ActivePlace != "" {
		t.Error("downstream context not cleared after thing use")
	}
}
