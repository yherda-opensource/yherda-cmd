package cmd

import (
	"testing"

	"github.com/yherda-opensource/yherda-cmd/internal/config"
)

// --- model show ---

func TestModelShow_ReachesAPI(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	rootCmd.SetArgs([]string{"model", "show", "42"})
	err := rootCmd.Execute()
	if err != nil && err.Error() == "not logged in — run 'yherda login' first" {
		t.Errorf("credentials should have satisfied login requirement: %v", err)
	}
}

func TestModelShow_MissingID_Error(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	rootCmd.SetArgs([]string{"model", "show"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error when id arg is missing")
	}
}

// --- model list ---

func TestModelList_NoIdeaContext_FallsBackToIdeaList(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	rootCmd.SetArgs([]string{"model", "list"})
	err := rootCmd.Execute()
	if err != nil && err.Error() == "no active idea — run: yherda ideas use <id>" {
		t.Error("should fall back to listing ideas rather than returning a hard error")
	}
}

func TestModelList_ExplicitIdeaID_ReachesAPI(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	rootCmd.SetArgs([]string{"model", "list", "some-idea"})
	err := rootCmd.Execute()
	if err != nil && err.Error() == "not logged in — run 'yherda login' first" {
		t.Errorf("credentials should have satisfied login requirement: %v", err)
	}
}

func TestModelList_ContextIdea_ReachesAPI(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000", Idea: "idea-1"})

	rootCmd.SetArgs([]string{"model", "list"})
	err := rootCmd.Execute()
	if err != nil && err.Error() == "not logged in — run 'yherda login' first" {
		t.Errorf("active idea in context should have satisfied the requirement: %v", err)
	}
}

func TestModelList_WithTypeAndSearchFlags_ReachesAPI(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000", Idea: "idea-1"})

	rootCmd.SetArgs([]string{"model", "list", "--type", "Belief", "--search", "king"})
	err := rootCmd.Execute()
	if err != nil && err.Error() == "not logged in — run 'yherda login' first" {
		t.Errorf("flags should not interfere with the request: %v", err)
	}
}

// --- model use ---

func TestModelUse_SetsSubjectOnly_DoesNotClearPersonPlaceThingStateGoal(t *testing.T) {
	withTempHome(t)
	saveContext(t, &config.Context{
		Workspace: "ws",
		Idea:      "idea-1",
		Person:    "person-1",
		Place:     "place-1",
		Thing:     "thing-1",
		State:     "state-1",
		Goal:      "goal-1",
	})

	rootCmd.SetArgs([]string{"model", "use", "subject-1"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	loaded, _ := config.LoadContext()
	if loaded.Subject != "subject-1" {
		t.Errorf("active_subject: got %q, want %q", loaded.Subject, "subject-1")
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
	if loaded.Goal != "goal-1" {
		t.Errorf("active_goal should be unchanged, got %q", loaded.Goal)
	}
}

func TestModelUse_MissingID_Error(t *testing.T) {
	withTempHome(t)
	saveContext(t, &config.Context{Workspace: "ws"})

	rootCmd.SetArgs([]string{"model", "use"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error when subject id arg is missing")
	}
}

// --- model dispositions create ---

func TestModelDispositionsCreate_MissingType_Error(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	rootCmd.SetArgs([]string{"model", "dispositions", "create", "42", "--name", "Grieving"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error when --type is missing")
	}
}

func TestModelDispositionsCreate_InvalidType_Error(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	rootCmd.SetArgs([]string{"model", "dispositions", "create", "42", "--type", "bogus", "--name", "Grieving"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error for invalid --type value")
	}
}

func TestModelDispositionsCreate_MissingName_Error(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	rootCmd.SetArgs([]string{"model", "dispositions", "create", "42", "--type", "emotional"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error when --name is missing")
	}
}

func TestModelDispositionsCreate_ValidType_ReachesAPI(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	rootCmd.SetArgs([]string{"model", "dispositions", "create", "42", "--type", "physical", "--name", "Injured"})
	err := rootCmd.Execute()
	if err != nil && err.Error() == "--type must be one of physical, emotional, mental, spiritual" {
		t.Errorf("valid type should have passed client-side validation: %v", err)
	}
}

// --- model dispositions delete ---

func TestModelDispositionsDelete_MissingFlag_Error(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	rootCmd.SetArgs([]string{"model", "dispositions", "delete", "42"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error when --disposition is missing")
	}
}

// --- model states create/delete ---

func TestModelStatesCreate_MissingName_Error(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	rootCmd.SetArgs([]string{"model", "states", "create", "42"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error when --name is missing")
	}
}

func TestModelStatesDelete_MissingFlag_Error(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	rootCmd.SetArgs([]string{"model", "states", "delete", "42"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error when --state is missing")
	}
}

// --- model states use ---

func TestModelStatesUse_SetsStateOnly_DoesNotClearPersonPlaceThing(t *testing.T) {
	withTempHome(t)
	saveContext(t, &config.Context{
		Workspace: "ws",
		Idea:      "idea-1",
		Person:    "person-1",
		Place:     "place-1",
		Thing:     "thing-1",
	})

	rootCmd.SetArgs([]string{"model", "states", "use", "state-1"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	loaded, _ := config.LoadContext()
	if loaded.State != "state-1" {
		t.Errorf("active_state: got %q, want %q", loaded.State, "state-1")
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
}

// --- model states dispositions ---

func TestModelStatesDispositionsSet_NoArgNoContext_Error(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	rootCmd.SetArgs([]string{"model", "states", "dispositions", "set", "12"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error when no state id arg and no active state in context")
	}
}

func TestModelStatesDispositionsSet_ExplicitStateID_ReachesAPI(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	rootCmd.SetArgs([]string{"model", "states", "dispositions", "set", "12", "7"})
	err := rootCmd.Execute()
	if err != nil && err.Error() == "no active state — pass a state id or run 'yherda model states use <state-id>'" {
		t.Errorf("explicit state id arg should have satisfied the requirement: %v", err)
	}
}

func TestModelStatesDispositionsSet_ContextState_ReachesAPI(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000", State: "state-1"})

	rootCmd.SetArgs([]string{"model", "states", "dispositions", "set", "12"})
	err := rootCmd.Execute()
	if err != nil && err.Error() == "no active state — pass a state id or run 'yherda model states use <state-id>'" {
		t.Errorf("active state in context should have satisfied the requirement: %v", err)
	}
}

func TestModelStatesDispositionsUnset_NoArgNoContext_Error(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	rootCmd.SetArgs([]string{"model", "states", "dispositions", "unset", "12"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error when no state id arg and no active state in context")
	}
}

func TestModelStatesDispositionsList_NoArgNoContext_Error(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	rootCmd.SetArgs([]string{"model", "states", "dispositions", "list"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error when no state id arg and no active state in context")
	}
}

func TestModelStatesDispositionsList_ContextState_ReachesAPI(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000", State: "state-1"})

	rootCmd.SetArgs([]string{"model", "states", "dispositions", "list"})
	err := rootCmd.Execute()
	if err != nil && err.Error() == "no active state — pass a state id or run 'yherda model states use <state-id>'" {
		t.Errorf("active state in context should have satisfied the requirement: %v", err)
	}
}
