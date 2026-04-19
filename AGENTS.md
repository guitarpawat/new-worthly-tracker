# Worthly Tracker Agent Notes

This file is for future agents working on this repo. Keep it practical. Prefer small, targeted changes.

## Current Reality

- The app is a Wails desktop app.
- Backend is Go.
- Frontend is plain modular JavaScript plus hand-written CSS.
- SQLite uses `modernc.org/sqlite` only. No CGO.
- The old `templ` / Tailwind experiment was removed. Do not assume it still exists.
- `main.go` is the only entrypoint now.

## Product Summary

Worthly Tracker is a personal net worth tracker built around date-based snapshots.

Main user flows already implemented:

- Home snapshot view
- New snapshot
- Edit snapshot
- Delete snapshot
- Asset management
- Asset / asset type reorder
- Progress & Goals

## Core Business Rules

### Snapshots

- Records are snapshot-based, not transaction-based.
- One asset can have at most one record per snapshot date.
- Latest snapshot = highest active snapshot date.
- Previous snapshot = second highest active snapshot date.

### Autofill

- New snapshot autofill is based on the latest snapshot.
- Respect:
  - asset type ordering
  - asset ordering
  - `is_active`
  - `auto_increment`
- `auto_increment` is a numeric amount stored on `assets.auto_increment`.
- New bought price rule:
  - non-cash asset: `previous bought price + auto_increment`
  - cash asset: bought price must stay `0`

### Active / Inactive Rules

- Removing an asset from the latest snapshot should make that asset inactive for future autofill.
- Inactive assets still remain visible in history and edit flows where appropriate.
- If an asset type is inactive, active assets under that type must not appear in new-asset choices.

### Financial Rules

- Profit = `current - bought`
- `% Profit = profit / bought`
- Division by zero must be handled safely
- Negative values are valid
- Cash rows do not use bought-price profit math

### Deletion

- Use soft delete for:
  - asset types
  - assets
  - snapshots
  - record items
  - goals
- Never silently destroy user data

## Architecture Map

### Backend

- `main.go`
  - app entrypoint
- `internal/app/`
  - Wails app wiring and startup
- `internal/service/`
  - business logic
- `internal/adapter/repository/`
  - SQL access
- `internal/validator/`
  - request/input validation
- `db/migrations/`
  - schema changes
- `db/seeds/dev_seed.sql`
  - demo data

### Frontend

- `frontend/app.js`
  - bootstrap only
- `frontend/js/shared.js`
  - shared state, formatting, backend resolution
- `frontend/js/controls.js`
  - custom select + date controls used across the app
- `frontend/js/home.js`
  - home page
- `frontend/js/snapshot_form.js`
  - new/edit snapshot page
- `frontend/js/snapshot_asset_modal.js`
  - create asset / asset type popup used inside snapshot form
- `frontend/js/snapshot_form_logic.js`
  - snapshot form logic helpers and keyboard behavior
- `frontend/js/asset_management.js`
  - asset management page wiring
- `frontend/js/asset_management_ui.js`
  - asset management form/table rendering helpers
- `frontend/js/asset_reorder.js`
  - reorder pages
- `frontend/js/progress_chart_logic.js`
  - Chart.js config and chart helpers
- `frontend/js/progress_goals.js`
  - progress page rendering and goal flow

### Charts

- Chart.js is vendored locally at:
  - `frontend/vendor/chart.bundle.min.js`
- Current version is `4.5.1`
- If chart behavior looks wrong, check against v4 docs, not v2 docs

## Editing Rules

- Prefer simple, explicit code
- Do not do unrelated refactors
- Ask when business behavior is unclear
- Keep file scope tight
- Avoid large abstractions for one-off UI behavior
- Do not push directly to `master`; create a feature branch first unless the user explicitly asks for a direct push
- Do not push to any remote unless the user explicitly permits it

## Validation Rules

- Validate input at handler/app boundary
- Service layer may assume sanitized input
- Still enforce business invariants in repository/service when required

Examples:

- duplicate snapshot date
- duplicate asset type name
- duplicate asset name within an asset type
- blocked edits to soft-deleted or unavailable entities

## File Size Guidance

Target:

- source files: under `800` lines
- tests: under `1200` lines

If a file grows past `500` lines, stop and consider whether it now contains more than one responsibility.

Current large-but-acceptable files are still focused by feature. Keep them from growing into multi-feature files.

## Testing Expectations

Write tests for non-trivial or risk-prone logic.

Must cover:

- financial calculations
- autofill behavior
- snapshot comparison
- latest-snapshot deactivation rules
- reorder behavior
- projection logic
- duplicate-name rules
- frontend keyboard behavior

Current test entrypoints:

- `go test ./...`
- `node --test frontend/*.test.js`

## Config and Runtime

Config loading:

- default config path: `./config/app.yaml`
- custom config path: `--config /path/to/config.yaml`

Important behavior:

- empty `db.path` => in-memory SQLite
- empty `log.path` => stdout only, no file output

## Privacy and Output Hygiene

- Do not generate files containing personal machine paths, OneDrive paths, JetBrains datasource paths, or real local usernames unless the user explicitly asks for them.
- Do not write real financial data into fixtures, demo seeds, screenshots, docs, tests, or migration helper scripts.
- For import or migration scripts, use:
  - project-relative paths
  - placeholders
  - sanitized sample values
- If inspecting the user DB is necessary, keep findings in chat or tests only. Do not commit their data back into the repo.

## Local Artifacts

These are local runtime files. Do not treat them as project source:

- `worthly-tracker.db`
- `logs/`
- `build/bin/`
- `.ai/`
- local `.idea/` data source files
- `frontend/wailsjs/` if Wails regenerates it locally

Do not delete the user’s DB or logs unless explicitly asked.

## Known UX / Implementation Notes

- Home, progress, and snapshot pages all use shared custom controls from `frontend/js/controls.js`
- Snapshot form keyboard navigation is dense and intentionally non-trivial; change carefully
- Asset creation from inside snapshot form should preserve current draft input
- Progress allocation opens from the summary table row into a popup
- Progress page live-updates date range changes; there is no Apply button anymore

## Logging

- Use structured `slog`
- Default level is `INFO`
- Avoid logging sensitive financial values unless there is a real debugging need

## Migrations

- Every schema change needs a migration
- Keep migrations small
- Update tests with any schema change
- Do not leave the app in a partially migrated state

## Release Hygiene

Before finishing a substantial task:

- run `go test ./...`
- run `node --test frontend/*.test.js`
- check for local absolute paths or personal-machine references
- check for dead files, dead imports, and stale docs
- update `README.md` only for user-facing behavior
- update `AGENTS.md` only for maintenance knowledge

## Project Skills

Project-local skills are intentionally trimmed to a small maintenance set. If a skill is missing, prefer normal repo exploration over re-adding a large generic skill bundle.

Keep only clearly useful repo-local skills:

- `wails`
- `modern-javascript-patterns`
- `golang-code-style`
- `golang-database`
- `golang-testing`
- `golang-troubleshooting`
- `golang-continuous-integration`
