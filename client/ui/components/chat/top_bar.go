package components

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"

	ui "github.com/albe194e/albz/client/ui"
)

func TopBar() fyne.CanvasObject {
	bar := canvas.NewRectangle(ui.CurrentTheme.PanelFill)
	bar.SetMinSize(fyne.NewSize(0, 50))

	return bar
}
