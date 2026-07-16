package cmd

import (
	"testing"

	"github.com/yherda-opensource/yherda-cmd/internal/config"
)

// --- identity create ---

func TestIdentityCreate_MissingName_Error(t *testing.T) {
	withTempHome(t)
	saveContext(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000", Person: "person-1"})

	rootCmd.SetArgs([]string{"identity", "create"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error when --name is missing")
	}
}

func TestIdentityCreate_NoPersonNoContext_Error(t *testing.T) {
	withTempHome(t)
	saveContext(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	rootCmd.SetArgs([]string{"identity", "create", "--name", "Hero"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error when no --person and no active context")
	}
}

func TestIdentityCreate_ContextPerson_ReachesAPI(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000", Person: "person-1"})

	rootCmd.SetArgs([]string{"identity", "create", "--name", "Hero"})
	err := rootCmd.Execute()
	if err != nil && (err.Error() == "--person is required (or set active person with 'yherda person use <id>')" ||
		err.Error() == "--name is required") {
		t.Errorf("context should have satisfied requirements, got: %v", err)
	}
}

// --- place create ---

func TestPlaceCreate_MissingName_Error(t *testing.T) {
	withTempHome(t)
	saveContext(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000", Idea: "idea-1"})

	rootCmd.SetArgs([]string{"place", "create"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error when --name is missing")
	}
}

func TestPlaceCreate_NoIdeaNoContext_Error(t *testing.T) {
	withTempHome(t)
	saveContext(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	rootCmd.SetArgs([]string{"place", "create", "--name", "The Forest"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error when no --idea and no active context")
	}
}

func TestPlaceCreate_ContextIdea_ReachesAPI(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000", Idea: "idea-1"})

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
	saveContext(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000", Place: "place-1"})

	rootCmd.SetArgs([]string{"setting", "create"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error when --name is missing")
	}
}

func TestSettingCreate_NoPlaceNoContext_Error(t *testing.T) {
	withTempHome(t)
	saveContext(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	rootCmd.SetArgs([]string{"setting", "create", "--name", "The Tavern"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error when no --place and no active context")
	}
}

func TestSettingCreate_ContextPlace_ReachesAPI(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000", Place: "place-1"})

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
	saveContext(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000", Idea: "idea-1"})

	rootCmd.SetArgs([]string{"thing", "create"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error when --name is missing")
	}
}

func TestThingCreate_NoIdeaNoContext_Error(t *testing.T) {
	withTempHome(t)
	saveContext(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000"})

	rootCmd.SetArgs([]string{"thing", "create", "--name", "The Sword"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error when no --idea and no active context")
	}
}

func TestThingCreate_ContextIdea_ReachesAPI(t *testing.T) {
	withTempHome(t)
	saveContextWithCreds(t, &config.Context{Workspace: "ws", APIServer: "https://ws.yherda.test:8000", Idea: "idea-1"})

	rootCmd.SetArgs([]string{"thing", "create", "--name", "The Sword"})
	err := rootCmd.Execute()
	if err != nil && (err.Error() == "--idea is required (or set active idea with 'yherda ideas use <id>')" ||
		err.Error() == "--name is required") {
		t.Errorf("context should have satisfied requirements, got: %v", err)
	}
}
