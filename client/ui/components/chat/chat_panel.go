package components

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"

	clientapp "github.com/albe194e/albz/client/app"
	ui "github.com/albe194e/albz/client/ui"
)

func ChatPanel(c *clientapp.Controller) fyne.CanvasObject {
	bg := canvas.NewRectangle(ui.CurrentTheme.PanelFill)

	messagesContainer := container.NewVBox()

	for _, msg := range c.State.Messages {
		fmt.Println(msg.Body)

		msgLabel := canvas.NewText(msg.Body, ui.CurrentTheme.PrimaryText)
		msgLabel.TextSize = 14

		messagesContainer.Add(msgLabel)
	}

	return container.NewStack(
		bg,
		container.NewPadded(messagesContainer),
	)
}
