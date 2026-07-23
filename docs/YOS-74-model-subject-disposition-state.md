# YOS-74: yherda model — base Subject show + generic Disposition/State commands

## Summary

Adds a new top-level `model` command family to yherda-cmd, operating on the platform's Subject base class (GEN-517) rather than per-type commands. Foundational story for the whole `yherda model` family (YOS-73) — YOS-75/76 build on the scaffold this introduces. Per the epic's governing principle: commands take a bare Subject id and call a Subject-generic capability, no client-side type gating.

## Commands in scope

```
yherda model show <id>
yherda model dispositions list <id>
yherda model dispositions create <id> --type physical|emotional|mental|spiritual --name "..."
yherda model dispositions delete <id> --disposition <disposition-id>
yherda model states list <id>
yherda model states create <id> --name "..."
yherda model states delete <id> --state <state-id>
yherda model states use <state-id>
yherda model states dispositions list [<state-id>]
yherda model states dispositions set <disposition-id> [<state-id>]
yherda model states dispositions unset <disposition-id> [<state-id>]
```

`[<state-id>]` falls back to the active State in context when omitted.

## API mapping

| Command | Endpoint | Notes |
| --- | --- | --- |
| `model show` | `GET /api/subject/{id}/` | `SubjectView`, retrieve-only. Response: `id`, `name`, `subject_type`, `has_perspective`, `has_self`. |
| `model dispositions list/create/delete` | `GET/POST/DELETE /api/subject/{id}/dispositions/` | `SubjectDispositionMixin`. Create: `type` + `name`. Delete: `disposition` id in body. |
| `model states list/create/delete` | `GET/POST/DELETE /api/subject/{id}/states/` | `SubjectStateMixin`. Create: `name`. Delete: `state` id in body. |
| `model states dispositions list/set/unset` | `GET/POST/DELETE /api/state/{id}/dispositions/` | `StateView.dispositions`. Server enforces one-per-type and same-Subject scoping via `StateDisposition.clean()` — CLI surfaces the 400 directly, no client-side pre-check. |

`model states use` writes local context only, no API call.

## Context integration

`internal/config/context.go`'s `Context` struct gains one new field: `State string \`json:"active_state,omitempty"\``.

No `model use <id>` / new generic `Subject` context slot in this story — `model show`/`dispositions`/`states` require an explicit id argument. A generic Subject slot is deferred until YOS-75/76 demonstrate real need. `model states use <state-id>` sets `ctx.State` only — does not clear Person/Place/Thing, since a State is scoped to whichever Subject already active, not a replacement for it.

## Confirmation prompts

This story's commands (Disposition/State CRUD, State-Disposition bundling) are low-stakes — no confirmation prompts here. Reserved for YOS-76's `goals create`.

## Out of scope

Perspective/Belief (YOS-75), Goal/Step (YOS-76), Identity Perspective-Context attachment (YOS-77, abandoned), Disposition-priority-within-State ranking (blocked on GEN-575).

## Implementation notes

- Added `Client.Delete(path string, body any) error` to `internal/api/client.go` — the client previously only had Get/Post/Patch.
- New file `cmd/model.go` holds the entire command family (show, dispositions, states, states dispositions) rather than splitting into multiple files, since the family is small and tightly related; can be split later if it grows.
- `resolveStateID` centralizes the state-id-arg-or-context-fallback logic shared by `states dispositions list/set/unset`.

Full ticket: https://momentsbyshawn.atlassian.net/browse/YOS-74
