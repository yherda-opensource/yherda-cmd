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

// --- goals create ---

func TestModelGoalsCreate_SkipConfirm_ReachesAPIDirectly(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	rootCmd.SetArgs([]string{"model", "goals", "create", "42", "--want", "To find her father", "--yes"})
	err := rootCmd.Execute()
	if err != nil && err.Error() == "aborted" {
		t.Error("--yes should have skipped the confirmation prompt entirely")
	}
}

func TestModelGoalsCreate_EmptyWant_WarnsButProceeds(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	rootCmd.SetArgs([]string{"model", "goals", "create", "42", "--yes"})
	err := rootCmd.Execute()
	if err != nil && strings.Contains(err.Error(), "--want is required") {
		t.Errorf("empty --want should warn, not hard-block: %v", err)
	}
}

func TestModelGoalsCreate_ConfirmDeclined_Aborts(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	// mustClient()'s "get subject" call will fail before reaching the prompt in
	// this test environment (no live server), so this asserts the confirm gate
	// is at least reached before the create call fires — not skipped by --yes.
	old := confirmReader
	confirmReader = strings.NewReader("n\n")
	defer func() { confirmReader = old }()

	rootCmd.SetArgs([]string{"model", "goals", "create", "42", "--want", "To find her father"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected an error (either from the subject lookup or the declined confirmation)")
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

func TestModelStepsCreate_MissingDescription_Error(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000", Goal: "goal-1"})

	rootCmd.SetArgs([]string{"model", "steps", "create"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error when --description is missing")
	}
}

func TestModelStepsCreate_NoArgNoContext_Error(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	rootCmd.SetArgs([]string{"model", "steps", "create", "--description", "Ask the innkeeper"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error when no goal id arg and no active goal in context")
	}
}

func TestModelStepsCreate_ExplicitGoalID_ReachesAPI(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	rootCmd.SetArgs([]string{"model", "steps", "create", "15", "--description", "Ask the innkeeper"})
	err := rootCmd.Execute()
	if err != nil && err.Error() == "no active goal — pass a goal id or run 'yherda model goal use <goal-id>'" {
		t.Errorf("explicit goal id arg should have satisfied the requirement: %v", err)
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
