# YOS-82: model use should be the default fallback for every bare-Subject-id command

## Summary

Every `model` command taking a bare Subject id falls back to `ctx.Subject` when no positional argument is passed, using the existing `resolveSubjectID` helper (moved from `cmd/model_add.go` to `cmd/model.go` alongside `resolveGoalID`/`resolveStateID`, since it's no longer `model add`-specific).

## Commands changed

| Command | Before | After |
| --- | --- | --- |
| `model show <id>` | `Args: cobra.ExactArgs(1)` | `Args: cobra.MaximumNArgs(1)`, resolves via `resolveSubjectID` |
| `model dispositions list/create/delete <id>` | same | same treatment |
| `model states list/create/delete <id>` | same | same treatment |
| `model perspective get <subject-id>` | same | same treatment |
| `model goals list/create <subject-id>` | same | same treatment |

`model belief create --subject` is unchanged — it's an optional relationship flag on a different primary resource (the Belief being created), not a command whose primary target is a Subject id.

## Why this wasn't done in YOS-78/79

Both design docs explicitly flagged it as a deliberate scope cut: "no command reads `ctx.Subject` yet... `model add` will be the first consumer." That was correct scoping at the time — `model use` was brand new and only had one real consumer planned. Now that `model use` is the actual daily workflow, leaving every other bare-Subject-id command unaware of it is the wrong default. Same shape as `model steps`/`model states dispositions` already establishes (`[<goal-id>]`/`[<state-id>]` optional, falls back to `ctx.Goal`/`ctx.State`) — this closes the gap for `ctx.Subject` to match.

## Footer interaction (YOS-81)

No conflict — `printContextWithSubject` already takes an explicit `subjectID` string regardless of how it was resolved (arg or `ctx.Subject`), so commands using the new fallback still get the same two-line footer treatment.

## Out of scope

`model belief create --subject` (see above). No changes to `model add perspective`/`model add goal` (already correct). No changes to `model use` itself.

## Operational Requirements

No new failure modes — this is strictly relaxing a requirement (arg becomes optional with a context fallback), not adding one. Same posture as every other `model` story: thin client over existing endpoints, no backend changes.

## Unit Test Plan

For each changed command: explicit id arg still works (existing tests already cover most of these); no arg + no `ctx.Subject` → clear error ("no active subject — pass a subject id or run 'yherda model use <subject-id>'"); no arg + `ctx.Subject` set → falls back correctly, reaches the API. Mirrors the existing `resolveGoalID`/`resolveStateID` test coverage pattern.

## Docs Update Plan

README: update the "Subject-generic commands are the exception" paragraph — `model show`/`dispositions`/`states`/`perspective`/`goals` no longer require an explicit id argument; only `model belief create --subject` (a flag, not a positional) is unaffected.

Full ticket: https://momentsbyshawn.atlassian.net/browse/YOS-82
