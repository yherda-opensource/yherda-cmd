package config

import (
	"os"
	"path/filepath"
	"testing"
)

func withTempHome(t *testing.T) {
	t.Helper()
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
}

func TestSaveAndLoadCredentials(t *testing.T) {
	withTempHome(t)

	creds := &Credentials{
		AccessToken:  "access-abc",
		RefreshToken: "refresh-xyz",
		TokenType:    "Bearer",
	}
	if err := SaveCredentials(creds); err != nil {
		t.Fatalf("SaveCredentials: %v", err)
	}

	loaded, err := LoadCredentials()
	if err != nil {
		t.Fatalf("LoadCredentials: %v", err)
	}
	if loaded == nil {
		t.Fatal("LoadCredentials returned nil")
	}
	if loaded.AccessToken != creds.AccessToken {
		t.Errorf("access_token: got %q, want %q", loaded.AccessToken, creds.AccessToken)
	}
	if loaded.RefreshToken != creds.RefreshToken {
		t.Errorf("refresh_token: got %q, want %q", loaded.RefreshToken, creds.RefreshToken)
	}
	if loaded.TokenType != creds.TokenType {
		t.Errorf("token_type: got %q, want %q", loaded.TokenType, creds.TokenType)
	}
}

func TestLoadCredentials_Missing(t *testing.T) {
	withTempHome(t)

	creds, err := LoadCredentials()
	if err != nil {
		t.Fatalf("LoadCredentials on missing file: %v", err)
	}
	if creds != nil {
		t.Errorf("expected nil for missing credentials, got %+v", creds)
	}
}

func TestDeleteCredentials(t *testing.T) {
	withTempHome(t)

	if err := SaveCredentials(&Credentials{AccessToken: "x"}); err != nil {
		t.Fatalf("SaveCredentials: %v", err)
	}
	if err := DeleteCredentials(); err != nil {
		t.Fatalf("DeleteCredentials: %v", err)
	}

	creds, err := LoadCredentials()
	if err != nil {
		t.Fatalf("LoadCredentials after delete: %v", err)
	}
	if creds != nil {
		t.Error("expected nil after delete")
	}
}

func TestDeleteCredentials_Idempotent(t *testing.T) {
	withTempHome(t)

	if err := DeleteCredentials(); err != nil {
		t.Errorf("DeleteCredentials on non-existent file should not error: %v", err)
	}
}

func TestCredentialsFilePermissions(t *testing.T) {
	withTempHome(t)

	if err := SaveCredentials(&Credentials{AccessToken: "secret"}); err != nil {
		t.Fatalf("SaveCredentials: %v", err)
	}

	home, _ := os.UserHomeDir()
	path := filepath.Join(home, ".yherdacmd", "credentials.json")
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat credentials file: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("credentials file permissions: got %o, want 0600", info.Mode().Perm())
	}
}

func TestSaveAndLoadConfig(t *testing.T) {
	withTempHome(t)

	cfg := &Config{ActiveWorkspace: "my-workspace"}
	if err := SaveConfig(cfg); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}

	loaded, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if loaded.ActiveWorkspace != cfg.ActiveWorkspace {
		t.Errorf("active_workspace: got %q, want %q", loaded.ActiveWorkspace, cfg.ActiveWorkspace)
	}
}

func TestLoadConfig_Missing(t *testing.T) {
	withTempHome(t)

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig on missing file: %v", err)
	}
	if cfg == nil {
		t.Fatal("LoadConfig returned nil for missing file")
	}
	if cfg.ActiveWorkspace != "" {
		t.Errorf("expected empty workspace, got %q", cfg.ActiveWorkspace)
	}
}
