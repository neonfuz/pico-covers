package gui

import (
	"context"
	"fmt"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"github.com/neonfuz/pico-covers/crawler"
	"github.com/neonfuz/pico-covers/database"
	"github.com/neonfuz/pico-covers/events"
)

type summaryData struct {
	Downloaded int
	Skipped    int
	NotFound   int
	Errors     int
	Cancelled  bool
}

type progressState struct {
	mu         sync.Mutex
	completed  int
	total      int
	currentROM string
	status     string
	items      []string
	done       bool
}

func (ps *progressState) addItem(s string) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.items = append([]string{s}, ps.items...)
	if len(ps.items) > 200 {
		ps.items = ps.items[:200]
	}
}

func (ps *progressState) snapshot() (completed, total int, rom, status string, items []string) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	return ps.completed, ps.total, ps.currentROM, ps.status, ps.items
}

func newProgressView(state *appState, romsDir, coversDir, dbPath string, concurrency int, refreshDB bool) fyne.CanvasObject {
	ps := &progressState{}

	progressBar := widget.NewProgressBar()
	progressLabel := widget.NewLabel("0 / 0 ROMs processed")
	currentROMLabel := widget.NewLabel("")
	statusLabel := widget.NewLabel("Initializing...")

	logList := widget.NewList(
		func() int {
			_, _, _, _, items := ps.snapshot()
			return len(items)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.ListItemID, o fyne.CanvasObject) {
			_, _, _, _, items := ps.snapshot()
			if id < len(items) {
				o.(*widget.Label).SetText(items[id])
			}
		},
	)

	cancelled := false
	cancelLabel := widget.NewLabel("Cancelling...")
	cancelLabel.Hide()

	var cancelBtn *widget.Button
	cancelBtn = widget.NewButton("Cancel", func() {
		if state.cancel != nil {
			cancelled = true
			state.cancel()
			cancelBtn.Disable()
			cancelLabel.Show()
		}
	})

	content := container.NewBorder(
		nil,
		container.NewVBox(
			cancelLabel,
			cancelBtn,
		),
		nil,
		nil,
		container.NewVBox(
			progressBar,
			progressLabel,
			widget.NewSeparator(),
			widget.NewLabel("Current ROM:"),
			currentROMLabel,
			widget.NewSeparator(),
			widget.NewLabel("Status:"),
			statusLabel,
			widget.NewSeparator(),
			widget.NewLabel("Completed:"),
			logList,
			layout.NewSpacer(),
		),
	)

	ctx, cancel := context.WithCancel(context.Background())
	state.cancel = func() {
		cancel()
		state.cancel = nil
	}

	eventCh := make(chan events.ProgressEvent, 128)
	doneCh := make(chan struct {
		summary *crawler.Summary
		err     error
	}, 1)

	handler := func(ev events.ProgressEvent) {
		select {
		case eventCh <- ev:
		default:
		}
	}

	go func() {
		db := database.NewDatabase()
		if err := db.Initialize(ctx, dbPath, refreshDB, handler); err != nil {
			doneCh <- struct {
				summary *crawler.Summary
				err     error
			}{nil, fmt.Errorf("database init: %w", err)}
			return
		}

		c := crawler.New(romsDir, coversDir, db)
		summary, runErr := c.Run(ctx, concurrency, handler)
		doneCh <- struct {
			summary *crawler.Summary
			err     error
		}{summary, runErr}
	}()

	ticker := time.NewTicker(50 * time.Millisecond)
	go func() {
		defer ticker.Stop()
		defer close(eventCh)

		for {
			select {
			case ev, ok := <-eventCh:
				if !ok {
					return
				}
				ps.mu.Lock()
				switch ev.Kind {
				case events.EventROMStart:
					ps.total = ev.Total
					ps.currentROM = ev.ROMFile
					ps.status = "Processing..."
				case events.EventROMSuccess:
					ps.completed = ev.Completed
					ps.status = "Done"
					if ev.GameTitle != "" {
						ps.items = append([]string{fmt.Sprintf("OK: %s", ev.GameTitle)}, ps.items...)
					}
				case events.EventROMSkipped:
					ps.completed = ev.Completed
					ps.status = "Skipped"
					ps.items = append([]string{fmt.Sprintf("SKIP: %s (%s)", ev.ROMFile, ev.GameTitle)}, ps.items...)
				case events.EventROMNotFound:
					ps.completed = ev.Completed
					ps.status = "Not found"
					ps.items = append([]string{fmt.Sprintf("NOT FOUND: %s", ev.ROMFile)}, ps.items...)
				case events.EventROMError:
					ps.completed = ev.Completed
					ps.status = "Error"
					ps.items = append([]string{fmt.Sprintf("ERROR: %s - %s", ev.ROMFile, ev.Detail)}, ps.items...)
				case events.EventInfo:
					ps.status = ev.Detail
				case events.EventDBInit:
					ps.status = ev.Detail
				case events.EventDBLoaded:
					ps.status = fmt.Sprintf("Loaded %d ROM records", ev.Total)
				}
				if len(ps.items) > 200 {
					ps.items = ps.items[:200]
				}
				ps.mu.Unlock()
			case <-ticker.C:
			}

			completed, total, rom, status, itemsLen := 0, 0, "", "", 0
			ps.mu.Lock()
			completed = ps.completed
			total = ps.total
			rom = ps.currentROM
			status = ps.status
			itemsLen = len(ps.items)
			ps.mu.Unlock()

			if total > 0 {
				progressBar.Max = float64(total)
				progressBar.Value = float64(completed)
				progressLabel.SetText(fmt.Sprintf("%d / %d ROMs processed", completed, total))
			}
			currentROMLabel.SetText(rom)
			statusLabel.SetText(status)
			progressBar.Refresh()
			progressLabel.Refresh()
			currentROMLabel.Refresh()
			statusLabel.Refresh()
			if itemsLen > 0 {
				logList.Refresh()
			}

			select {
			case result := <-doneCh:
				state.cancel = nil
				if result.err != nil && !cancelled {
					dialog.ShowError(result.err, state.window)
					state.switchToConfig()
				} else if result.summary != nil {
					sd := &summaryData{
						Downloaded: result.summary.Succeeded,
						Skipped:    result.summary.Skipped,
						NotFound:   result.summary.NotFound,
						Errors:     result.summary.Errors,
					}
					state.switchToSummary(sd)
				} else {
					state.switchToConfig()
				}
				return
			default:
			}
		}
	}()

	return content
}
