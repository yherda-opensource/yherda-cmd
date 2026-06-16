package cmd

import (
	"testing"

	"github.com/yherda-opensource/yherda-cmd/internal/config"
)

// --- identity create ---

func TestIdentityCreate_MissingName_Error(t *testing.T) {
	withTempHome(t)
	saveContext(t, &config.Config{ActiveWorkspace: "ws", APIServer: "https://ws.yherda.test:8000", ActivePerson: "person-1"})

	rootCmd.SetArgs([]string{"identity", "create"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error when --name is missing")
	}
}

func TestIdentityCreate_NoPersonNoContext_Error(t *testing.T) {
	withTempHome(t)
	saveContext(t, &config.Config{ActiveWorkspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	rootCmd.SetArgs([]string{"identity", "create", "--name", "Hero"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error when no --person and no active context")
	}
}

func TestIdentityCreate_ContextPerson_ReachesAPI(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Config{ActiveWorkspace: "ws", APIServer: "https://ws.yherda.test:8000", ActivePerson: "person-1"})

	rootCmd.SetArgs([]string{"identity", "create", "--name", "Hero"})
	err := rootCmd.Execute()
	// network call will fail; we only verify it didn't fail due to missing context
	if err != nil && (err.Error() == "--person is required (or set active person with 'yherda person use <id>')" ||
		err.Error() == "--name is required") {
		t.Errorf("context should have satisfied requirements, got: %v", err)
	}
}

// --- arc create ---

func TestArcCreate_MissingWant_Error(t *testing.T) {
	withTempHome(t)
	saveContext(t, &config.Config{ActiveWorkspace: "ws", APIServer: "https://ws.yherda.test:8000", ActivePerson: "person-1"})

	rootCmd.SetArgs([]string{"arc", "create"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error when --want is missing")
	}
}

func TestArcCreate_NoPersonNoContext_Error(t *testing.T) {
	withTempHome(t)
	saveContext(t, &config.Config{ActiveWorkspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	rootCmd.SetArgs([]string{"arc", "create", "--want", "To find the truth"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error when no --person and no active context")
	}
}

// --- beat create ---

func TestBeatCreate_MissingDescription_Error(t *testing.T) {
	withTempHome(t)
	saveContext(t, &config.Config{ActiveWorkspace: "ws", APIServer: "https://ws.yherda.test:8000", ActiveArc: "arc-1"})

	rootCmd.SetArgs([]string{"beat", "create"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error when --description is missing")
	}
}

func TestBeatCreate_NoArcNoContext_Error(t *testing.T) {
	withTempHome(t)
	saveContext(t, &config.Config{ActiveWorkspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	rootCmd.SetArgs([]string{"beat", "create", "--description", "The call to adventure"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error when no --arc and no active context")
	}
}

func TestBeatCreate_ContextArc_ReachesAPI(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Config{ActiveWorkspace: "ws", APIServer: "https://ws.yherda.test:8000", ActiveArc: "arc-1"})

	rootCmd.SetArgs([]string{"beat", "create", "--description", "The call to adventure"})
	err := rootCmd.Execute()
	if err != nil && (err.Error() == "--arc is required (or set active arc with 'yherda arc use <id>')" ||
		err.Error() == "--description is required") {
		t.Errorf("context should have satisfied requirements, got: %v", err)
	}
}

// --- place create ---

func TestPlaceCreate_MissingName_Error(t *testing.T) {
	withTempHome(t)
	saveContext(t, &config.Config{ActiveWorkspace: "ws", APIServer: "https://ws.yherda.test:8000", ActiveIdea: "idea-1"})

	rootCmd.SetArgs([]string{"place", "create"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error when --name is missing")
	}
}

func TestPlaceCreate_NoIdeaNoContext_Error(t *testing.T) {
	withTempHome(t)
	saveContext(t, &config.Config{ActiveWorkspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	rootCmd.SetArgs([]string{"place", "create", "--name", "The Forest"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error when no --idea and no active context")
	}
}

func TestPlaceCreate_ContextIdea_ReachesAPI(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Config{ActiveWorkspace: "ws", APIServer: "https://ws.yherda.test:8000", ActiveIdea: "idea-1"})

	rootCmd.SetArgs([]string{"place", "create", "--name", "The Forest"})
	err := rootCmd.Execute()
	if err != nil && (err.Error() == "--idea is required (or set active idea with 'yherda ideas use <id>')" ||
		err.Error() == "--name is required") {
		t.Errorf("context should have satisfied requirements, got: %v", err)
	}
}

// --- setting create ---

func TestSettingCreate_MissingName_Error(t *testing.T) {
	withTempHome(t)
	saveContext(t, &config.Config{ActiveWorkspace: "ws", APIServer: "https://ws.yherda.test:8000", ActivePlace: "place-1"})

	rootCmd.SetArgs([]string{"setting", "create"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error when --name is missing")
	}
}

func TestSettingCreate_NoPlaceNoContext_Error(t *testing.T) {
	withTempHome(t)
	saveContext(t, &config.Config{ActiveWorkspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	rootCmd.SetArgs([]string{"setting", "create", "--name", "The Tavern"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error when no --place and no active context")
	}
}

func TestSettingCreate_ContextPlace_ReachesAPI(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Config{ActiveWorkspace: "ws", APIServer: "https://ws.yherda.test:8000", ActivePlace: "place-1"})

	rootCmd.SetArgs([]string{"setting", "create", "--name", "The Tavern"})
	err := rootCmd.Execute()
	if err != nil && (err.Error() == "--place is required (or set active place with 'yherda place use <id>')" ||
		err.Error() == "--name is required") {
		t.Errorf("context should have satisfied requirements, got: %v", err)
	}
}

// --- thing create ---

func TestThingCreate_MissingName_Error(t *testing.T) {
	withTempHome(t)
	saveContext(t, &config.Config{ActiveWorkspace: "ws", APIServer: "https://ws.yherda.test:8000", ActiveIdea: "idea-1"})

	rootCmd.SetArgs([]string{"thing", "create"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error when --name is missing")
	}
}

func TestThingCreate_NoIdeaNoContext_Error(t *testing.T) {
	withTempHome(t)
	saveContext(t, &config.Config{ActiveWorkspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	rootCmd.SetArgs([]string{"thing", "create", "--name", "The Sword"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error when no --idea and no active context")
	}
}

func TestThingCreate_ContextIdea_ReachesAPI(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Config{ActiveWorkspace: "ws", APIServer: "https://ws.yherda.test:8000", ActiveIdea: "idea-1"})

	rootCmd.SetArgs([]string{"thing", "create", "--name", "The Sword"})
	err := rootCmd.Execute()
	if err != nil && (err.Error() == "--idea is required (or set active idea with 'yherda ideas use <id>')" ||
		err.Error() == "--name is required") {
		t.Errorf("context should have satisfied requirements, got: %v", err)
	}
}
