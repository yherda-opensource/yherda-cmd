# YOS-56 — `yherda expression export` with pluggable format drivers

## Reframe

Export is expression-driven. The segment tree of an expression already encodes the full structure — hierarchy, content, ordering. Export serializes that tree into a target format. `expression print` is the same operation with stdout as the sink; both share the same tree walk and the same driver interface.

The original idea-centric model (arcs → folders, beats → scenes) is discarded. The expression format template made those structural decisions already.

## Command Shape

```
yherda expression export --format scriv [--expression <id>] [--output <path>]
```

- `--format` required. Values: `scriv` (v1), `stdout` (what `print` currently does). Future: `obsidian`, `pdf`.
- `--expression` optional; falls back to active expression context (set via `expression use <id>`).
- `--output` optional for file formats; defaults to `./<expression-id>.<ext>`. Not used for `stdout`.

`expression print` becomes a thin alias:

```
yherda expression print <id>   →   yherda expression export --format stdout --expression <id>
```

Both share the same code path. `print` is kept as a convenience command but delegates to the export machinery.

## Architecture

New package: `internal/export/`

```
internal/export/
  exporter.go       — Exporter interface, SegmentNode type, registry
  stdout.go         — stdout driver (replaces current expression print walk)
  scriv.go          — Scrivener driver
  exporter_test.go
  stdout_test.go
  scriv_test.go
```

`SegmentNode` is built once in `cmd/expression.go` from the existing `/segment/?template=<id>&root=true` response. Both `print` and `export` use it.

Data fetch (two API calls):
1. `GET /api/expression/{id}/` — get template ID and expression title
2. `GET /api/segment/?template={id}&root=true` — full nested segment tree in one call

## Scrivener Package Structure

```
Title.scriv/
  Title.scrivx           — XML binder manifest
  Files/
    Data/
      <UUID>/
        content.rtf      — one file per binder item
```

Segment → Scrivener mapping:
- `TypeName` becomes the binder item title
- Nesting preserved from segment tree
- `Type` derived at serialization: children-only → `Folder`; has content → `Text`
- Content files: minimal valid RTF
- UUIDs generated with `crypto/rand` at export time (not stored)

## Key Decisions

- `print` shares the driver — `StdoutExporter` replaces the current walk; the Cobra command delegates
- Segment tree already fetched nested — no extra API calls
- `Type` is derived, never declared — calculated from `(hasContent, hasChildren)`
- Output path collision check — error if `--output` already exists
- Registry, not switch — adding a new format is one file + one registry entry
