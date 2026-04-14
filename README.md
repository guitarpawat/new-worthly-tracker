# Worthly Tracker

Worthly Tracker is a desktop app for tracking personal net worth as date-based snapshots. 
It is built with Go, Wails, and SQLite, with a spreadsheet-style light UI aimed at quick monthly updates and long-term review.

## What It Does

- Record net worth snapshots by date
- Group assets by asset type
- Compare the current snapshot with the previous snapshot
- Add, edit, delete, and reorder assets and asset types
- Track progress over time with trend charts, allocation charts, and goals
- Keep historical snapshots intact with soft delete

## Main Concepts

- A snapshot is a dated record set.
- One asset can appear at most once per snapshot date.
- The latest snapshot is the highest active snapshot date.
- The previous snapshot is the second latest active snapshot date.
- Removing an asset from the latest snapshot makes it inactive for future autofill.

## Requirements

- Go 1.26+
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

- Go 1.26+
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

- Go 1.26+
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

## Future Plans
- Add summary for each asset type in homepage
- Summary table should be sortable by date. The latest should be on the top by default.
- Order allocation summary by snapshot ordering.

# Notices
The author of this project uses AI to generate code for almost every part in this project. 
The purpose of this project is to learn how to use AI agents and see what is the current state of what it could do.

## License

This program is free software. It comes without any warranty, to
the extent permitted by applicable law. You can redistribute it
and/or modify it under the terms of the Do What The Fuck You Want
To Public License, Version 2, as published by Sam Hocevar. See
http://www.wtfpl.net/ for more details.
