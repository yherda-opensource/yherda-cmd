**YOS-86** — CLI companion to GEN-627 (backend, `python` repo, PR #148, not yet merged). GEN-627 caps `yherda-cmd`'s server-side `OAuthApplicationScopeAllowlist` to `idea_owner_read collaborator_read collaborator_write` — the CLI's current hardcoded `scope = "read write"` (`internal/auth/pkce.go:24`) will be rejected as `invalid_scope` once that allowlist change is live (confirmed 2026-08-14 testing against a GEN-627-branch backend locally).

**Extended 2026-08-14**: Shawn also wants the ability to log in *without* requesting `idea_owner_read` for a given CLI instance — an opt-out at login time, not a hardcoded scope string. Collaborator scope stays defaulted and unconditional — only the Idea Owner read portion is optional.

## Scope of this ticket

1. Change the default requested scope from `read write` to `idea_owner_read collaborator_read collaborator_write`.
2. Add a `yherda login` flag to opt out of requesting `idea_owner_read` for that login — `--idea-owner-read=false`. When set, the login flow requests only `collaborator_read collaborator_write`.
3. No flag to opt out of Collaborator scope — that's always requested, unconditionally; the CLI has no purpose without it post-GEN-627.
4. No flag to request `idea_owner_write` or Workspace Owner scope — the backend's allowlist for this client permanently excludes both regardless of what's requested, so exposing a CLI flag for them would just produce a confusing `invalid_scope` error. Not built.

## Current state (verified in code, 2026-08-14)

- `internal/auth/pkce.go:21-25` — `scope` was a package-level `const`, used identically in `buildAuthURL` (`/oauth/authorize/`, line 123) and *not* re-sent on `RefreshTokens` (`/oauth/token/` with `grant_type=refresh_token`, lines 173-212) — refresh requests don't carry a `scope` param at all, so whatever scope the original token was issued with persists across refreshes automatically (OAuth2 refresh-token semantics: omitting `scope` on refresh preserves the original grant). This means the opt-out only needed to affect the initial `Login()` call, not `RefreshTokens()`.
- `cmd/auth.go:11-32` — `loginCmd` was a bare cobra command, no flags. `RunE` called `auth.Login()` with no arguments and no way to pass options through.
- `internal/config.Credentials` (`internal/config/config.go:9-13`) — stores `AccessToken`/`RefreshToken`/`TokenType` only; does not persist which scopes were actually granted. The access token itself is opaque to the CLI (a bearer string) — the CLI has no client-side record of what scope a stored credential carries beyond what the server enforces per-request.
- Existing flag convention in this codebase: package-level var + `Flags().BoolVar(&var, "name", default, help)` in `init()` (e.g. `cmd/context.go:242-254`).

## Implementation (shipped as designed)

`internal/auth/pkce.go`: `scope` const replaced with two named constants (`collaboratorScope`, `ideaOwnerReadScope`) and a `requestedScope(includeIdeaOwnerRead bool) string` helper. `Login()` now takes `includeIdeaOwnerRead bool`, threaded through to `buildAuthURL(redirectURI, challenge, includeIdeaOwnerRead)`.

`cmd/auth.go`: `loginCmd` gained `--idea-owner-read` (bool, default `true`), registered in a new `init()` in that file, passed through as `auth.Login(loginIncludeIdeaOwnerRead)`.

`RefreshTokens()` unchanged — no scope param needed, per the "Current state" note above.

## Not addressed by this design, flagged for later refinement

- Whether the granted scope should be persisted anywhere client-side (e.g. in `Credentials` or a separate file) so other commands can tell a Collaborator-only login apart from a full login without making a server round-trip that 403s. Today nothing surfaces this distinction to the user after login succeeds.
- `yherda login` doesn't detect/warn when re-logging-in with a *different* scope selection than a previously stored credential.

## Sequencing

This CLI change is only meaningfully testable/mergeable once GEN-627 (backend) has merged to `main` and `register_oauth_app --app yherda-cmd --owner-email <email>` has been re-run against whatever environment the CLI under test points at. Per Shawn (2026-08-14): GEN-627's backend scoping will ship to prod alongside a working Human Factor release — 0.3.0 of the CLI is understood to not work against prod until that happens, and that's an accepted tradeoff since Shawn is currently the CLI's only user.
