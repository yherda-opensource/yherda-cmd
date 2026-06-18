# yherda-cmd

A command line interface for [Yherda](https://doc.yherda.com) — manage ideas,
identities, arcs, beats, places, things, and expressions from the terminal,
and export your work into the formats your other tools expect.

For end-user documentation (installing, getting started, common workflows),
see [doc.yherda.com](https://doc.yherda.com). This README is for people
working on the CLI itself.

## Installation

### macOS (Homebrew)

```
brew tap yherda-opensource/yherda
brew install yherda
```

### Linux / Windows

Download the binary for your platform from the
[GitHub Releases](https://github.com/yherda-opensource/yherda-cmd/releases) page,
extract the archive, and add the binary to your `PATH`.

## Source layout

| Path | Responsibility |
|---|---|
| `main.go` | Entry point; delegates straight to `cmd.Execute()`. |
| `cmd/root.go` | Root cobra command, global flags (`--json`, `--no-context`), shared helpers (`mustClient`, `printContext`, `printJSON`, `newTabWriter`). |
| `cmd/auth.go` | `login` / `logout`. |
| `cmd/workspace.go`, `cmd/workspacelist.go` | Active workspace selection and listing. |
| `cmd/ideas.go` | Idea CRUD + setting the active idea. |
| `cmd/ideas_export.go` | `ideas export` — structural export of an idea's entity graph. |
| `cmd/identity.go` | Identity (character) management. |
| `cmd/arc.go`, `cmd/beat.go` | Arc and beat management. |
| `cmd/person.go`, `cmd/place.go`, `cmd/setting.go`, `cmd/thing.go`, `cmd/disposition.go` | Person/place/thing entities and their settings and dispositions. |
| `cmd/projects.go` | Project listing (`/api/ideaproject/`). |
| `cmd/format.go` | Expression format management. |
| `cmd/expression.go` | Expression list/show/print/export — the segment-tree-to-output pipeline. |
| `cmd/docs.go` | Idea *documents* (freeform notes attached to an idea) — unrelated to this README despite the name. |
| `internal/api/client.go` | Thin HTTP client (`Get`/`Post`/`Patch`) with bearer-token auth and one automatic refresh-and-retry on 401. |
| `internal/auth/` | OAuth2 PKCE login flow (`pkce.go`) and the local browser launch (`browser.go`). |
| `internal/config/` | Local persistence: `config.go` for credentials (`~/.yherdacmd/credentials.json`), `context.go` for per-directory active context (`.yherda` file). |
| `internal/export/` | **Content export pipeline.** `exporter.go` defines the `Exporter` interface, the format registry, and `BuildTree`, which converts the API's nested segment JSON into the shared `SegmentNode` tree and reads each segment's content out of its `data` array (currently hardcoded to the `writer` plugin's entry — see `defaultPlugin`). `stdout.go` and `scriv.go` are the two registered formats today. |
| `internal/structural/` | **Structural export pipeline.** Exports an idea's entity graph (identities, arcs, beats, places, things, and attached documents) into formats that understand story structure. `manifest.go` declares which entities a format needs and enforces dependency rules (Beats → Arcs → Identities). `resolver.go` fetches only what the manifest requests. `exporter.go` defines the `Exporter` interface and format registry. `obsidian.go` is the first registered format, writing a vault of `.md` files. |

## Dependencies

- [`spf13/cobra`](https://github.com/spf13/cobra) — command/flag framework. Everything in `cmd/` is built on it.
- [`spf13/viper`](https://github.com/spf13/viper) and its supporting libraries (`afero`, `pelletier/go-toml`, `fsnotify`, etc.) — pulled in transitively by cobra; not used directly today.
- No third-party HTTP client — `internal/api/client.go` uses `net/http` directly.

Run `go mod tidy` after adding any new import.

## Environment variables

| Variable | Effect |
|---|---|
| `YHERDA_PUBLIC_HOST` | Overrides the public (no-workspace) API host used during login (`internal/auth/pkce.go`) and by `mustPublicClient()`. Defaults to `https://public.a.yherda.com`. |
| `YHERDA_INSECURE=1` | Skips TLS certificate verification on the HTTP client. Development only — never set this against a real deployment. |
| `DEVELOPER=1` | Prints every outgoing request's method and URL to stderr (`[dev] GET ...`), useful when debugging which endpoint a command actually calls. |

## Local state

- **Credentials** — `~/.yherdacmd/credentials.json` (mode 0600), written by `login` and refreshed automatically by the API client on a 401.
- **Active context** — a `.yherda` file in the current working directory, holding the active workspace, idea, person, arc, place, and thing. Most resource commands fall back to this context when an explicit ID flag isn't passed, and `useParent` (in `cmd/root.go`) persists newly-created or `use`d resources back into it.

## Build and test

```sh
go build ./...
go test ./...
```

No build tags or external services are required to run the test suite — tests exercise the cobra command tree and local config/auth logic directly.

## Export pipelines

There are two independent export pipelines:

| Pipeline | Package | Command | What it exports |
|---|---|---|---|
| Content | `internal/export/` | `yherda expression export` | Segment tree → rendered format (manuscript, screenplay, …) |
| Structural | `internal/structural/` | `yherda ideas export` | Entity graph → structural format (Obsidian vault, …) |

A format that needs both (e.g. a film production package wanting a scene breakdown _and_ scene content) implements both interfaces and registers in both packages. The pipelines stay independent — they operate on different object graphs and will diverge.

### Adding a new content export format (`internal/export/`)

1. Implement `export.Exporter` (`Export(title string, roots []SegmentNode, output string) error` + `DefaultExt() string`).
2. Register it in the `registry` map in `exporter.go`.

### Adding a new structural export format (`internal/structural/`)

1. Implement `structural.Exporter`:
   - `Manifest() Manifest` — declare which entity types you need. Dependency rules (Beats→Arcs→Identities) are enforced by `Manifest.Resolve()` automatically.
   - `Export(graph IdeaGraph, output string) error` — write the output.
   - `DefaultExt() string` — file extension, or `""` for directory-based formats.
2. Register it in the `registry` map in `exporter.go`.

The `--format` flag help text in `ideas export` is auto-populated from `structural.Formats()` at init time — no other changes needed.

## Adding a new resource command

Existing resource files (`cmd/person.go`, `cmd/place.go`, `cmd/thing.go`, etc.) all follow the same shape:

1. A parent `cobra.Command` for the resource (e.g. `personCmd`).
2. Subcommands for `list`, `show <id>`, `create`, and `use <id>` (not every resource needs all four — pick what fits).
3. Each `RunE` builds an authenticated client with `mustClient()`, falls back to the active context for parent IDs when no flag is passed, and calls `client.Get`/`Post`/`Patch` against the relevant REST path.
4. Output: if `jsonOutput` is set, call `printJSON(result)` and return; otherwise format with `newTabWriter()` and finish with `printContext()` so the user sees what context the command ran against.
5. Register the new top-level command in `cmd/root.go`'s `init()`.

Following this pattern keeps every resource consistent and means a new contributor (or an agent) can infer how an unfamiliar command works from any other one.

## Releasing

Push a version tag to trigger goreleaser:

```
git tag v0.1.0
git push origin v0.1.0
```

goreleaser builds binaries for all platforms, creates a GitHub Release, and updates the Homebrew tap formula automatically.
