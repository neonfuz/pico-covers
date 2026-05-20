package gui

import (
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

func Run() {
	a := app.NewWithID("com.neonfuz.pico-covers")
	w := a.NewWindow("pico-covers")
	w.Resize(fyne.NewSize(600, 500))

	if data, err := os.ReadFile("icon.png"); err == nil {
		w.SetIcon(fyne.NewStaticResource("icon.png", data))
	}

	state := newAppState(a, w)
	w.SetContent(newConfigView(state))
	w.ShowAndRun()
}

type appState struct {
	app    fyne.App
	window fyne.Window
	prefs  fyne.Preferences
	cancel contextFunc
}

type contextFunc func()

func newAppState(a fyne.App, w fyne.Window) *appState {
	s := &appState{
		app:    a,
		window: w,
		prefs:  a.Preferences(),
	}
	w.SetOnClosed(func() {
		if s.cancel != nil {
			s.cancel()
		}
	})
	return s
}

func (s *appState) switchToConfig() {
	s.window.SetContent(newConfigView(s))
}

func (s *appState) switchToProgress(romsDir, coversDir, dbPath string, concurrency int, refreshDB bool) {
	pv := newProgressView(s, romsDir, coversDir, dbPath, concurrency, refreshDB)
	s.window.SetContent(pv)
}

func (s *appState) switchToSummary(summary *summaryData) {
	s.window.SetContent(newSummaryView(s, summary))
}
