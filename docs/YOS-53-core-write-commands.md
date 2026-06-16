# YOS-53: Core Write Commands

## Overview

Six `create` subcommands that let users and agents construct a complete Yherda model from the command line. All commands output JSON and are composable — the `id` from one command can be piped into the next.

## Commands

```
yherda identity create --person <person-id> --name <name>
yherda arc create --person <person-id> --want <want>
yherda beat create --arc <arc-id> --description <description>
yherda place create --idea <idea-id> --name <name>
yherda setting create --place <place-id> --name <name>
yherda thing create --idea <idea-id> --name <name>
```

Context fallback: each `--person`, `--arc`, `--idea`, or `--place` flag falls back to the active context (`person use`, `arc use`, `ideas use`, `place use`) when omitted.

## Agent Pipeline Example

Build a complete idea structure in sequence:

```sh
# Create the idea
IDEA=$(yherda ideas create --name "The Journey" | jq -r '.id')

# Add a person
PERSON=$(yherda person create --idea $IDEA --name "Hero" | jq -r '.id')

# Add an identity for that person
yherda identity create --person $PERSON --name "The Chosen One"

# Add an arc
ARC=$(yherda arc create --person $PERSON --want "To find the lost artifact" | jq -r '.id')

# Add beats to the arc
yherda beat create --arc $ARC --description "The call to adventure"
yherda beat create --arc $ARC --description "The mentor appears"

# Add a place and its setting
PLACE=$(yherda place create --idea $IDEA --name "The Ancient Forest" | jq -r '.id')
yherda setting create --place $PLACE --name "The Hidden Temple"

# Add a thing
yherda thing create --idea $IDEA --name "The Lost Artifact"
```

## API Endpoints

| Command | Method | Endpoint | Body |
|---|---|---|---|
| `identity create` | POST | `/api/role/{id}/identities/` | `{name}` |
| `arc create` | POST | `/api/role/{id}/arcs/` | `{want}` |
| `beat create` | POST | `/api/arc/{id}/beats/` | `{description}` |
| `place create` | POST | `/api/storyline/{id}/places/` | `{name}` |
| `setting create` | POST | `/api/place/{id}/settings/` | `{name}` |
| `thing create` | POST | `/api/storyline/{id}/things/` | `{name}` |

## Error Handling

- Missing required flag → error on stderr, non-zero exit before any API call
- No context and no flag → clear error on stderr, non-zero exit
- API 400/404 → error body on stderr, non-zero exit
- All API errors surface as non-zero exit codes; stdout is reserved for JSON of the created resource
