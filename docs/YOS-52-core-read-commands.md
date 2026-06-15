# YOS-52 — Core read commands: list and show ideas, identities, arcs, beats

## Summary

Core read commands for the Yherda CLI. Most commands were scaffolded in the initial skeleton;
this ticket adds unit tests and confirms all commands are correctly wired.

## Commands

| Command | API path |
|---|---|
| `yherda ideas list` | `GET /storyline/` |
| `yherda ideas show {id}` | `GET /storyline/{id}/` |
| `yherda identities list --idea {id}` | `GET /ideas/{id}/identities/` |
| `yherda identities show {id}` | `GET /identities/{id}/` |
| `yherda arcs list --idea {id}` | `GET /storyline/{id}/arcs/` |
| `yherda arcs show {id}` | `GET /arcs/{id}/` |
| `yherda beats list --arc {id}` | `GET /arcs/{id}/beats/` |
| `yherda beats show {id}` | `GET /beats/{id}/` |

All commands output JSON by default. `--pretty` (root flag) produces indented output.
Errors go to stderr with a non-zero exit code.

## Naming

"storyline" does not exist in the CLI or UI. The canonical term is "idea". The `/storyline/`
API path is an internal backend detail only.

## Unit Test Approach

Mock HTTP server per test. Verify:
- Correct API path is called
- Flag value is interpolated into the URL
- Missing required flags print usage and exit non-zero
- JSON response is printed to stdout
