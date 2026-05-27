package base

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"github.com/albe194e/albz/client/ui"
)

func H1(text string) *canvas.Text {
	t := canvas.NewText(text, ui.CurrentTheme.PrimaryText)
	t.TextSize = 22
	t.TextStyle = fyne.TextStyle{Bold: true}
	return t
}

func H2(text string) *canvas.Text {
	t := canvas.NewText(text, ui.CurrentTheme.PrimaryText)
	t.TextSize = 18
	t.TextStyle = fyne.TextStyle{Bold: true}
	return t
}

func H3(text string) *canvas.Text {
	t := canvas.NewText(text, ui.CurrentTheme.PrimaryText)
	t.TextSize = 14
	t.TextStyle = fyne.TextStyle{Bold: true}
	return t
}
