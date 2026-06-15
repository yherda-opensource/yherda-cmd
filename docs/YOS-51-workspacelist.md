# Design Doc: YOS-51 — Workspace: switch and show active workspace

## Status

Implemented

## Summary

Complete the `workspace` command group. Two of three acceptance criteria are already implemented on `main`. This ticket adds `workspacelist` and confirms the existing commands are wired correctly.

## What's Already Built (on main)

* `yherda workspace <slug>` — sets active workspace in `~/.yherdacmd/config.json` ✅
* `yherda workspace` (no arg) — shows active workspace ✅

Both are implemented in `cmd/workspace.go` and registered in `cmd/root.go`.

## What Needs to Be Built

* `yherda workspacelist` — lists all workspaces available to the authenticated user

## Architecture

### `workspacelist` command

New top-level command (not a subcommand of `workspace` to match the ticket's AC).

**API endpoint:** `GET /tenants/tenant/mine/` on the public base URL (`public.a.yherda.com`)

Returns all workspaces the authenticated user is a member of. No active workspace required — routes to the public URL, not a tenant subdomain.

**Implementation:** `mustPublicClient()` in `root.go` loads credentials (required) but passes an empty workspace to `api.New`, which already routes to `public.{domain}/api` when workspace is empty. No changes to `api.Client` needed.

**Display:** Tabular `SLUG | NAME` columns; raw JSON with `--json`.

## Unit Test Plan

* `TestWorkspaceList_RejectsArgs` in `read_commands_test.go` — confirms the command rejects unexpected arguments
* Public URL routing already tested in `internal/api/client_test.go` (`TestBaseURL_*` cases with empty workspace)

## Docs Update Plan

* `README.md`: add `workspacelist` to the command reference section
