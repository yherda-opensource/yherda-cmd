package cmd

import (
	"strings"
	"testing"

	"github.com/yherda-opensource/yherda-cmd/internal/config"
)

// --- goals list ---

func TestModelGoalsList_MissingID_Error(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	rootCmd.SetArgs([]string{"model", "goals", "list"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error when subject id arg is missing")
	}
}

// --- goal use ---

func TestModelGoalUse_SetsGoalOnly_DoesNotClearPersonPlaceThingState(t *testing.T) {
	withTempHome(t)
	saveContext(t, &config.Context{
		Workspace: "ws",
		Idea:      "idea-1",
		Person:    "person-1",
		Place:     "place-1",
		Thing:     "thing-1",
		State:     "state-1",
	})

	rootCmd.SetArgs([]string{"model", "goal", "use", "goal-1"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	loaded, _ := config.LoadContext()
	if loaded.Goal != "goal-1" {
		t.Errorf("active_goal: got %q, want %q", loaded.Goal, "goal-1")
	}
	if loaded.Person != "person-1" {
		t.Errorf("active_person should be unchanged, got %q", loaded.Person)
	}
	if loaded.Place != "place-1" {
		t.Errorf("active_place should be unchanged, got %q", loaded.Place)
	}
	if loaded.Thing != "thing-1" {
		t.Errorf("active_thing should be unchanged, got %q", loaded.Thing)
	}
	if loaded.State != "state-1" {
		t.Errorf("active_state should be unchanged, got %q", loaded.State)
	}
}

// --- steps list/create ---

func TestModelStepsList_NoArgNoContext_Error(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	rootCmd.SetArgs([]string{"model", "steps", "list"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error when no goal id arg and no active goal in context")
	}
}

func TestModelStepsList_ContextGoal_ReachesAPI(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000", Goal: "goal-1"})

	rootCmd.SetArgs([]string{"model", "steps", "list"})
	err := rootCmd.Execute()
	if err != nil && err.Error() == "no active goal — pass a goal id or run 'yherda model goal use <goal-id>'" {
		t.Errorf("active goal in context should have satisfied the requirement: %v", err)
	}
}

// --- confirm helper ---

func TestConfirm_YesVariants(t *testing.T) {
	old := confirmReader
	defer func() { confirmReader = old }()

	for _, input := range []string{"y\n", "Y\n", "yes\n", "YES\n", " yes \n"} {
		confirmReader = strings.NewReader(input)
		if !confirm("continue? ") {
			t.Errorf("expected confirm(%q) to be true", input)
		}
	}
}

func TestConfirm_NoOrEmpty(t *testing.T) {
	old := confirmReader
	defer func() { confirmReader = old }()

	for _, input := range []string{"n\n", "no\n", "\n", ""} {
		confirmReader = strings.NewReader(input)
		if confirm("continue? ") {
			t.Errorf("expected confirm(%q) to be false", input)
		}
	}
}
