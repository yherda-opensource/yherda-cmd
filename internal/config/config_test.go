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

// --- YHERDA_SANDBOX (YOS-84) ---

func TestCredentialsDir_SandboxUnset_UsesHomeDir(t *testing.T) {
	withTempHome(t)

	if err := SaveCredentials(&Credentials{AccessToken: "x"}); err != nil {
		t.Fatalf("SaveCredentials: %v", err)
	}

	home, _ := os.UserHomeDir()
	path := filepath.Join(home, ".yherdacmd", "credentials.json")
	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected credentials at %s, stat failed: %v", path, err)
	}
}

func TestCredentialsDir_SandboxSet_UsesCwd(t *testing.T) {
	withTempHome(t)
	t.Setenv("YHERDA_SANDBOX", "1")
	t.Chdir(t.TempDir())

	if err := SaveCredentials(&Credentials{AccessToken: "x"}); err != nil {
		t.Fatalf("SaveCredentials: %v", err)
	}

	path := filepath.Join(".", ".yherdacmd", "credentials.json")
	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected credentials at %s, stat failed: %v", path, err)
	}

	home, _ := os.UserHomeDir()
	homePath := filepath.Join(home, ".yherdacmd", "credentials.json")
	if _, err := os.Stat(homePath); err == nil {
		t.Error("sandboxed save should not have touched ~/.yherdacmd/")
	}
}

func TestLoadCredentials_SandboxSet_NoFile_ReturnsNil(t *testing.T) {
	withTempHome(t)
	t.Setenv("YHERDA_SANDBOX", "1")
	t.Chdir(t.TempDir())

	creds, err := LoadCredentials()
	if err != nil {
		t.Fatalf("LoadCredentials: %v", err)
	}
	if creds != nil {
		t.Errorf("expected nil for missing sandboxed credentials, got %+v", creds)
	}
}

func TestCredentials_SandboxAndHome_AreIndependent(t *testing.T) {
	withTempHome(t)
	cwd := t.TempDir()

	if err := SaveCredentials(&Credentials{AccessToken: "home-token"}); err != nil {
		t.Fatalf("SaveCredentials (home): %v", err)
	}

	t.Setenv("YHERDA_SANDBOX", "1")
	t.Chdir(cwd)
	if err := SaveCredentials(&Credentials{AccessToken: "sandbox-token"}); err != nil {
		t.Fatalf("SaveCredentials (sandbox): %v", err)
	}

	sandboxCreds, err := LoadCredentials()
	if err != nil {
		t.Fatalf("LoadCredentials (sandbox): %v", err)
	}
	if sandboxCreds.AccessToken != "sandbox-token" {
		t.Errorf("sandboxed token = %q, want %q", sandboxCreds.AccessToken, "sandbox-token")
	}

	t.Setenv("YHERDA_SANDBOX", "")
	homeCreds, err := LoadCredentials()
	if err != nil {
		t.Fatalf("LoadCredentials (home): %v", err)
	}
	if homeCreds.AccessToken != "home-token" {
		t.Errorf("home token = %q, want %q (sandboxed save should not have overwritten it)", homeCreds.AccessToken, "home-token")
	}
}
