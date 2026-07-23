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
| `cmd/ideas_export.go` | `ideas export` — structural export of an idea's entity graph. **Disabled** — being replaced, see YOS-80. |
| `cmd/ideas_import.go` | `ideas import` — structural import of a source format into an idea's entity graph. **Disabled** — being replaced, see YOS-80. |
| `cmd/identity.go` | Identity (character) management. |
| `cmd/person.go`, `cmd/place.go`, `cmd/setting.go`, `cmd/thing.go`, `cmd/disposition.go` | Person/place/thing entities and their settings and dispositions. |
| `cmd/model.go` | The `model` command family — Subject-generic commands (`show`, `dispositions`, `states`, `states dispositions`) that operate on the platform's Subject base class by bare id, working the same way regardless of concrete subtype. Distinct from the per-type resource files above. |
| `cmd/model_perspective.go` | `model perspective` — get/materialize a Subject's Perspective and manage the Contexts attached to it (`ContextPerspective`). |
| `cmd/model_belief.go` | `model belief` — create a Belief (idea-scoped root create) and manage its attachment to Contexts (`BeliefContext`). |
| `cmd/model_goal.go` | `model goals`, `model goal use`, `model steps` — Goal/Step CRUD on a Subject, plus the active-goal context slot. |
| `cmd/model_add.go` | `model add perspective`/`model add goal` — the first commands that consume `ctx.Subject` (set by `model use`): add a capability to the active Subject, with an optional positional id override. |
| `cmd/projects.go` | Project listing (`/api/ideaproject/`). |
| `cmd/format.go` | Expression format management. **Disabled** — being replaced, see YOS-80. |
| `cmd/expression.go` | Expression list/show/print/export — the segment-tree-to-output pipeline. **Disabled** — being replaced, see YOS-80. |
| `cmd/docs.go` | Idea *documents* (freeform notes attached to an idea) — unrelated to this README despite the name. **Disabled** — being replaced, see YOS-80. |
| `internal/api/client.go` | Thin HTTP client (`Get`/`Post`/`Patch`) with bearer-token auth and one automatic refresh-and-retry on 401. |
| `internal/auth/` | OAuth2 PKCE login flow (`pkce.go`) and the local browser launch (`browser.go`). |
| `internal/config/` | Local persistence: `config.go` for credentials (`~/.yherdacmd/credentials.json`), `context.go` for per-directory active context (`.yherda` file). |
| `internal/export/` | **Content export pipeline.** `exporter.go` defines the `Exporter` interface, the format registry, and `BuildTree`, which converts the API's nested segment JSON into the shared `SegmentNode` tree and reads each segment's content out of its `data` array (currently hardcoded to the `writer` plugin's entry — see `defaultPlugin`). `stdout.go` and `scriv.go` are the two registered formats today. |
| `internal/structural/` | **Structural export and import pipeline.** `manifest.go` declares which entities a format needs. `resolver.go` fetches the entity graph from the API. `exporter.go` defines the `Exporter` interface and export format registry; `obsidian.go` is the first registered export format. `importer.go` defines the `Importer` interface and import format registry. `scriv_importer.go` parses Scrivener 3 `.scriv` packages into an `IdeaGraph` (manuscript/scene structure is preserved as unmapped idea documents — Arc/Beat was removed in GEN-555 pending a future structural rebuild). `marshaller.go` is the reverse resolver — POSTs an `IdeaGraph` to the API in dependency order (identities → places → things → docs). |

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
- **Active context** — a `.yherda` file in the current working directory, holding the active workspace, idea, person, place, and thing. Most resource commands fall back to this context when an explicit ID flag isn't passed, and `useParent` (in `cmd/root.go`) persists newly-created or `use`d resources back into it.

## Build and test

```sh
go build ./...
go test ./...
```

No build tags or external services are required to run the test suite — tests exercise the cobra command tree and local config/auth logic directly.

## Export and import pipelines

There are two independent export pipelines and one import pipeline:

| Pipeline | Package | Command | What it handles |
|---|---|---|---|
| Content export | `internal/export/` | `yherda expression export` | Segment tree → rendered format (manuscript, screenplay, …) |
| Structural export | `internal/structural/` | `yherda ideas export` | Entity graph → structural format (Obsidian vault, …) |
| Structural import | `internal/structural/` | `yherda ideas import` | Source format → entity graph → API |

The two structural pipelines share `IdeaGraph` as their interchange format. Export: `resolver.go` fetches from the API → `IdeaGraph` → format driver writes files. Import: format driver reads files → `IdeaGraph` → `marshaller.go` POSTs to the API. Import → export gives format conversion for free (e.g. Scrivener → Obsidian via the Yherda model).

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

### Adding a new structural import format (`internal/structural/`)

1. Implement `structural.Importer`:
   - `Import(source string) (IdeaGraph, error)` — parse the source path and return a populated `IdeaGraph`.
   - Use `_id`, `_identity_id`, and `_arc_id` fields on graph entities to express dependency relationships; the marshaller resolves these to platform IDs at write time.
   - Mark unmapped items with `"_unmapped": "true"` so they appear in `--dry-run` output separately.
2. Register it in the `importerRegistry` map in `importer.go`.

The `--format` flag help text in `ideas import` is auto-populated from `structural.ImportFormats()` at init time.

## Adding a new resource command

Existing resource files (`cmd/person.go`, `cmd/place.go`, `cmd/thing.go`, etc.) all follow the same shape:

1. A parent `cobra.Command` for the resource (e.g. `personCmd`).
2. Subcommands for `list`, `show <id>`, `create`, and `use <id>` (not every resource needs all four — pick what fits).
3. Each `RunE` builds an authenticated client with `mustClient()`, falls back to the active context for parent IDs when no flag is passed, and calls `client.Get`/`Post`/`Patch` against the relevant REST path.
4. Output: if `jsonOutput` is set, call `printJSON(result)` and return; otherwise format with `newTabWriter()` and finish with `printContext()` so the user sees what context the command ran against.
5. Register the new top-level command in `cmd/root.go`'s `init()`.

Following this pattern keeps every resource consistent and means a new contributor (or an agent) can infer how an unfamiliar command works from any other one.

Subject-generic commands (`cmd/model.go`) are the exception: they operate on a bare Subject id rather than a specific subtype, so they don't get their own `use`-and-cascade context slot unless a specific need is demonstrated — `model show`/`dispositions`/`states` all take an explicit id argument rather than falling back to an existing Person/Place/Thing context. `model states use`, `model goal use`, and `model use` are the exceptions. The first two exist because their active-context values are meaningfully reused across `states dispositions` and `steps` subcommands respectively; `model use` sets `ctx.Subject`, a fully generic active-Subject slot, whose first and so-far-only consumer is `model add` (`cmd/model_add.go`) — `perspective`/`goal` both resolve their target from `ctx.Subject` by default, with an optional positional id override, same fallback shape as `model steps`/`model states dispositions`. None of the three clear Person/Place/Thing/State/Goal — a State, Goal, or generic Subject is scoped to whichever Subject already has it, not a replacement for the active Person/Place/Thing, so clearing them on `use` would lose context unnecessarily.

Some commands confirm before a side effect bigger than the command name suggests — `model goals create` prompts before creating a Goal on a Subject with no existing Self, since that cascades a default Identity and its own Perspective/Disposition into existence too. Pass `--yes`/`-y` to skip the prompt (e.g. from a script). See insight_cli_capability_grant_confirmation_pattern.md for when this pattern should be reached for elsewhere in the CLI: case-by-case, only where a command's blast radius isn't obvious from its name.

### Belief-attachment specificity

`model belief create` and `model perspective contexts add` don't enforce this — it's a convention for whoever is typing the command: an unqualified belief ("the king is not to be trusted") generally attaches at a Subject's own (or base) Perspective, while a fully-qualified one ("X, in identity Y, when Z") attaches at that specific Identity/Disposition's own Perspective instead.

`model dispositions contexts` (Disposition-Context attachment, reusing the same Perspective-Contexts mechanism since Disposition is itself a Subject) is not implemented yet — blocked on a backend gap (no Subject-generic `/api/subject/{id}/contexts/` endpoint exists today, tracked as GEN-574). Don't mistake its absence for an oversight.

### Idea ids vs Subject ids

These are separate id spaces on the backend and can collide numerically (Idea `5` and Subject `5` are unrelated records that happen to share a number) — see YOS-81. `model list` normally lists Subjects for an idea (`ID` column = Subject ids), but when it falls back to listing all Ideas (no active idea, no `<idea-id>` arg), the fallback table is deliberately shaped differently — header `IDEA ID` instead of `ID`, plus an explicit warning line — so it's obvious those ids don't work with `model show` or any other `model` command.

Every `model` command that acts on a bare Subject id also prints a one-line confirmation of what that id actually resolves to (`Subject: #<id> "<name>" (<subject_type>)`) before acting, via the shared `printSubjectContext` helper in `cmd/root.go` — skipped in `--json` mode, since a JSON caller already knows what it asked for. This is visibility, not validation: no `model` command hard-blocks any Subject id or subtype, per the epic's no-type-gating principle — the goal is that a wrong-context id (e.g. one copied from `model list`'s idea-fallback table) is immediately visible, not silently accepted.

### Why there's no `model add identity`

Two reasons, not one. First, a backend gap: Identity creation is hardcoded to `Person` server-side (`POST /api/person/{id}/identities/`), with no Subject-generic path — tracked as GEN-576. Second, and more durable even if GEN-576 ships: Perspective is the generic "this Subject can hold context/opinion" capability, cheap and universal on any Subject. Identity is specifically the character layer, and it only becomes load-bearing once a Subject needs a Goal — which already cascades Self → Identity → Perspective on its own via `model add goal`/`model goals create`. Nobody adds Identity to a chair that just needs an opinion; they give it a Perspective. A standalone "give this Subject an Identity" command may not be a real need at all — don't assume its absence is an oversight to fix.

## Releasing

Push a version tag to trigger goreleaser:

```
git tag v0.1.0
git push origin v0.1.0
```

goreleaser builds binaries for all platforms, creates a GitHub Release, and updates the Homebrew tap formula automatically.
