# YOS-81: model list's idea-fallback output looks like a Subject list, causing silent wrong-id use in model show

## Summary

Found while dogfooding: `yherda model list` (no active idea) falls back to listing Ideas, printing a table shaped exactly like a Subject list. Copying an id from that table into `model show` silently returns a completely unrelated Subject, since Idea ids and Subject ids are separate id spaces that happen to collide numerically. Two fixes, both about visibility, not restriction — consistent with the epic's no-type-gating principle.

## Fix 1: Label `model list`'s idea-fallback output distinctly

`model list`'s fallback no longer delegates to `ideasListCmd.RunE`. It now calls `GET /idea/` directly and renders its own table with `IDEA ID` as the column header (not `ID`), plus an explicit warning line printed before the table: "Note: the ids below are IDEA ids, not Subject ids — they are a different id space and will not work with 'model show' or other model commands."

## Fix 2: Surface `subject_type` in every bare-Subject-id command's output

New shared helper `printSubjectContext(client *api.Client, subjectID string) error` in `cmd/root.go` — fetches `GET /api/subject/{id}/` and prints `Subject: #<id> "<name>" (<subject_type>)` before the command's own output. Skipped entirely in `--json` mode (a JSON caller already knows what it asked for; a text line would pollute the output). Not a confirmation prompt — no blocking, purely visibility.

Wired into every command that acts on a bare Subject id:
- `model dispositions list/create/delete`
- `model states list/create/delete`
- `model perspective get`
- `model goals list/create` (via `createGoalOnSubject`, shared with `model add goal`)
- `model belief create` (only when `--subject` is passed)
- `model add perspective`/`model add goal`

`createGoalOnSubject`'s existing has_self:false confirmation check already fetches the Subject — that GET's response is reused to print the confirmation line rather than fetching twice; the `--yes`/skip-confirm path calls `printSubjectContext` separately since no GET happens otherwise.

## Out of scope

No backend changes. No new confirmation prompts. No hard blocking of any id — every `model` command still accepts a bare Subject id and works on any subtype.

## Implementation notes

- `cmd/model.go`: `model list`'s fallback branch rewritten to call `/idea/` directly instead of delegating to `ideasListCmd.RunE`.
- `cmd/root.go`: new `printSubjectContext` helper, placed alongside `printContext`/`confirm`.
- Test file `cmd/model_subject_context_test.go` introduces `captureStdout` (an `os.Pipe`-based stdout capture helper) — a deliberate, minimal deviation from this suite's usual error-return-only assertion style, since this bug is specifically about output *text* being misleading, and the fix has to be verified by reading that text.

Full ticket: https://momentsbyshawn.atlassian.net/browse/YOS-81
