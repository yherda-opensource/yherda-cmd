# YOS-88: find and read expressions

Design doc (shared with the platform story GEN-721): Confluence page 41517058,
"yherda CLI: list expressions by module or type, browse an outline, read from
any segment".

Child of epic YOS-49.

## What it adds

One enabled command group, `yherda expression`, following the CLI's existing
conventions (`person list`, `person use <id>`, flags that override the active
context):

| Command | What it does |
|---|---|
| `expression list [--module "<name>"] [--type <format>] [--idea <id>]` | The expressions you can see in the workspace. `--module` is the module's exact name (quoted when it has spaces) and is resolved with `GET /api/module/?name=`; `--type` is a format name; both are server-side filters on `GET /api/expression/` (GEN-721). |
| `expression use <id>` | Sets the active expression in `.yherda`, after checking it is readable. Cleared by `ideas use` and by changing workspace. |
| `expression outline [--expression <id>]` | The segment tree: id, type, number, name, indented by depth. |
| `expression read [--expression <id>] [--segment <id>]` | The whole expression top to bottom, or that segment and all its children. |

Every command takes the global `--json`. For `read`, `--json` is the raw stored
content and status per segment (so scripts do not change when a module changes
a layout); the text form renders each segment's declared fields, then its body,
then any other stored keys so nothing is hidden.

## Scope and safety

- Read-only. The only request `read` sends that is not a GET is the platform's
  own get-or-create of the expression's form instance, which is how the web
  todo screen reads content too.
- Owner/occupation scope: what you can already read in the web app. Nothing
  here widens it.
- No shell, no module-supplied commands. (Layout-driven rendering of a
  segment's TUI layout is the next step of this story and runs only commands the
  CLI already has.)

## Not in this change

- `packages export --out <folder>`: a thin composition over `list`, `outline`
  and `read --json`, the first follow-up (it is what gets package versions into
  git for staging).
- The legacy `cmd/expression.go` (list/show/print/export) stays disabled and
  untouched: it belongs to YOS-80.
