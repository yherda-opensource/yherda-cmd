# YOS-81: model list's idea-fallback output looks like a Subject list, causing silent wrong-id use in model show

## Summary

Found while dogfooding: `yherda model list` (no active idea) falls back to listing Ideas, printing a table shaped exactly like a Subject list. Copying an id from that table into `model show` silently returns a completely unrelated Subject, since Idea ids and Subject ids are separate id spaces that happen to collide numerically. Two fixes, both about visibility, not restriction — consistent with the epic's no-type-gating principle.

## Fix 1: Label `model list`'s idea-fallback output distinctly

`model list`'s fallback no longer delegates to `ideasListCmd.RunE`. It now calls `GET /idea/` directly and renders its own table with `IDEA ID` as the column header (not `ID`), plus an explicit warning line printed before the table: "Note: the ids below are IDEA ids, not Subject ids — they are a different id space and will not work with 'model show' or other model commands."

## Fix 2 + 3: Two-line footer — the record acted on, then full context

Revised during review (Shawn: "it could really be 2 lines. First line the record used after whatever the command was. Second line is a row of context showing all the ids... hopefully names if we have them"). Combined into one two-line footer, printed *after* a command's own output (not a pre-output confirmation line — that was the first pass, superseded here):

- **Line 1** — the Subject record the command just acted on: `Subject: #<id> "<name>" (<subject_type>)`.
- **Line 2** — the usual `context: workspace: ... | idea: ... | ...` row, with the `subject` field now showing the full row (id, name, subject_type, has_perspective, has_self — the same fields `model list` shows), not just the bare id.

New helpers in `cmd/root.go`:
- `printContextWithSubject(client *api.Client, subjectID string)` — prints both lines. Fetches `subjectID` once and reuses that fetch for line 2's `subject` field when it matches the active `ctx.Subject`, rather than fetching twice.
- `subjectContextLabel(subjectID string) string` — used by plain `printContext()` (no acted-on Subject in scope) and by `printContextWithSubject` when the acted-on Subject differs from `ctx.Subject`. Fetched live, not cached, so a renamed Subject doesn't show stale data; falls back silently to the bare id if the fetch fails.
- `printContextRow(ctx *config.Context, subjectLabel func(string) string)` — the shared field-list/join logic both `printContext()` and `printContextWithSubject` render through.

Both skipped in `--json`/`--no-context` mode.

Wired into every command that acts on a bare Subject id (replacing the earlier `printSubjectContext` pre-output call):
- `model dispositions list/create/delete`
- `model states list/create/delete`
- `model perspective get`
- `model goals list/create` (via `createGoalOnSubject`, shared with `model add goal`)
- `model belief create` (only when `--subject` is passed)
- `model add perspective`/`model add goal`

`createGoalOnSubject`'s has_self:false confirmation prompt still fetches the Subject and prints its own inline warning before the cascade — that's a distinct, necessary pre-action prompt, not the footer. The footer's `printContextWithSubject` call happens once at the end, after the Goal is created, regardless of whether confirmation ran.

This applies globally — any command that calls `printContext()` (not just `model` commands) shows the active Subject's full row in line 2 once `model use` has been run, the same way `state`/`goal`/etc. already appear once set.

## Out of scope

No backend changes. No new confirmation prompts beyond the pre-existing has_self:false one. No hard blocking of any id — every `model` command still accepts a bare Subject id and works on any subtype. Names for the *other* context fields (idea, person, place, thing, state, goal) are out of scope for this ticket — only Subject gets the name lookup; expanding the rest is a separate, bigger change (each would need its own fetch).

## Implementation notes

- `cmd/model.go`: `model list`'s fallback branch rewritten to call `/idea/` directly instead of delegating to `ideasListCmd.RunE`.
- `cmd/root.go`: `printContextWithSubject`, `subjectContextLabel`, `printContextRow` — the earlier `printSubjectContext` (pre-output, per-command confirmation) was replaced entirely, not kept alongside.
- Test file `cmd/model_subject_context_test.go` introduces `captureStdout` (an `os.Pipe`-based stdout capture helper) — a deliberate, minimal deviation from this suite's usual error-return-only assertion style, since this bug is specifically about output *text* being misleading, and the fix has to be verified by reading that text.

Full ticket: https://momentsbyshawn.atlassian.net/browse/YOS-81
