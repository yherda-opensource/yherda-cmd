# Design Doc — YOS-57: Structural idea export (`yherda ideas export`)

> **Revised 2026-06-18 (v3)** — added developer extension guidance (content vs structural vs both); added plugin data forward-compatibility note.

---

## Summary

Add `yherda ideas export --format obsidian [--idea <id>] [--output <path>]` to the CLI.

This is a **structural export** — it exports the entity graph of an idea (identities, arcs, beats, places, things, and their attached docs) into a format that understands story structure. This is distinct from `yherda expression export`, which is a **content export** of a segment tree into a rendered format (manuscript, screenplay, etc.).

No new backend endpoints required — all entity data is already available via existing API routes.

---

## Two export types in the CLI

| Command | What it exports | Examples |
| --- | --- | --- |
| `yherda expression export` | Segment tree → rendered content | scriv, stdout |
| `yherda ideas export` | Entity graph → structural format | obsidian, sbinder (future) |

A format that needs both (e.g. a film production package wanting a breakdown sheet _and_ scene content) implements both interfaces and registers in both packages. The CLI can wire them together at the command level. The pipelines stay independent.

---

## Architecture

### `internal/structural/` — new package

`resolver.go` — fetches the entity graph for an idea, driven by the format's declared requirements.

```go
type IdeaGraph struct {
    Idea        map[string]any
    Identities  []map[string]any
    Arcs        []map[string]any   // always paired with Identities
    Beats       []map[string]any   // always paired with Arcs
    Places      []map[string]any
    Things      []map[string]any
    Docs        []map[string]any   // attached to their entity via entity_type/entity_id
    PluginData  map[string]any     // reserved — see Plugin data section below
}
```

`manifest.go` — declares which entity types a format needs. Dependencies are enforced here, not by the caller: requesting `Beats` implies `Arcs` implies `Identities`. The format driver declares its needs; the resolver satisfies them.

```go
type Manifest struct {
    Identities bool
    Arcs       bool  // forces Identities=true
    Beats      bool  // forces Arcs=true (and therefore Identities=true)
    Places     bool
    Things     bool
    Docs       bool
    Plugins    []string  // reserved — plugin keys to include; ignored until plugin registry exists
}
```

`exporter.go` — interface all structural format drivers implement:

```go
type Exporter interface {
    Manifest() Manifest
    Export(idea IdeaGraph, output string) error
    DefaultExt() string
}
```

`Formats()` returns a sorted list of registered driver names — embedded in the `--format` flag description at init time so `--help` always reflects what's available.

### `internal/structural/obsidian.go` — first format driver

Manifest: all entity types + docs.

Output layout:

```
obsidian-export/
  identities/
    jo-ann-hayes.md       # front matter: id, arc ids; body from attached doc if present
  arcs/
    the-reluctant-hero.md
  beats/
    act-1-beat-1.md       # front matter: arc_id, identity_id, moment
  places/
    the-warehouse.md
  things/
    the-briefcase.md
  docs/                   # freeform docs not attached to any entity
    worldbuilding-notes.md
```

Each file: YAML front matter (entity id, relationships, created) + body from attached IdeaDocument if one exists, otherwise empty body.

Collision handling: append `-2`, `-3` on duplicate filenames.
Error on existing files in output dir — no `--force` in v1.

### `cmd/ideas_export.go`

Adds `ideasExportCmd` under the existing `ideasCmd`:

```
yherda ideas export --format obsidian [--idea <id>] [--output <path>]
```

* `--format` required; description auto-populated from `structural.Formats()` at init time
* `--idea` overrides active context
* `--output` defaults to `./obsidian-export/`

Flow:

1. Resolve idea ID from flag or context
2. Look up driver: `structural.Get(format)`
3. `resolver.Resolve(client, ideaID, driver.Manifest())` → `IdeaGraph`
4. `driver.Export(graph, outputPath)`

---

## API calls the resolver makes (driven by manifest)

| Entity | Endpoint |
| --- | --- |
| Idea | `GET /storyline/{id}/` |
| Identities | `GET /storyline/{id}/identities/` |
| Arcs | `GET /storyline/{id}/arcs/` |
| Beats | `GET /storyline/{id}/beats/` |
| Places | `GET /storyline/{id}/places/` |
| Things | `GET /storyline/{id}/things/` |
| Docs | `GET /storyline/{id}/documents/` |

Resolver only calls the endpoints the manifest declares.

---

## Plugin data — forward compatibility

Future plugins (e.g. a Shot Planner module) will contribute data to segments — a `shots: []` array stored in `SegmentData` alongside the existing `writer` plugin data. When a plugin registry exists, the manifest's `Plugins []string` field will drive the resolver to fetch and attach that plugin's data into `IdeaGraph.PluginData`. Format drivers that know about a specific plugin can render it; generic drivers get the raw map.

`IdeaGraph.PluginData` and `Manifest.Plugins` are reserved in this ticket but not populated — the plugin registry doesn't exist yet. The fields are placeholders so the struct shape doesn't need a breaking change when the registry lands.

---

## Unit test plan

`internal/structural/obsidian_test.go`

| Test | What it checks |
| --- | --- |
| `TestObsidian_CreatesDirectoryLayout` | Correct subdirectories created for declared entity types |
| `TestObsidian_EntityFile` | Entity file has correct front matter + body from attached doc |
| `TestObsidian_EntityNoDoc` | Entity with no attached doc → front matter only, empty body |
| `TestObsidian_TitleSanitization` | Entity names with spaces/punctuation → clean filenames |
| `TestObsidian_TitleCollision` | Two entities same name → `name.md` and `name-2.md` |
| `TestObsidian_OrphanDocs` | Docs not attached to any entity → written to `docs/` subdirectory |
| `TestObsidian_OutputDirExistsWithFiles` | Pre-existing files in output dir → error |
| `TestObsidian_OutputDirCreated` | Non-existent output dir → created |

`internal/structural/manifest_test.go`

| Test | What it checks |
| --- | --- |
| `TestManifest_BeatsImpliesArcsImpliesIdentities` | Setting Beats=true forces Arcs=true and Identities=true |
| `TestManifest_ArcsImpliesIdentities` | Setting Arcs=true forces Identities=true |
| `TestManifest_IndependentEntities` | Places and Things are independent of identity chain |

`internal/structural/resolver_test.go` — mock API client; verifies correct endpoints called for a given manifest.

All tests use `t.TempDir()` for any file I/O.

---

## Docs update plan

* `README.md`: add `ideas export` to command reference; note distinction from `expression export`; document the two-pipeline model for format developers
* No backend changes

---

## Key decisions

1. `yherda ideas export`, not `yherda doc export` — structural, not a doc dump. Docs are passengers on the entity graph.
2. **Manifest enforces dependency rules** — beats without arcs is incoherent; impossible by design, not runtime validation.
3. **Two separate packages** — `internal/export/` for content (segment tree); `internal/structural/` for structure (entity graph). A format needing both implements both interfaces. Keeping them separate is intentional — they operate on different object graphs and will diverge.
4. **Format driver owns its manifest** — adding a new format is: implement `Exporter`, declare manifest, register. No other changes.
5. **Plugin fields reserved now** — `IdeaGraph.PluginData` and `Manifest.Plugins` exist as placeholders so the plugin registry can be wired in without a breaking struct change.
6. **Error on existing output files, no `--force` in v1.**
