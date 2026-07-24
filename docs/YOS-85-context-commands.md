# YOS-85: Context Command Family for yherda-cmd

Design doc: https://momentsbyshawn.atlassian.net/wiki/spaces/SD/pages/2916354/YOS-85+Context+Command+Family+for+yherda-cmd

Child of epic YOS-49.

## Problem

Context is a Subject subtype used to attach and rank Beliefs (via BeliefContext) and to weight
what a Perspective surfaces first (via ContextPerspective/SubjectContext). The CLI has no way to
create, list, or select a Context, and no context-first way to link a Belief to one — the only
existing surface is `model belief contexts add`/`model perspective contexts add`, both of which
require an existing `--context <id>` that nothing in the CLI can produce.

## Confirmed backend shape

No backend changes are required — the create endpoint already exists, just unused by the CLI.

- `Context(Subject)` — plain Subject subtype, no extra fields, independent of any one
  Perspective/Identity (`api/model/BeliefSystemModel.py`)
- `POST /api/idea/{id}/contexts/` — creation, nested under Idea (`IdeaView.contexts` action), same
  pattern Identity uses nested under Person
- `GET /api/idea/{id}/contexts/` — list for an idea
- `GET/PATCH/DELETE /api/context/{id}/` — standalone retrieve/update/delete (`ContextView`,
  deliberately no create mixin — matches the nested-create pattern)
- `POST/GET/PATCH/DELETE /api/context/{id}/beliefs/` — BeliefContext join (belief id in body, not
  URL)

Confirmed by reading `yherda-idea-builder`'s `lib/api/contexts.ts` (`contextsApi`) and the
matching Django views/models directly.

## Approach

Add a new `context` command family, mirroring the existing `person` pattern (idea-scoped noun,
`list`/`create`/`use`), plus a `belief` sub-family for context-first belief linking.

### Commands

| Command | Behavior |
|---|---|
| `yherda context list [--idea <id>]` | GET `/idea/{id}/contexts/`, uses active idea unless `--idea` passed |
| `yherda context create --name "..." [--idea <id>]` | POST `/idea/{id}/contexts/`, sets active context + active idea on success (mirrors `person create`) |
| `yherda context use <id>` | Sets active context in `.yherda` context file |
| `yherda context belief add --belief <id> [--context <id>] [--status ...] [--mode ...]` | POST `/context/{id}/beliefs/`, uses active context unless `--context` passed |
| `yherda context belief list [--context <id>]` | GET `/context/{id}/beliefs/` |
| `yherda context belief remove --belief <id> [--context <id>]` | DELETE `/context/{id}/beliefs/` with belief id in body |

### Config changes

`internal/config/context.go`: add `Context string` (json `active_context,omitempty`) field,
following the existing Person/Place/Thing pattern. `context use` sets it directly (no cascade —
Context is independent, not a descendant of Person/Place/Thing in the breadcrumb sense).
`ideas use` should additionally clear `Context`, matching how it already clears
Person/Place/Thing (Contexts belong to an Idea, same as Person).

## Out of scope

- No backend changes — create endpoint already exists and works as-is.
- No changes to `model belief contexts` or `model perspective contexts` — those stay as the
  Subject-generic surface; `context belief` is an additive, context-first convenience wrapper
  over the same API, not a replacement.
- `context` does not get a `delete` command in this pass (no corresponding need identified yet;
  `ContextView` supports DELETE if this becomes needed later).

## Test plan

- Unit tests for each new command following the existing `cmd/person_test.go`-style pattern:
  active-context fallback, `--idea`/`--context` override, JSON output mode.
- Verify `context use` persists to `.yherda` and `ideas use` clears it.

## Docs

Update yherda-cmd's README command list to include `context`, per the documentation-process rule.
