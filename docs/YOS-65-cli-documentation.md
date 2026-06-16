# YOS-65: Full Documentation for yherda-cmd

## Scope

Three deliverables, three audiences:

1. **Developer README** (`yherda-cmd/README.md`) — source layout, dependencies, env vars, build/test instructions.
2. **Customer-facing docs** (`public-docs` repo → doc.yherda.com) — how authors use the CLI, written for writers integrating Yherda into their toolchain, not for AI agents.
3. **Inline CLI help** (cobra `Short`/`Long`/`Example` on every command) — `yherda help` / `yherda <cmd> --help` should be comprehensive enough that a human or an agent can discover the entire surface without leaving the terminal.

Both repos share this ticket: `yherda-cmd` and `public-docs` each get a branch/PR named `YOS-65-cli-documentation`.

## 1. Developer README

`README.md` at repo root: source layout table (`cmd/*.go` and `internal/{api,auth,config,export}`), dependencies, environment variables (`YHERDA_PUBLIC_HOST`, `YHERDA_INSECURE`, `DEVELOPER`), local config/credentials/context file locations, build/test commands, and a short pattern guide for adding a new resource command so new contributors match existing conventions.

## 2. Customer-facing docs (doc.yherda.com)

New `public-docs/cli/` directory, mirroring the existing `concepts/` pattern:

- `getting-started.md` — install, login, set active idea
- `concepts.md` — idea/person/identity/arc/beat/place/setting/thing/disposition/expression/format/project mapped to CLI commands
- `workflows.md` — task-oriented: reading drafts, exporting manuscripts, switching projects, building out a character arc, capturing notes, checking project status
- `integrations.md` — scriptable `--json` output, editor/build integration, stdin/file input, non-interactive design. Agent-friendliness mentioned once, at the end, as an incidental side effect — not the lead.

`index.md` updated with a new **CLI** nav section.

Out of scope: full per-command flag reference — that's what `--help` is for; the docs site points to it rather than duplicating it.

## 3. Inline CLI help

Audited every `cobra.Command` across all 14 `cmd/*.go` files. Every command/subcommand now has `Short` (already mostly present), every parent and most leaf commands have `Long`, and every leaf command (`list`, `show`, `create`, `use`, `export`, `print`) has at least one `Example`. Root command's `Long` reframed to lead with "integrates with your existing tools and scripts," with agent-friendliness mentioned as a single clause, not the headline.

## Non-goals

- No man pages / shell completion generation (cobra's `completion` subcommand is free, not part of this ticket).
- No rewrite of `cmd/docs.go` (idea-document management — unrelated naming collision with "documentation"; its help text now calls this out explicitly so it isn't confused with this work).
- No changes to command behavior — text/doc only.

## Test plan

- `go build ./...` and `go test ./...` — green, no behavior change (confirmed).
- Manual: `yherda help` and several `yherda <cmd> --help` invocations reviewed for completeness and tone.
- `public-docs`: verified every relative markdown link in the new `cli/*.md` pages and the `index.md` update resolves to a real file.

## Docs update plan

- `yherda-cmd/README.md` — new file (this ticket's developer-facing artifact).
- `public-docs/index.md` — new CLI nav section.
- `public-docs/cli/*.md` — new directory: getting-started, concepts, workflows, integrations.
