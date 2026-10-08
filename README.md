# Worthly Tracker

Worthly Tracker is a desktop app for tracking personal net worth as date-based snapshots. 
It is built with Go, Wails, and SQLite, with a spreadsheet-style light UI aimed at quick monthly updates and long-term review.

## What It Does

- Record net worth snapshots by date
- Export the displayed snapshot from the hamburger menu as CSV using a native Save dialog
- Group assets by asset type
- Show per-asset-type summaries on the home page
- Compare the current snapshot with the previous snapshot
- Add, edit, delete, and reorder assets and asset types
- Use custom in-app date pickers and dropdowns with keyboard support
- Track progress over time with trend charts, summary tables, allocation popups, and goals
- Create custom allocation charts with percentage-based categories for each asset
- Project total net worth with a hybrid forecast: cash and liabilities follow recent monthly drift, while non-cash assets compound from recent growth
- Jump to the latest 12/18/24/36-month progress range or all records quickly
- Keep historical snapshots intact with soft delete

## Main Concepts

- A snapshot is a dated record set.
- One asset can appear at most once per snapshot date.
- The latest snapshot is the highest active snapshot date.
- The previous snapshot is the second latest active snapshot date.
- Removing an asset from the latest snapshot makes it inactive for future autofill.

## Managing Assets and Types

The management menu is ordered **Manage Asset → Reorder Asset → Manage Type → Reorder Type → Manage Chart**. Type is short for Asset Type.

Manage Asset and Manage Type show lists with **Add Asset** / **Add Type** buttons. Click a row to edit it in a popup. Deletion requires confirmation and is a soft delete. Assets can only be deleted if they have no snapshot records, including records in deleted snapshots. Types can only be deleted after all their assets have been moved or deleted; inactive assets still count.

## Custom Allocations

Open **Manage Asset → Manage Chart** and choose **Add Chart** to create a chart such as **Asset Allocation Country** or **Asset Allocation Type**. Click an existing chart name to rename or delete it in a popup.

To assign tags, open **Manage Asset**, select an asset, and choose a chart from the **Chart Tags** dropdown. Add a **Category** and **Allocation (%)**, then choose **Save Asset**. Switching charts preserves your draft tags. Asset details and all chart tags save together; closing the form discards the draft.

Each asset can have different categories in each chart. Percentages allow two decimal places and must total at most 100% per asset. The remaining percentage, including assets with no categories, appears as **Unallocated**. These settings apply to all snapshots; each chart uses the asset values from the snapshot being viewed.

Assets are included in every custom chart by default. In **Manage Asset → Chart Tags**, choose a chart and turn off **Include in chart** to exclude that asset from its values, percentages, and Unallocated total. Save Asset to apply the change. Previously saved tags are retained for re-inclusion; other charts and snapshot values are unaffected.

Once a chart is saved, the Allocation popup on Home and Progress shows **Custom** beside the built-in chart buttons. Click it to choose a chart from the menu. Negative category values use a bar chart.

## Requirements

- Go 1.27+
- Node.js for frontend tests
- Wails CLI
- SQLite is embedded through `modernc.org/sqlite`, so CGO is not required

## Development

Start the Wails development server with auto reload:

### Linux

On Fedora and similar newer distributions:

```bash
wails dev -tags webkit2_41
```

### Windows

```powershell
wails dev
```

### macOS

```bash
wails dev
```

## Build

### Linux

```bash
wails build -tags webkit2_41
```

The built binary will be written to `build/bin/`.

### Windows

Install first:

- Go 1.27+
- Node.js
- Wails CLI
- Microsoft WebView2 runtime

Then build from the project root:

```powershell
wails build
```

If you want Wails to verify your environment first:

```powershell
wails doctor
```

The built executable will be written to `build/bin/`.

### macOS

Install first:

- Go 1.27+
- Node.js
- Wails CLI
- Xcode Command Line Tools

Install the Apple build tools if needed:

```bash
xcode-select --install
```

Then build from the project root:

```bash
wails build
```

If you want Wails to verify your environment first:

```bash
wails doctor
```

The built application will be written to `build/bin/`.

## Test

Backend:

```bash
go test ./...
```

Frontend:

```bash
node --test frontend/*.test.js
```

## Configuration

Default config path:

- `./config/app.yaml`

The app also accepts:

- `--config /path/to/config.yaml`

Example:

```yaml
name: Worthly Tracker
env: development
db:
  path: /path/to/worthly-tracker.db
log:
  path: /path/to/worthly-tracker.log
  level: INFO
```

Behavior:

- `db.path` is optional
  - if omitted or empty, the app uses an in-memory database
- `log.path` is optional
  - if omitted or empty, no log file is created
- `log.level` defaults to `INFO`

## Data and Logs

Common local files created while running the app:

- `worthly-tracker.db`
- `logs/`
- `build/bin/`

These are local runtime artifacts and are ignored by git.

## Dependency Updates

Update Go dependencies with:

```bash
./scripts/update-deps.sh
```

This runs:

- `go get -u ./...`
- `go mod tidy`

## Project Status

Current implemented areas:

- Home snapshot view
- New snapshot
- Edit snapshot
- Delete snapshot
- Asset management
- Asset/asset-type reorder
- Progress and goals

# Notices
The author of this project uses AI to generate code for almost every part in this project. 
The purpose of this project is to learn how to use AI agents and see what is the current state of what it could do.

## License

This program is free software. It comes without any warranty, to
the extent permitted by applicable law. You can redistribute it
and/or modify it under the terms of the Do What The Fuck You Want
To Public License, Version 2, as published by Sam Hocevar. See
http://www.wtfpl.net/ for more details.
