# YOS-76: yherda model — Goal and Step commands, including goal use context

## Summary

Adds Goal/Step commands to the `yherda model` family, including a `goal use` context command mirroring the existing person/place/thing `use`-and-cascade pattern. Builds on YOS-74's scaffold.

## Commands in scope

```
yherda model goals list <subject-id>
yherda model goals create <subject-id> --want "..." [--need "..."] [--tragedy] [--description "..."] [--yes]
yherda model goal use <goal-id>
yherda model steps list [<goal-id>]
yherda model steps create [<goal-id>] --description "..." [--number N]
```

`[<goal-id>]` falls back to the active Goal in context when omitted.

## API mapping

| Command | Endpoint | Notes |
| --- | --- | --- |
| `model goals list/create` | `GET/POST /api/subject/{id}/goals/` | `SubjectPurposeMixin`. POST cascades `get_or_create_self()` + `get_or_create_purpose()` before creating the Goal. `GoalSerializer` fields: `want`, `need`, `tragedy` (bool), `description` — all writable, none required at the DB level, but `--want` is treated as practically-required in CLI UX (warns, doesn't block, when empty). |
| `model goal use` | none — local context only | Sets `ctx.Goal`. |
| `model steps list/create` | `GET/POST /api/goal/{id}/steps/` | `GoalView.steps`. `StepSerializer` fields: `description`, `number`. `goal` is server-set from the URL — the CLI never sends it in the POST body. List is ordered by `number` server-side. |

## Context integration

`internal/config/context.go`'s `Context` struct gains one new field: `Goal string \`json:"active_goal,omitempty"\``.

**Cascade-reset decision:** `goal use <id>` sets `ctx.Goal` only — does NOT clear Person/Place/Thing/State, same reasoning as YOS-74's `states use` (which Subject has the active Goal is orthogonal to which Person/Place/Thing is active).

## Confirmation prompt

`model goals create <subject-id>` on a Subject with no existing Self (`has_self: false` per `model show`) cascades a full Self → default Identity → that Identity's own Perspective/Disposition → Purpose stack. The command calls `GET /api/subject/{id}/` internally first to check `has_self`, and if false, prompts:

```
Subject #42 "The Study" (Place) has no Self yet — creating a Goal will also create its Self,
a default Identity, and that Identity's own Perspective/Disposition. Continue? [y/N]
```

`--yes`/`-y` skips the prompt (for scripts/agents). Skipped entirely when `has_self` is already true.

## Out of scope

Perspective/Belief attachment on a Goal's own Perspective — Goal is itself a Subject, so `model perspective get <goal-id>` (YOS-75) already covers this generically.

## Implementation notes

- New file `cmd/model_goal.go` — `model goals` (plural, resource CRUD), `model goal` (singular, `use` only, mirroring `person use`/`place use`), and `model steps` as three separate top-level subcommands under `model`, matching the design doc's flat command shape rather than nesting `steps` under `goal`.
- Added a shared `confirm(prompt string) bool` helper to `cmd/root.go`, reading from a package-level `confirmReader` (defaults to `os.Stdin`, swappable in tests) — the first interactive confirmation prompt in the CLI.
- `steps create`'s `--number` flag is only included in the POST body when explicitly passed (`cmd.Flags().Changed("number")`), matching the zero-value-vs-omitted handling already used for `perspective contexts`'s `--priority` in YOS-75.

Full ticket: https://momentsbyshawn.atlassian.net/browse/YOS-76
