# YOS-78: yherda model — list Subjects for an idea, and model use context

## Summary

Adds `yherda model list <idea-id>` — lists every Subject belonging to an idea, regardless of concrete subtype (Person/Place/Thing/Identity/Disposition/Goal/Step/Belief/...). Subject-generic, per YOS-73's governing principle. Also adds `yherda model use <subject-id>`, the generic Subject context slot YOS-74's design doc explicitly deferred.

## Backend (confirmed, already built — no gap)

`GET /api/idea/{id}/subjects/` — `IdeaView.subjects` action (`api/view/IdeaView.py:338-359`), landed under GEN-538/GEN-543. `Subject` has a direct FK to `Idea` (`api/model/IdeaModel.py:262-266`), so this is a single query, no per-subtype fan-out.

Optional query params: `subject_type` (exact match), `search` (case-insensitive name substring). Response: list of `SubjectSerializer` rows — `id`, `name`, `subject_type`, `has_perspective`, `has_self` (same shape as `model show`).

## Commands

```
yherda model list [<idea-id>] [--type <subject_type>] [--search <text>]
yherda model use <subject-id>
```

`<idea-id>` falls back to the active idea in context when omitted, matching `person list`/`place list`/`thing list` (including the "no active idea — showing ideas instead" fallback).

## API mapping

| Command | Endpoint | Notes |
| --- | --- | --- |
| `model list` | `GET /api/idea/{id}/subjects/` | `--type` maps to `?subject_type=`, `--search` maps to `?search=`. Both omitted from the query string entirely when not passed. |
| `model use` | none — local context only | Sets `ctx.Subject`. |

## Context integration

`internal/config/context.go`'s `Context` struct gains one new field: `Subject string \`json:"active_subject,omitempty"\``.

**Cascade-reset decision:** `model use <id>` sets `ctx.Subject` only — does NOT clear Person/Place/Thing/State/Goal, same reasoning as every other Subject-generic context slot (`states use`, `goal use`): the active generic Subject is orthogonal to the active Person/Place/Thing.

**What consumes it:** nothing yet, by design. No existing `model` command falls back to `ctx.Subject` — `model show`, `dispositions`, `states`, `perspective`, `belief`, `goals` all still require an explicit id argument. This story only adds the `use` command and the context slot; wiring fallback behavior into other commands is tracked separately (YOS-79, `model add perspective`/`model add goal`, will be the first real consumer).

## Output

`model list`: tabular `ID\tNAME\tTYPE\tHAS PERSPECTIVE\tHAS SELF`. `--json` returns the raw list.

## Out of scope

No new backend work (already shipped). No fallback wiring into other `model` commands (deferred to YOS-79).

## Implementation notes

- `model list` and `model use` added to the existing `cmd/model.go` (small additions, no new file).
- `model use` mirrors `model states use`/`model goal use`'s shape exactly — same non-clearing cascade-reset decision, same test pattern.

Full ticket: https://momentsbyshawn.atlassian.net/browse/YOS-78
