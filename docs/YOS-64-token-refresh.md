# YOS-64: Auth token not refreshed — CLI fails with 401 after 1 hour

## Problem

`internal/api/client.go` `do()` makes a request with the stored `access_token`. On a 401 it returns the error immediately. The stored `refresh_token` is never used. The user must re-run `yherda login` manually after ~1 hour.

## Approach

Add a `RefreshFunc` to the `Client` that is called on a 401 response before giving up. The refresh hits the OAuth token endpoint with `grant_type=refresh_token`, persists the new credentials, updates the in-memory `creds` on the client, and retries the original request once.

The refresh logic lives in `internal/auth/pkce.go` (alongside the existing `exchangeCode`). The `Client` in `internal/api/client.go` receives a `RefreshFunc` at construction time — this keeps the `api` package from importing `auth` (avoiding a circular dep) and makes the client testable without OAuth machinery.

## Changes

**`internal/auth/pkce.go`**
- Add `RefreshTokens(refreshToken string) (*config.Credentials, error)` — posts `grant_type=refresh_token` to `/oauth/token/`, returns new credentials.

**`internal/api/client.go`**
- Add `RefreshFunc func() (*config.Credentials, error)` field to `Client`.
- In `do()`: after a 401, call `RefreshFunc()` if set, update `c.creds`, retry the request once. If `RefreshFunc` is nil or the refresh itself fails, return a message prompting `yherda login`.

**`cmd/root.go`**
- In `mustClient()` and `mustPublicClient()`: wire a closure as the `RefreshFunc` that calls `auth.RefreshTokens(creds.RefreshToken)` then `config.SaveCredentials(...)`.

## Key Decisions

- **Retry once, not in a loop.** Prevents infinite loops on revoked/expired refresh tokens.
- **No circular import.** `api` package does not import `auth`. The closure is injected from `cmd/root.go`.
- **`RefreshFunc` is optional.** Clients without it retain current behavior — 401 = error.
- **Persist immediately.** New credentials saved before the retry so the next command uses the fresh token even if the retry fails.

## Operational Requirements ✅

### Healthy State
All CLI commands succeed transparently after access token expiry. Refresh happens inline: 401 triggers token exchange, credentials written to `~/.yherdacmd/credentials.json`, original request retried. User sees no error.

### Failure State
- **Expired/revoked refresh token:** refresh returns non-200. User sees: `error: session expired — run 'yherda login' to re-authenticate`.
- **Network failure during refresh:** transport error surfaced with login prompt.
- **Silent failure impossible by design:** retry-once pattern ensures success or a visible error.

### Observability Signal
- `DEVELOPER=1` stderr logs `[dev] METHOD URL` — 401 → retry sequence visible there.
- `~/.yherdacmd/credentials.json` mtime confirms whether refresh occurred.

### Performance Expectation
One extra round-trip (~200–500ms) once per hour per user session. Invisible to the user in practice.

### Alerting Triggers
- Refresh token expired: user-facing error + login prompt. No alert needed.
- Token endpoint 5xx: visible to user; systemic failures appear in Heroku/Django logs.

### Notification Path
No automated alerting for the CLI. Failures surface as terminal errors.

### Gaps
None.

## Unit Test Plan

- **`internal/auth/pkce.go`:** `TestRefreshTokens` — table-driven against `httptest.Server`: success (200), bad refresh token (400), network error.
- **`internal/api/client.go`:**
  - `TestDo_401_TriggersRefreshAndRetry` — server returns 401 then 200; assert `RefreshFunc` called once, final response 200.
  - `TestDo_401_RefreshFails` — server returns 401; `RefreshFunc` errors; assert error returned, no second HTTP call.
  - `TestDo_401_NoRefreshFunc` — no `RefreshFunc` set; 401 returned as error (backward compat).

## Docs Update Plan

- `README.md` — no change needed; refresh is transparent to users.
