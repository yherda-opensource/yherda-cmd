# YOS-75: yherda model — Perspective and Belief creation on any Subject

## Summary

Adds Perspective-Context commands (Subject-generic, per YOS-73's governing principle) and Belief creation/Context-attachment commands to the `yherda model` family. Builds on YOS-74's scaffold. Corrected scope per insight_identity_perspective_inherits_from_self_owner.md: Belief attaches to Context via `BeliefContext`, not to an Identity/Subject directly — `IdentityView.contexts`/`adopt_belief` are stale/superseded, not used here.

## Commands in scope

```
yherda model perspective get <subject-id>
yherda model perspective contexts list <perspective-id>
yherda model perspective contexts add <perspective-id> --context <context-id> [--priority N]
yherda model perspective contexts update <perspective-id> --context <context-id> --priority N
yherda model perspective contexts remove <perspective-id> --context <context-id>

yherda model belief create --idea <idea-id> --statement "..." [--subject <subject-id>]
yherda model belief contexts add --context <context-id> --belief <belief-id> [--status Emerging|Active|Strained|Former] [--mode Masking|Surfacing]
yherda model belief contexts update --context <context-id> --belief <belief-id> [--status ...] [--mode ...]
yherda model belief contexts remove --context <context-id> --belief <belief-id>

yherda model dispositions contexts list|add|update|remove <disposition-id>   [BLOCKED on GEN-574, not implemented]
```

## API mapping

| Command | Endpoint | Notes |
| --- | --- | --- |
| `model perspective get` | `POST /api/subject/{id}/perspective/` | `SubjectPerspectiveMixin`. Lazily materializes; response is `{id, name}`. POST despite being a "get" from the CLI's perspective. |
| `model perspective contexts list/add/update/remove` | `GET/POST/PATCH/DELETE /api/perspective/{id}/contexts/` | `PerspectiveView.contexts`. Uses `ContextPerspective` — Perspective's own ranked view of Contexts, independent of `SubjectContext.priority`. `add` takes `context` id + optional `priority`; `update` requires `priority`; all require `context` id. |
| `model belief create` | `POST /api/idea/{id}/beliefs/` | `IdeaView.beliefs`. Fields: `statement` (required), `subject` (optional, belief-about-subject FK), `name` (optional, auto-generated from `statement` if omitted). Idea-scoped root create, same pattern as Person/Place/Thing — NOT created via `/api/subject/.../`. |
| `model belief contexts add/update/remove` | `GET/POST/PATCH/DELETE /api/context/{id}/beliefs/` | `ContextView.beliefs`. Uses `BeliefContext` — `status` (default Active) and `mode` (default Surfacing) both settable on POST, both independently patchable. |
| `model dispositions contexts *` | none yet | **Blocked on GEN-574** — no Subject-generic `/api/subject/{id}/contexts/` endpoint exists. Not implemented in this story. |

## Belief-attachment specificity rule

Not enforced by the CLI — documented in `README.md` and command help text as guidance: unqualified beliefs generally attach at a Subject's own (or base) Perspective; fully-qualified ones ("X, in identity Y, when Z") attach at the specific Identity/Disposition's own Perspective.

## Out of scope / blocked

- Disposition-Context attachment blocked on GEN-574 (backend).
- Priority-ranking semantics for `perspective contexts` beyond what `ContextPerspective.priority` already provides.
- Runtime Perspective resolution (composite merge) — unbuilt on the backend (GEN-534, deferred).

## Implementation notes

- New files: `cmd/model_perspective.go`, `cmd/model_belief.go`.
- `perspective contexts add`/`update` use `cmd.Flags().Changed("priority")` rather than a zero-value check, since `0` is a valid priority.
- `belief create`'s `--subject` flag is omitted from the request body entirely when not passed (not sent as `""`), matching `Belief.subject` being `null=True, blank=True`.
- `belief contexts add/update`'s `--status`/`--mode` flags are likewise omitted from the body when not passed, letting the server apply its own defaults (Active/Surfacing) rather than the CLI hardcoding them.

Full ticket: https://momentsbyshawn.atlassian.net/browse/YOS-75
