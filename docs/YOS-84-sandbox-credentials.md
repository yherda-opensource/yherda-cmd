# YOS-84: YHERDA_SANDBOX env var for a directory-local, hidden credentials store

## Summary

`YHERDA_SANDBOX=1` makes `internal/config.credentialsDir()` resolve to a hidden directory-local path (`./.yherdacmd/`, in the current working directory) instead of `~/.yherdacmd/`. This is the single choke point all three credential functions (`LoadCredentials`, `SaveCredentials`, `DeleteCredentials`) already call through, so the change is entirely inside `internal/config/config.go` — no changes needed to `internal/auth/` (`Login`/`RefreshTokens` return `*Credentials` and never touch the filesystem themselves) or to any `cmd/` call site.

## Why this unblocks the `model/local` + `model/prod` setup

Both directories already have their own `.envrc` controlling `YHERDA_PUBLIC_HOST` (prod relies on the default, local overrides it). Credentials are the one piece still global — logging in under one directory overwrites the same `~/.yherdacmd/credentials.json` the other reads, so `cd`-switching changes which API host is hit but not which login is active. Adding `export YHERDA_SANDBOX=1` to `model/local/.envrc` (prod stays on the default `~/.yherdacmd/`) gives local a fully independent, directory-scoped login.

## Implementation

```go
// internal/config/config.go
func credentialsDir() (string, error) {
	var d string
	if os.Getenv("YHERDA_SANDBOX") == "1" {
		d = filepath.Join(".", ".yherdacmd")
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		d = filepath.Join(home, ".yherdacmd")
	}
	if err := os.MkdirAll(d, 0700); err != nil {
		return "", err
	}
	return d, nil
}
```

Matches the existing `os.Getenv("YHERDA_INSECURE") == "1"` convention already used in `internal/auth/pkce.go` and `internal/api/client.go`.

## Scope: credentials only, not `.yherda` context

`~/.yherdacmd/` today holds *only* `credentials.json`, so "relocate credentials" and "relocate everything under `~/.yherdacmd/`" are the same change. The separate `.yherda` context file (workspace/idea/person/.../Subject stack) already lives in the current working directory unconditionally — untouched by this change, already directory-scoped from the start.

## `.gitignore`

Added `.yherdacmd/` alongside the existing `.yherda` entry, so anyone using `YHERDA_SANDBOX=1` inside a git-tracked directory doesn't risk committing live OAuth tokens.

## Backward compatibility

Default (unset `YHERDA_SANDBOX`) behavior is unchanged — the new branch only activates on the explicit opt-in.

## Out of scope

No change to `internal/auth/`. No change to `.yherda` context file handling. No flag-based alternative (env var only). No auto-detection convention.

## Operational Requirements

No server-side change. Purely local filesystem behavior — `os.MkdirAll` with `0700` (unchanged permissions posture). Opt-in only, so no risk to default behavior.

## Unit Test Plan

- `credentialsDir()` with `YHERDA_SANDBOX` unset → returns `~/.yherdacmd` (regression-covered).
- `credentialsDir()` with `YHERDA_SANDBOX=1` → returns `./.yherdacmd` relative to cwd.
- `SaveCredentials`/`LoadCredentials` round-trip under `YHERDA_SANDBOX=1` in a temp dir → confirms it doesn't touch `~/.yherdacmd/`.
- `LoadCredentials` under `YHERDA_SANDBOX=1` with no file present → returns `(nil, nil)`.
- Sandboxed and home-dir saves land in genuinely different files, not the same one.

## Docs Update Plan

- README: added `YHERDA_SANDBOX` to the "Environment variables" table and the "Local state" section.
- `.gitignore`: added `.yherdacmd/`.

Full ticket: https://momentsbyshawn.atlassian.net/browse/YOS-84
