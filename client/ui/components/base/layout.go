package base

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"

	ui "github.com/albe194e/albz/client/ui"
)

func Card(content fyne.CanvasObject) fyne.CanvasObject {
	return Panel(content, PanelStyle{
		Fill:   ui.CurrentTheme.ElevatedCard,
		Stroke: ui.CurrentTheme.Border,
		Radius: 18,
		Padding: &Padding{
			Top:    20,
			Bottom: 20,
			Left:   20,
			Right:  20,
		},
	})
}

func Section(title string, body fyne.CanvasObject) fyne.CanvasObject {
	return container.NewVBox(
		H2(title, TextStyle{
			Color: ui.CurrentTheme.PrimaryText,
			Bold:  true,
		}),
		body,
	)
}

func Centered(content fyne.CanvasObject, margin float32) fyne.CanvasObject {
	return container.New(
		layout.NewCustomPaddedLayout(margin, margin, margin, margin),
		container.NewCenter(content),
	)
}
