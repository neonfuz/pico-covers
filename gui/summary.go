package gui

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func newSummaryView(state *appState, data *summaryData) fyne.CanvasObject {
	if data.Cancelled {
		return container.NewPadded(
			container.NewVBox(
				widget.NewLabelWithStyle("Cancelled", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
				layout.NewSpacer(),
				widget.NewButton("Run Again", func() {
					state.switchToConfig()
				}),
				widget.NewButton("Close", func() {
					state.window.Close()
				}),
			),
		)
	}

	greenColor := color.NRGBA{R: 0, G: 180, B: 0, A: 255}
	amberColor := color.NRGBA{R: 200, G: 160, B: 0, A: 255}
	orangeColor := color.NRGBA{R: 220, G: 120, B: 0, A: 255}
	redColor := color.NRGBA{R: 220, G: 0, B: 0, A: 255}

	title := canvas.NewText("Scan Complete", color.White)
	title.TextSize = 18
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Alignment = fyne.TextAlignCenter

	downloaded := canvas.NewText(fmt.Sprintf("%d downloaded", data.Downloaded), greenColor)
	skipped := canvas.NewText(fmt.Sprintf("%d skipped (already have it)", data.Skipped), amberColor)
	notFound := canvas.NewText(fmt.Sprintf("%d not found", data.NotFound), orangeColor)
	errors := canvas.NewText(fmt.Sprintf("%d errors", data.Errors), redColor)

	runAgainBtn := widget.NewButton("Run Again", func() {
		state.switchToConfig()
	})
	closeBtn := widget.NewButton("Close", func() {
		state.window.Close()
	})

	return container.NewPadded(
		container.NewVBox(
			title,
			widget.NewSeparator(),
			downloaded,
			skipped,
			notFound,
			errors,
			widget.NewSeparator(),
			layout.NewSpacer(),
			runAgainBtn,
			closeBtn,
		),
	)
}
