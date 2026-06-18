# Design Doc — YOS-58: Scrivener Import

## Summary

Add `yherda ideas import` as the inverse of `yherda ideas export`. The import pipeline follows the same MVC shape as the export pipeline (established in YOS-57) but in reverse: a format driver parses source files into an `IdeaGraph`, then a marshaller POSTs the graph to the API in dependency order.

---

## Architecture

### Pipeline symmetry

```
Export:  API → resolver → IdeaGraph → driver → files
Import:  files → driver → IdeaGraph → marshaller → API
```

### New files

| File | Role |
|---|---|
| `cmd/ideas_import.go` | CLI controller — parses `--format`, `--source`, `--idea`, `--dry-run`; picks driver; calls `structural.Import()` |
| `internal/structural/importer.go` | `Import()` entry point — sequences driver → marshaller; handles dry-run gate |
| `internal/structural/marshaller.go` | Reverse resolver — POSTs `IdeaGraph` to API in dependency order (identities → arcs → beats → places → things → docs) |
| `internal/structural/scriv_importer.go` | Scrivener format driver — parses `.scriv` package into `IdeaGraph` |
| `internal/structural/scriv_importer_test.go` | Unit tests for the Scrivener driver (file-in → IdeaGraph assertions) |
| `internal/structural/marshaller_test.go` | Unit tests for the marshaller (IdeaGraph-in → API call assertions, using a stub `Poster`) |

### Existing files — no structural changes

`IdeaGraph` stays as-is. `Exporter` interface and registry are untouched. The Scrivener import driver is registered in a parallel `importerRegistry` following the same pattern as `registry` in `exporter.go`.

---

## CLI interface

```
yherda ideas import --format scriv --source ./MyProject.scriv --idea 42
yherda ideas import --format scriv --source ./MyProject.scriv --dry-run
```

Flags:
- `--format` (required) — import format; initially `scriv`
- `--source` (required) — path to the source file or directory
- `--idea` — target idea ID (overrides active context); if omitted and no active idea exists, error
- `--dry-run` — print what would be created without writing to the API

---

## Importer interface

```go
// Importer parses a source path and returns an IdeaGraph.
type Importer interface {
    Import(source string) (IdeaGraph, error)
}
```

Registered in `importerRegistry` (map in `importer.go`). A parallel `ImportFormats()` and `GetImporter()` give the CLI the same lookup pattern as `Formats()` / `Get()`.

---

## Marshaller

`marshaller.go` provides `Marshal(client Poster, graph IdeaGraph, ideaID string) error`.

```go
// Poster is the subset of the API client used by the marshaller.
type Poster interface {
    Post(path string, body any, out any) error
}
```

Dependency order for POST calls:
1. Identities → `/storyline/{ideaID}/roles/` + `/role/{roleID}/identities/`
2. Arcs → `/role/{identityRoleID}/arcs/`
3. Beats → `/arc/{arcID}/beats/`
4. Places → `/storyline/{ideaID}/places/`
5. Things → `/storyline/{ideaID}/things/`
6. Docs → `/storyline/{ideaID}/documents/` (or entity-attached doc endpoint)

The marshaller maintains a local ID map (`scrivID → platformID`) as it POSTs, so beats and arcs can reference the correct parent IDs that the API assigned.

---

## Scrivener format driver

Scrivener 3 `.scriv` packages contain:
- `project.scrivx` — XML binder tree (folder/document hierarchy, UUIDs, titles, types)
- `Files/Data/{UUID}/content.rtf` or `content.xml` — document body per binder item

### Mapping heuristics

| Scrivener binder item | Yherda entity |
|---|---|
| Top-level folder named "Characters" (or containing character sheets) | Identities |
| Top-level folder named "Manuscript" (or "Draft") — sub-folders | Arcs |
| Scene documents under arc folders | Beats |
| Top-level folder named "Research" | Idea documents |
| Anything else / unmapped | Idea documents (preserve, don't drop) |

**User confirmation at import time:** When the mapping is ambiguous, the driver emits a warning in `--dry-run` output. For the MVP, mapping is heuristic-only — no interactive prompt.

### RTF parsing

Minimal RTF-strip approach to extract plain text from `.rtf` files. Import goal is content preservation, not formatting fidelity — plain text body is sufficient for the initial story pass.

---

## Dry-run mode

`structural.Import()` receives a `dryRun bool`. When true:
1. The driver still runs and produces a full `IdeaGraph`
2. The marshaller is skipped; `Import()` prints a summary instead

---

## Test plan

### Unit tests

**`scriv_importer_test.go`**
- Given a minimal hand-crafted `.scriv` directory (embedded test fixture), assert the correct `IdeaGraph` shape
- Character folder → identities, Manuscript folder → arcs + beats, Research folder → docs
- Unmapped items land as docs, not silently dropped

**`marshaller_test.go`**
- Given an `IdeaGraph`, assert POSTs happen in dependency order using a stub `Poster`
- Local ID map correctness (arc references platform identity ID, not scriv UUID)
- Dry-run: no POST calls, correct summary output

### Integration test

Manual: import a real `.scriv` project, verify entity counts match `--dry-run` output, verify round-trip via `yherda ideas export --format obsidian`.

---

## Docs update plan

- `README.md` — add `ideas import` to the command reference section
- No new env vars or config keys

---

## Out of scope (this story)

- Interactive mapping prompt
- Obsidian vault import driver (separate story)
- Formatting fidelity (bold, italic) — plain text only
- Partial failure recovery / resume
