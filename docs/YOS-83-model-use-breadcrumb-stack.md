# YOS-83: model use should push a breadcrumb stack, not overwrite a single Subject

## Summary

`ctx.Subject` (`internal/config.Context`, previously `Subject string`) becomes a stack: `SubjectStack []string`, most-recent-last, JSON field `active_subject_stack` (replaces `active_subject`). `model use <id>` pushes onto it. A new `model back` command pops. Every existing consumer of `ctx.Subject` (`resolveSubjectID`, `printContextRow`, `printContextWithSubject`) reads the top of the stack instead of a flat field.

## `.yherda` file format — breaking change, no migration

`active_subject` (string) becomes `active_subject_stack` ([]string). This CLI has no stable release yet, so this is a breaking change with no migration/back-compat shim — an existing `.yherda` with the old field simply won't populate the new one (empty stack, same as never having run `model use`).

## Context API shape

Accessor methods on `*Context` rather than scattering `len(stack)-1` indexing across `cmd/`:

```go
type Context struct {
    ...
    SubjectStack []string `json:"active_subject_stack,omitempty"`
}

func (c *Context) Subject() string { ... }         // top of stack, or ""
func (c *Context) PushSubject(id string) { ... }   // append
func (c *Context) PopSubject() (id string, ok bool) { ... }
func (c *Context) ResetSubject(id string) { ... }  // collapse to a single id
```

`ctx.Subject` became `ctx.Subject()` (method, not field) at both call sites: `cmd/root.go` (`printContextRow`) and `cmd/model.go` (`resolveSubjectID`).

## `model use <id>` — push, and no-arg mode to inspect the stack

`Args: cobra.MaximumNArgs(1)`.

- With an id: `ctx.PushSubject(id)`, save, print `Active subject set to <id>` (unchanged message).
- With no arg: print the current breadcrumb trail (`Subject stack: 1 -> 7 -> 42 (active)`), or `No active subject stack.` if empty.

## New command: `model back`

Pops the stack. Popping the last remaining entry is allowed (leaves an empty stack, not an error) — only popping an *already-empty* stack errors ("no previous subject to go back to").

## Explicit id overrides reset the stack, not push

The one real behavior change beyond plumbing: `resolveSubjectID` becomes side-effecting. When an explicit arg is passed to any bare-Subject-id command (`model perspective get 99`, `model dispositions list 99`, etc.), the persisted stack resets to `[99]`, discarding whatever trail existed — a one-off override is a fresh start, not a deeper drill. This diverges from `resolveGoalID`/`resolveStateID`, which stay pure reads (Goal/State aren't stacks).

## Footer interaction (YOS-81/82)

No structural change — `printContextRow`'s `subject:` field and `printContextWithSubject`'s line 1 continue to show only the *current* top of stack, not the whole trail. The full trail is only shown by `model use` (no-arg mode).

## Cascade-reset interaction (YOS-74/76/78)

Confirmed: pushing/popping the Subject stack via `model use`/`model back` does not touch Person/Place/Thing/State/Goal, consistent with the existing decision that Subject is orthogonal to those.

## Files touched

- `internal/config/context.go` — `Subject string` → `SubjectStack []string` + accessor methods
- `cmd/model.go` — `resolveSubjectID` (side-effecting per above), `modelUseCmd` (push + no-arg inspect mode), new `modelBackCmd`
- `cmd/root.go` — `ctx.Subject` → `ctx.Subject()` at the one call site in `printContextRow`

## Out of scope

No change to `resolveGoalID`/`resolveStateID` (stay flat/pure). No change to `model states use`/`model goal use` (separate context slots, not part of this stack). No UI/breadcrumb rendering beyond the plain-text stack print in `model use`.

## Operational Requirements

No server-side change — entirely local CLI state (`.yherda` file format + in-memory context struct). No new network calls. The only "failure mode" is the breaking `.yherda` format change, which silently drops to an empty stack rather than crashing.

## Unit Test Plan

- `Context.PushSubject`/`PopSubject`/`ResetSubject`/`Subject()` — direct unit tests on the accessor methods.
- `model use <id>` twice in a row → stack has both ids, top is the second.
- `model use` with no args, empty stack → "No active subject stack" message, no error.
- `model use` with no args, non-empty stack → prints the full trail.
- `model back` with 2+ items → pops one, new top is previous item.
- `model back` with exactly 1 item → pops to empty, not an error.
- `model back` with 0 items → returns an error.
- `resolveSubjectID` with explicit arg + pre-existing multi-item stack → stack resets to `[arg]`.
- `resolveSubjectID` with no arg → unchanged behavior.

## Docs Update Plan

- README: updated the "Subject-generic commands are the exception" section to describe `model use`/`model back` as a stack, not a single slot, and note that explicit id overrides on other bare-Subject-id commands reset the stack.

Full ticket: https://momentsbyshawn.atlassian.net/browse/YOS-83
