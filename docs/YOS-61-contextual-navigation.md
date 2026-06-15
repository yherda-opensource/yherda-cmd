# YOS-61 — Contextual navigation commands

## Summary

Replace the existing `identities`, `arcs`, and `beats` commands with a new set of contextual
navigation commands: `person`, `identity`, `arc`, `beat`, `place`, `setting`, `thing`,
`disposition`. Each entity group gets `list` and (where applicable) `use` subcommands. `use`
sets a sticky context in config that `list` falls back to when no explicit flag is provided.
Cascade reset rules wipe downstream context on `use`.

## Model hierarchy

```
Idea (Storyline)
├── Person (Role)         — yherda person
│   ├── Identity          — yherda identity (child of person/role)
│   └── Arc               — yherda arc (child of person/role)
│       └── Beat          — yherda beat (child of arc)
├── Place                 — yherda place
│   └── Setting           — yherda setting
└── Thing                 — yherda thing
    └── Disposition       — yherda disposition
```

- **Person = Role** (a character scoped to a storyline)
- **Identity** = a child of Role — a perspective/belief-state a person holds
- **Arc** = a child of Role, not of Idea directly

## API paths

| Command | API path |
|---|---|
| `person list [--idea {id}]` | `GET /api/storyline/{id}/roles/` |
| `identity list [--person {id}]` | `GET /api/role/{id}/identities/` |
| `arc list [--person {id}]` | `GET /api/role/{id}/arcs/` |
| `arc list --idea {id}` | `GET /api/storyline/{id}/arcs/` (aggregate convenience) |
| `beat list [--arc {id}]` | `GET /api/arcs/{id}/beats/` |
| `place list [--idea {id}]` | `GET /api/storyline/{id}/places/` |
| `setting list [--place {id}]` | `GET /api/place/{id}/settings/` |
| `thing list [--idea {id}]` | `GET /api/storyline/{id}/things/` |
| `disposition list [--thing {id}]` | `GET /api/thing/{id}/dispositions/` |

## Context fields (config.json)

```json
{
  "active_workspace": "...",
  "active_idea": "...",
  "active_person": "...",
  "active_arc": "...",
  "active_place": "...",
  "active_thing": "..."
}
```

## Cascade reset rules

| `use` command | Sets | Clears |
|---|---|---|
| `ideas use {id}` | `active_idea` | `active_person`, `active_arc`, `active_place`, `active_thing` |
| `person use {id}` | `active_person` | `active_arc`, `active_place`, `active_thing` |
| `arc use {id}` | `active_arc` + `active_person` (from arc.role) | `active_place`, `active_thing` |
| `place use {id}` | `active_place` | `active_person`, `active_arc`, `active_thing` |
| `thing use {id}` | `active_thing` | `active_person`, `active_arc`, `active_place` |

Note: `active_idea` is not auto-updated on `use` because current serializers do not expose
`storyline_id`. A follow-up ticket should add `storyline_id` to Role, Place, and Thing
serializers to enable full cascade.

## Supersedes

`docs/YOS-52-core-read-commands.md` — the `identities`, `arcs`, and `beats` commands described
there are removed and replaced by this implementation.
