**YOS-87** — removes the CLI write commands made permanently non-functional by GEN-627 + YOS-86's persona-scope cap. Not a bug fix: Shawn, explicit, 2026-08-14 — "this is expected by me. I don't want a cli access token to be broadly functional." The CLI is deliberately narrowed to Collaborator + read-only Idea Owner scope; write commands outside that boundary are removed rather than left to silently 403. A future, separate `yherda studio *` command family (not yet ticketed) will be the deliberately narrow replacement, built against `/forminstance/` and additional Collaborator-scoped endpoints TBD.

Design doc: [YOS-87: Remove CLI Write Commands Broken by Persona Scope Cap](https://momentsbyshawn.atlassian.net/wiki/spaces/SD/pages/13795329/YOS-87+Remove+CLI+Write+Commands+Broken+by+Persona+Scope+Cap) (Confluence, source of truth; this file is the point-in-time snapshot for this branch).

## Exact removal set (verified 2026-08-14 against current code, two independent static-analysis passes)

17 distinct command functions, 26 cobra command registrations (some share an underlying helper — see below), all confirmed as real `client.Post`/`client.Patch`/`client.Delete` calls against non-`/forminstance/` paths:

- `cmd/ideas.go`: `ideasCreateCmd` (POST `/idea/`)
- `cmd/person.go`: `personCreateCmd` (POST `/idea/{id}/persons/`)
- `cmd/place.go`: `placeCreateCmd` (POST `/idea/{id}/places/`)
- `cmd/thing.go`: `thingCreateCmd` (POST `/idea/{id}/things/`)
- `cmd/setting.go`: `settingCreateCmd` (POST `/place/{id}/settings/`) — `settingListCmd` survives, file stays
- `cmd/identity.go`: `identityCreateCmd` (POST `/person/{id}/identities/`)
- `cmd/context.go`: `contextCreateCmd` (POST `/idea/{id}/contexts/`), `contextBeliefAddCmd` (POST `/context/{id}/beliefs/`), `contextBeliefRemoveCmd` (DELETE `/context/{id}/beliefs/`) — `contextListCmd`/`contextUseCmd`/`contextBeliefListCmd` survive
- `cmd/model.go`: `modelDispositionsCreateCmd`/`modelDispositionsDeleteCmd` (POST/DELETE `/subject/{id}/dispositions/`), `modelStatesCreateCmd`/`modelStatesDeleteCmd` (POST/DELETE `/subject/{id}/states/`), `modelStatesDispositionsSetCmd`/`modelStatesDispositionsUnsetCmd` (POST/DELETE `/state/{id}/dispositions/`) — `modelShowCmd`/`modelListCmd`/`modelUseCmd`/`modelBackCmd`/list variants survive
- `cmd/model_add.go`: `modelAddPerspectiveCmd` (POST `/subject/{id}/perspective/`), `modelAddGoalCmd` (POST `/subject/{id}/goals/`, shares `createGoalOnSubject` helper with `modelGoalsCreateCmd` below)
- `cmd/model_belief.go`: `modelBeliefCreateCmd` (POST `/idea/{id}/beliefs/`), `modelBeliefContextsAddCmd`/`modelBeliefContextsUpdateCmd`/`modelBeliefContextsRemoveCmd` (POST/PATCH/DELETE `/context/{id}/beliefs/`)
- `cmd/model_goal.go`: `modelGoalsCreateCmd` (POST `/subject/{id}/goals/`, shares `createGoalOnSubject` with `modelAddGoalCmd`), `modelStepsCreateCmd` (POST `/goal/{id}/steps/`) — `modelGoalsListCmd`/`modelGoalUseCmd`/`modelStepsListCmd` survive
- `cmd/model_perspective.go`: `modelPerspectiveGetCmd` (POST `/subject/{id}/perspective/` — **named "get" but is a real write, get-or-create semantics; must not be skipped as read-only by name**), `modelPerspectiveContextsAddCmd`/`modelPerspectiveContextsUpdateCmd`/`modelPerspectiveContextsRemoveCmd` (POST/PATCH/DELETE `/perspective/{id}/contexts/`) — `modelPerspectiveContextsListCmd` survives

`createGoalOnSubject` (cmd/model_goal.go) is a shared helper called by both `modelAddGoalCmd` (model_add.go) and `modelGoalsCreateCmd` (model_goal.go) — two independent cobra registrations, same underlying write. Delete the helper once; remove both command registrations.

## Explicitly NOT removed — local-only, no server call

`ideasUseCmd`, `personUseCmd`, `placeUseCmd`, `thingUseCmd`, `contextUseCmd`, `modelUseCmd`, `modelBackCmd`, `modelStatesUseCmd`, `modelGoalUseCmd` — all write only to the local `.yherda` active-context file, no HTTP call. Shawn, explicit: "use should stay."

## Dormant/unregistered — out of scope for this ticket

`doc create`/`doc update` (cmd/docs.go) and `ideas import` (cmd/ideas_import.go) both perform non-forminstance writes but are already unregistered (`docsCmd`/`ideasImportCmd` commented out of `root.go`, per YOS-80 — "being replaced"). Not touched here; YOS-80's own disposition governs them.

## Passing test

`go build ./...`, `go vet ./...`, `gofmt -l .` all clean. `go test ./...` passes with no reference to a removed command remaining. `yherda --help` (and each parent command's `--help`) no longer lists any of the 17 removed subcommands; `ideasUseCmd`-style local commands remain listed and functional.
