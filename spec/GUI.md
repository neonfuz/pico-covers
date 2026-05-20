# GUI Interface

Fyne-based graphical interface for pico-covers. Runs alongside the CLI under a `gui` subcommand.

## Subcommand

```
pico-covers gui
```

Launches a native Fyne window. The CLI and GUI share all backend code via an event handler.

## Architecture

### Event System

The crawler does not print to stdout directly. Instead, it emits typed events through an `EventHandler` callback:

```go
type EventKind int

const (
    EventDBInit       EventKind = iota  // downloading database
    EventDBLoaded                       // "Loaded N ROM records"
    EventScanStart                      // "Scanning /path..."
    EventScanComplete                   // "Found N ROM files"
    EventROMStart                       // [index/total] RomName
    EventROMProgress                    // "Trying GameTDB...", "Trying LibRetro: URL"
    EventROMResult                      // ok / skip / notfound / error
    EventComplete                       // final summary
)

type ProgressEvent struct {
    Kind     EventKind
    Index    int
    Total    int
    RomName  string
    Status   string        // "ok", "skip", "notfound", "error"
    Detail   string        // URL, output path, error text, etc.
    Summary  *Summary
}

type EventHandler func(ProgressEvent)
```

- The CLI (`-v` / non-verbose) implements `EventHandler` by printing formatted lines to stdout/stderr.
- The GUI implements `EventHandler` by updating Fyne widgets.
- `database.Initialize` also accepts an optional `EventHandler` (nil means silent).

### Context Cancellation

`Crawler.Run(ctx, ...)` checks `ctx.Err()` between ROMs. The GUI provides a cancel button that cancels the context. The CLI never cancels (`context.Background()`).

## Window

- **Title**: "pico-covers"
- **Size**: 600x500 (resizable)
- **Icon**: Optional (app icon from bundled asset)
- **Theme**: Default Fyne light theme; dark theme support via Fyne's built-in toggle

## Views

Three views swapped in a single window (no tabs):

### 1. Config View (startup)

Form layout for all settings:

| Field | Widget | Default | Validation |
|-------|--------|---------|------------|
| ROM directory | `Entry` + `Button("Browse")` -> `dialog.ShowFolderOpen` | "" | Required, must exist |
| Covers directory | `Entry` + `Button("Browse")` | `_pico/covers` | Optional (defaults) |
| DB path | `Entry` + `Button("Browse")` -> `dialog.ShowFileOpen` | `~/.cache/pico-covers/nointro.db` | Optional |
| Concurrency | `Slider` (1-32, step 1, with label showing value) | `runtime.NumCPU()` | - |
| Refresh DB | `Check` | false | - |
| **Start** | `Button` (importance: `HighImportance`) | - | Disabled until ROM dir is valid |

Progress label shown during DB initialization ("Downloading NoIntro database...").

On Start: validate -> switch to progress view -> launch goroutine.

### 2. Progress View (during processing)

| Element | Widget | Description |
|---------|--------|-------------|
| Overall progress | `ProgressBar` (0.0-1.0) | `completed / total` from `EventROMResult` events |
| Progress label | `Label` | "12 / 50 ROMs processed" |
| Current ROM | `Label` | Current ROM filename |
| Status detail | `Label` | "Trying GameTDB...", "Downloaded, processing...", etc. |
| Status log | `List` (bind to `[]string`) | Scrollable log of completed items, newest at top |
| Cancel | `Button` | Cancels context, reverts to config view on completion |

Background: run `Crawler.Run()` in a goroutine. The `EventHandler` writes to a channel that the main goroutine drains via `fyne.Canvas.Refresh()` or `window.Canvas().Refresh()`. Do NOT update widgets from background goroutines -- Fyne requires UI updates on the main goroutine.

When processing finishes (both success and cancellation):
- If cancelled: show "Cancelled" message, return to config view
- If complete: show summary view with final counts

### 3. Summary View (after completion)

| Element | Widget | Description |
|---------|--------|-------------|
| Title | `Label` (bold, heading) | "Scan Complete" |
| Downloaded | `Label` (green) | "2 downloaded" |
| Skipped | `Label` (yellow/gold) | "1 skipped (already have it)" |
| Not found | `Label` (orange) | "3 not found" |
| Errors | `Label` (red) | "0 errors" |
| **Run Again** | `Button` | Switch back to config view |
| **Close** | `Button` | `window.Close()` |

## Fyne-Specific Notes

### Thread Safety

- All Fyne widget updates must happen on the main goroutine
- Use a channel-based approach: handler writes event -> channel -> goroutine reads and updates widgets
- Use `fyne.Do()` or schedule a `canvas.Refresh()` to ensure widgets repaint

### Folder Dialogs

Fyne's `dialog.ShowFolderOpen` returns a `fyne.ListableURI`. Convert to filesystem path with `uri.Path()`. The dialog is modal and returns via callback -- block on a channel or use a state variable updated by the callback.

### Window Close Handling

If the user closes the window while processing:
1. Cancel the context (the Cancel button already does this)
2. Wait for the crawler goroutine to finish via a done channel
3. The `window.SetOnClosed(func() { ... })` callback should handle cleanup

## Dependencies

| Library | Import | Purpose |
|---------|--------|---------|
| `fyne.io/fyne/v2` | `fyne.io/fyne/v2` | Cross-platform GUI toolkit |
| `fyne.io/fyne/v2/app` | `fyne.io/fyne/v2/app` | Application lifecycle |
| `fyne.io/fyne/v2/widget` | `fyne.io/fyne/v2/widget` | Widgets (Entry, Button, Label, ProgressBar, Slider, Check, List) |
| `fyne.io/fyne/v2/container` | `fyne.io/fyne/v2/container` | Layout containers (VBox, HBox, Form, Border, AppTabs) |
| `fyne.io/fyne/v2/dialog` | `fyne.io/fyne/v2/dialog` | Folder/file picker dialogs |
| `fyne.io/fyne/v2/theme` | `fyne.io/fyne/v2/theme` | Built-in icon theme |


## Build

Fyne requires Cgo and platform graphics libraries:
- **macOS**: Xcode command line tools (already present on most dev machines)
- **Linux**: X11 or Wayland development headers, OpenGL
- **Windows**: MSYS2 + MinGW

Build command:
```sh
CGO_ENABLED=1 go build -o pico-covers
```

For CLI-only builds (no Cgo required):
```sh
CGO_ENABLED=0 go build -tags "!fyne" -o pico-covers-cli
```
