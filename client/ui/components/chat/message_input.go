package components

import (
	"context"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	ui "github.com/albe194e/albz/client/ui"
	"github.com/albe194e/albz/client/ui/components/base"
)

func MessageInputPanel(r *ui.Router) fyne.CanvasObject {
	input := widget.NewEntry()
	input.SetPlaceHolder("Type your message...")
	input.Wrapping = fyne.TextWrapWord

	button := base.Button("Send",
		func() {
			err := r.Controller.AddMessage(context.Background(), r.State.ActiveConversationID, input.Text)
			if err != nil {
				fmt.Printf("Error sending message: %v\n", err)
				return
			}
			input.SetText("")
		},
		nil,
	)

	content := container.NewBorder(nil, nil, nil, button, base.FramedEntry(input))

	return base.Panel(content, base.PanelStyle{
		Fill:   ui.CurrentTheme.Surface,
		Stroke: ui.CurrentTheme.Border,
		Radius: 18,
		Padding: &base.Padding{
			Top:    12,
			Bottom: 12,
			Left:   12,
			Right:  12,
		},
		Margin: &base.Margin{
			Top: 12,
		},
	})
}
