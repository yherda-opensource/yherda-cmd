# YOS-79: yherda model add — perspective and goal on the active Subject (model use)

## Summary

Adds `yherda model add perspective` and `yherda model add goal` — the first commands to consume `ctx.Subject` (set by `model use`, YOS-78). Both operate against the active Subject by default, with an optional positional id override, matching the existing fallback pattern used by `model steps list [<goal-id>]` and `model states dispositions list [<state-id>]`.

`model add identity` is explicitly excluded — see the design conclusion below, not just the GEN-576 backend blocker.

## Commands in scope

```
yherda model add perspective [<subject-id>]
yherda model add goal [<subject-id>] --want "..." [--need "..."] [--tragedy] [--description "..."] [--yes]
```

`[<subject-id>]` falls back to `ctx.Subject` when omitted; errors clearly ("no active subject — pass a subject id or run 'yherda model use <subject-id>'") when neither is available.

## API mapping

| Command | Endpoint | Notes |
| --- | --- | --- |
| `model add perspective` | `POST /api/subject/{id}/perspective/` | Same call as YOS-75's `model perspective get` — a thin alias with `ctx.Subject` fallback, framed as "add" per the `model use` + `model add <capability>` mental model. Lazily materializes, idempotent. |
| `model add goal` | `POST /api/subject/{id}/goals/` | Identical to YOS-76's `model goals create <subject-id>`, including the has_self:false confirmation prompt and `--yes` skip flag — just resolves its target from `ctx.Subject` (with override) instead of requiring an explicit id argument. |

## `model add identity` — excluded by design, not just blocked

Perspective is the generic "this Subject can hold context/opinion" capability, cheap and universal. Identity is the character layer, only load-bearing once a Subject needs a Goal — which already cascades Self → Identity → Perspective on its own. A standalone "give this Subject an Identity" command may not be a real use case at all. GEN-576 (Identity creation is hardcoded to `Person` server-side, no Subject-generic path) remains filed as a low-priority backend gap, not a blocker for this story.

## Confirmation prompt

`model add goal` reuses YOS-76's exact confirmation logic (`GET /api/subject/{id}/` to check `has_self`, prompt if false, `--yes` to skip) — same code path, different target-resolution.

## Out of scope

`model add identity` (see above). Wiring `ctx.Subject` fallback into any other existing `model` command (`show`, `dispositions`, `states`, `perspective contexts`, `belief`) — that stays deliberately unconsumed elsewhere per YOS-78's decision, unless a specific need shows up.

## Implementation notes

- New file `cmd/model_add.go` — `modelAddCmd` ("add") with `perspective`/`goal` subcommands.
- New `resolveSubjectID(args []string) (string, error)` helper, same three-line shape as `resolveGoalID`/`resolveStateID` in `cmd/model.go`/`cmd/model_goal.go`.
- `model_goal.go`'s `modelGoalsCreateCmd` RunE was refactored to call a new shared `createGoalOnSubject(subjectID string, skipConfirm bool) error` — both `model goals create <subject-id>` and `model add goal [<subject-id>]` call it, differing only in how they resolve `subjectID`. No duplicated cascade/confirmation logic.
- `model add goal`'s flags (`--want`/`--need`/`--tragedy`/`--description`/`--yes`) bind to the same package-level vars as `model goals create`'s flags — safe since each is a distinct `cobra.Command` with its own flag set, parsed independently per invocation.

Full ticket: https://momentsbyshawn.atlassian.net/browse/YOS-79
