package gui

import (
	"fmt"
	"os"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func newConfigView(state *appState) fyne.CanvasObject {
	prefs := state.prefs

	romDirEntry := widget.NewEntry()
	romDirEntry.SetPlaceHolder("Select ROM directory...")
	if v := prefs.StringWithFallback("roms_dir", ""); v != "" {
		romDirEntry.SetText(v)
	}

	romBrowseBtn := widget.NewButton("Browse", func() {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil || uri == nil {
				return
			}
			romDirEntry.SetText(uri.Path())
			prefs.SetString("roms_dir", uri.Path())
		}, state.window)
	})

	coversDirEntry := widget.NewEntry()
	if v := prefs.StringWithFallback("covers_dir", "_pico/covers"); v != "" {
		coversDirEntry.SetText(v)
	} else {
		coversDirEntry.SetText("_pico/covers")
	}
	coversDirEntry.OnChanged = func(s string) {
		prefs.SetString("covers_dir", s)
	}

	coversBrowseBtn := widget.NewButton("Browse", func() {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil || uri == nil {
				return
			}
			coversDirEntry.SetText(uri.Path())
			prefs.SetString("covers_dir", uri.Path())
		}, state.window)
	})

	dbPathEntry := widget.NewEntry()
	dbPathEntry.SetPlaceHolder("~/.cache/pico-covers/nointro.db")
	if v := prefs.StringWithFallback("db_path", ""); v != "" {
		dbPathEntry.SetText(v)
	}
	dbPathEntry.OnChanged = func(s string) {
		prefs.SetString("db_path", s)
	}

	dbBrowseBtn := widget.NewButton("Browse", func() {
		dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil || reader == nil {
				return
			}
			dbPathEntry.SetText(reader.URI().Path())
			prefs.SetString("db_path", reader.URI().Path())
			reader.Close()
		}, state.window)
	})

	concurrencyLabel := widget.NewLabel("")
	defaultConc := runtime.NumCPU()
	if v := prefs.IntWithFallback("concurrency", defaultConc); v > 0 {
		defaultConc = v
	}
	if defaultConc < 1 {
		defaultConc = 1
	}
	if defaultConc > 32 {
		defaultConc = 32
	}
	concurrencyLabel.SetText(fmt.Sprintf("Concurrency: %d", defaultConc))

	concurrencySlider := widget.NewSlider(1, 32)
	concurrencySlider.Step = 1
	concurrencySlider.Value = float64(defaultConc)
	concurrencySlider.OnChanged = func(v float64) {
		c := int(v)
		concurrencyLabel.SetText(fmt.Sprintf("Concurrency: %d", c))
		prefs.SetInt("concurrency", c)
	}

	refreshCheck := widget.NewCheck("Refresh database", func(b bool) {
		prefs.SetBool("refresh_db", b)
	})
	refreshCheck.SetChecked(prefs.BoolWithFallback("refresh_db", false))

	statusLabel := widget.NewLabel("")
	statusLabel.Hide()

	startBtn := widget.NewButton("Start", nil)
	startBtn.Importance = widget.HighImportance
	startBtn.Disable()

	validateROMDir := func() {
		s := romDirEntry.Text
		if s != "" {
			if info, err := os.Stat(s); err == nil && info.IsDir() {
				startBtn.Enable()
				return
			}
		}
		startBtn.Disable()
	}

	romDirEntry.OnChanged = func(s string) {
		prefs.SetString("roms_dir", s)
		validateROMDir()
	}

	startBtn.OnTapped = func() {
		romDir := romDirEntry.Text
		if romDir == "" {
			dialog.ShowError(fmt.Errorf("ROM directory is required"), state.window)
			return
		}
		if info, err := os.Stat(romDir); err != nil || !info.IsDir() {
			dialog.ShowError(fmt.Errorf("ROM directory does not exist: %s", romDir), state.window)
			return
		}

		coversDir := coversDirEntry.Text
		if coversDir == "" {
			coversDir = "_pico/covers"
		}

		dbPath := dbPathEntry.Text
		concurrency := int(concurrencySlider.Value)
		refreshDB := refreshCheck.Checked

		state.switchToProgress(romDir, coversDir, dbPath, concurrency, refreshDB)
	}

	romRow := container.NewBorder(nil, nil, nil, romBrowseBtn,
		widget.NewLabel("ROM directory"),
	)
	coversRow := container.NewBorder(nil, nil, nil, coversBrowseBtn,
		widget.NewLabel("Covers directory"),
	)
	dbRow := container.NewBorder(nil, nil, nil, dbBrowseBtn,
		widget.NewLabel("DB path"),
	)

	form := container.NewVBox(
		romRow,
		romDirEntry,
		widget.NewSeparator(),
		coversRow,
		coversDirEntry,
		widget.NewSeparator(),
		dbRow,
		dbPathEntry,
		widget.NewSeparator(),
		concurrencyLabel,
		concurrencySlider,
		widget.NewSeparator(),
		refreshCheck,
		widget.NewSeparator(),
		statusLabel,
		layout.NewSpacer(),
		startBtn,
	)

	return container.NewPadded(form)
}
