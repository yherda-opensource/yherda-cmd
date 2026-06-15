# YOS-54 — Expression Format and Expression Commands

## Summary

Add `yherda format` and `yherda expression` subcommand groups to the CLI, scoped to read-only
operations plus `expression print` for rendering an expression's segment tree in reading order.

Creates are out of scope — managed via the web UI or future tickets.

## Commands

| Command | API path |
|---|---|
| `yherda format list` | `GET /api/expressionformat/` |
| `yherda format show <id>` | `GET /api/expressionformat/{id}/` |
| `yherda expression list [--idea <id>]` | `GET /api/expression/?idea=<id>` |
| `yherda expression show <id>` | `GET /api/expression/<id>/` |
| `yherda expression print <id>` | `GET /api/expression/<id>/` → `GET /api/segment/?template=<id>&root=true` |

## Backend change

`SegmentView` gained `ListModelMixin` and a `get_queryset` override supporting:
- `?template=<id>` — filters to segments on a given template
- `?root=true` — further filters to root segments only (no parent)

`SegmentOutputSerializer` already returns the full nested `children` tree recursively,
so `expression print` gets the complete tree in a single API call per expression.

## expression print output

Default (plain text), depth-first:

```
[Act 1]
  [Scene 1] Content of scene one...
  [Scene 2] Content of scene two...
[Act 2]
  [Scene 3] Content of scene three...
```

`--json` flag: array of `{type, number, depth, content}`.

## Idea context

`expression list` requires an idea. Uses `--idea <id>` flag or falls back to `cfg.ActiveIdea`.
Falls back to listing ideas if neither is set.
