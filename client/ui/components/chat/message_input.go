package components

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	clientapp "github.com/albe194e/albz/client/app"
	ui "github.com/albe194e/albz/client/ui"
)

func MessageInputPanel(c *clientapp.Controller) fyne.CanvasObject {
	panel := canvas.NewRectangle(ui.CurrentTheme.PanelFill)
	panel.SetMinSize(fyne.NewSize(0, 50))

	button := widget.NewButton("Send", func() {
		c.AddMessage()
	})

	input := widget.NewEntry()
	input.SetPlaceHolder("Type your message...")

	return container.NewBorder(
		nil,
		nil,
		nil,
		button,
		container.NewStack(
			panel,
			input,
		),
	)
}
