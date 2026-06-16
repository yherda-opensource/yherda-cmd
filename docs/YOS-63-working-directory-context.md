# Design Doc: YOS-63 — Thread-safe context: working directory state

## Problem

All mutable CLI state is currently persisted to `~/.yherdacmd/config.json` — a single shared file. Two terminals or two parallel agent processes on the same workstation stomp each other's context. One Claude instance updating an idea in one window, a user working in another window, or two agents in separate workspaces simultaneously all collide on that file.

## Approach

Split state by its natural scope:

- `~/.yherdacmd/credentials.json` — global, shared. One auth per workstation. Unchanged.
- `./.yherda` in the current working directory — local context for this project/session.

Each terminal or agent process operates in its own directory and gets independent context naturally. No `eval`, no shell wrappers, no env var ceremony — `yherda ideas use <id>` just works and writes to the local `.yherda` file. Two agents in different directories never touch each other's state.

`config.json` is removed entirely.

## What moves where

| Current field | New location |
|---|---|
| `ActiveWorkspace` | `./.yherda` |
| `APIServer` | `./.yherda` |
| `ActiveIdea` | `./.yherda` |
| `ActivePerson` | `./.yherda` |
| `ActiveArc` | `./.yherda` |
| `ActivePlace` | `./.yherda` |
| `ActiveThing` | `./.yherda` |
| Credentials | `~/.yherdacmd/credentials.json` (unchanged) |

## `internal/config/config.go`

Remove `Config` struct, `SaveConfig`, `LoadConfig`, and the `dir()` helper for the global config dir. Keep all credentials functions unchanged.

## `internal/config/context.go` (new file)

```go
type Context struct {
    Workspace string `json:"active_workspace,omitempty"`
    APIServer string `json:"api_server,omitempty"`
    Idea      string `json:"active_idea,omitempty"`
    Person    string `json:"active_person,omitempty"`
    Arc       string `json:"active_arc,omitempty"`
    Place     string `json:"active_place,omitempty"`
    Thing     string `json:"active_thing,omitempty"`
}

func LoadContext() (*Context, error)  // reads ./.yherda
func SaveContext(*Context) error      // writes ./.yherda
```

Reads and writes `.yherda` in the current working directory. If the file doesn't exist, returns an empty `Context` (no error).

## Command behavior

All `use` commands and `workspace` write to `./.yherda` as before — same syntax, same UX, just a different file location. Nothing changes from the user's perspective except context is now scoped to where they're working.

`workspace` still looks up the API server from the account and writes both `active_workspace` and `api_server` to `./.yherda`.

## `.yherda` and `.gitignore`

`.yherda` should be added to `.gitignore` by default — it's ephemeral session state, not something to commit. The CLI can warn if it detects the cwd is a git repo and `.yherda` is not ignored, but this is a nice-to-have.

## Files affected

- `internal/config/config.go` — remove `Config` struct, `SaveConfig`, `LoadConfig`; keep credentials functions
- `internal/config/context.go` — new file: `Context` struct, `LoadContext()`, `SaveContext()`
- `cmd/root.go` — `mustClient()` and `printContext()` call `LoadContext()`; `useParent()` calls `SaveContext()`
- `cmd/workspace.go` — writes to `SaveContext()` instead of `SaveConfig()`
- `cmd/ideas.go`, `cmd/person.go`, `cmd/place.go`, `cmd/thing.go`, `cmd/arc.go` — `use` subcommands call `SaveContext()`
- All tests: write `.yherda` in a temp dir instead of `config.json`

## Unit Test Plan

### `internal/config/context_test.go` (new)

- `TestLoadContext_FileNotExist` — no `.yherda` file, assert empty context returned, no error
- `TestSaveAndLoadContext` — save a full context, load it back, assert all fields match
- `TestLoadContext_PartialFile` — file with only `active_workspace` set, assert other fields empty
- `TestSaveContext_CascadeReset` — save idea + downstream fields, re-save with idea changed and downstream cleared, assert correct

All tests use `t.Chdir()` or write to a temp dir to avoid touching the real cwd.

### `internal/config/config_test.go` (update)

Remove all `Config`/`SaveConfig`/`LoadConfig` tests. Keep credential tests unchanged.

### `cmd/context_commands_test.go` (update)

Replace `saveContextWithCreds(t, &config.Config{...})` setup with writing a `.yherda` file in a temp cwd.

### `cmd/workspace_test.go` (new)

- `TestWorkspaceUse_WritesContextFile` — assert `.yherda` contains correct workspace and api_server after `yherda workspace <slug>`
- `TestWorkspaceShow_NoFile` — assert "no active workspace" when `.yherda` absent
- `TestWorkspaceShow_WithFile` — assert workspace name and endpoint printed when `.yherda` present
