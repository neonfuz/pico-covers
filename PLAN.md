# Implementation Plan: pico-covers

A CLI and GUI tool written in Go that downloads boxart for ROM files and saves them as 8bpp BMP covers for pico-launcher.

## Spec Documents

- [Cover format](spec/Covers.md) -- pico-launcher BMP requirements
- [ROM parsing](spec/ROM-Parsing.md) -- header detection, SHA1, type identification
- [Database](spec/Database.md) -- NoIntro DB download, cache, lookup
- [Cover sources](spec/Cover-Sources.md) -- GameTDB and LibRetro download logic
- [Image processing](spec/Image-Processing.md) -- resize, quantize, dither, BMP encode
- [Output layout](spec/Output-Layout.md) -- file naming and directory structure
- [CLI](spec/CLI.md) -- command-line interface design
- [GUI](spec/GUI.md) -- Fyne-based graphical interface

## Phase 1-5: Core Implementation (DONE)

All ROM parsing, database, cover download, image processing, crawler, and CLI are implemented.

## Phase 6: Polish

- [ ] Add `golangci-lint` config and fix all warnings
- [ ] Verify BMP output is valid 8bpp with correct dimensions (128x96) using an image viewer
- [ ] Test on a real pico-launcher SD card layout
- [ ] Update README.md with GUI usage instructions and `gui` subcommand

## Phase 7: Package Layout (internal/ -> top-level)

Move all packages from `internal/` to project root. Since pico-covers is an application (not a library), top-level packages are cleaner and more idiomatic.

- [ ] Move `internal/rom/` -> `rom/`
- [ ] Move `internal/cover/` -> `cover/`
- [ ] Move `internal/database/` -> `database/`
- [ ] Move `internal/crawler/` -> `crawler/`
- [ ] Delete empty `internal/` directory
- [ ] Update all import paths from `internal/X` to `X` across all Go files
- [ ] Run `go vet ./...` and existing tests to verify nothing broke

## Phase 8: Event System Refactor

Replace `fmt.Printf`-based output with an event handler system so both CLI and GUI can consume progress updates.

### 8.1 Crawler event types

- [ ] Add `EventKind` enum and `ProgressEvent` struct to `crawler/events.go` per [GUI spec](spec/GUI.md)
- [ ] Add `EventHandler func(ProgressEvent)` callback type
- [ ] Refactor `Crawler.Run()` signature:
  - Add `handler EventHandler` parameter
  - Remove `verbose bool` parameter
- [ ] Replace all `fmt.Printf` calls in `crawler.go` with `handler(event)` calls
- [ ] Add context cancellation: check `ctx.Err()` between ROM processing iterations; return early if cancelled

### 8.2 Database progress events

- [ ] Add optional `EventHandler` parameter to `database.Initialize()`
- [ ] Emit `EventDBInit` when starting download, `EventDBLoaded` with record count
- [ ] Nil handler = silent (no events emitted)

### 8.3 CLI integration

- [ ] Update `main.go` to pass a CLI event handler to `crawler.Run()`
- [ ] Handler respects `-v` flag: verbose shows all event types, non-verbose shows only ROM results and summary
- [ ] Verify CLI output is identical to current behavior

### 8.4 Test

- [ ] Run `go vet ./...`
- [ ] Run existing tests (`go test ./...`)
- [ ] Manual smoke test: `./pico-covers -roms testdata/`
- [ ] Manual smoke test with `-v` flag

## Phase 9: Fyne GUI

### 9.1 Dependency

- [ ] Run `go get fyne.io/fyne/v2`
- [ ] Verify `go.sum` is updated

### 9.2 GUI package

Create `gui/` package per [GUI spec](spec/GUI.md):

- [ ] `gui/app.go` -- `func Run()` entrypoint
  - Create `app.New()` and `window.NewWindow("pico-covers")`
  - Window size 600x500, resizable
  - Manage view switching (config -> progress -> summary)
  - Handle `window.ShowAndRun()`

- [ ] `gui/config.go` -- settings form view
  - ROM dir entry + browse button (folder dialog)
  - Covers dir entry + browse button
  - DB path entry + browse button (file dialog)
  - Concurrency slider (1-32)
  - Refresh DB checkbox
  - Start button (high importance), disabled until ROM dir is valid and exists
  - "Loading database..." label during init
  - On Start: build `database.Database`, initialize it, then build `crawler.Crawler`, switch to progress view

- [ ] `gui/progress.go` -- live processing view
  - Progress bar (0.0-1.0 from completed/total)
  - Label: "N / Total ROMs processed"
  - Label: current ROM filename
  - Label: current status detail
  - List: scrollable log of completed items
  - Cancel button (cancels context)
  - Uses goroutine for `crawler.Run()` with channel-based UI updates
  - On completion: switch to summary view
  - On cancellation: show cancelled message, switch to config view

- [ ] `gui/summary.go` -- results view
  - Downloaded / skipped / not found / errors counts with colored labels
  - "Run Again" button -> switch to config view
  - "Close" button -> `window.Close()`

### 9.3 Main entrypoint

- [ ] Add `gui` subcommand detection to `main.go`:
  ```go
  if len(os.Args) > 1 && os.Args[1] == "gui" {
      gui.Run()
      return
  }
  ```

### 9.4 Thread safety

- [ ] Verify all Fyne widget updates happen on main goroutine (use channel + refresh)
- [ ] Handle window close during processing (cancel context, wait for goroutine)

### 9.5 Test

- [ ] Test GUI launch on macOS
- [ ] Verify all three views render correctly
- [ ] Verify cancel button stops processing
- [ ] Verify summary shows correct counts after a full run

## Phase 10: Final Polish

- [ ] Run `go vet ./...`, `golangci-lint run ./...` -- no errors
- [ ] Run all tests -- all pass
- [ ] Update README.md:
  - Add GUI section
  - Document `pico-covers gui` usage
  - Note Cgo requirement for GUI builds
- [ ] Add `pico-covers gui` to help text
