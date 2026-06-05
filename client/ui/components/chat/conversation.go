package components

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"github.com/albe194e/albz/client/db/sqlc/sql"
	ui "github.com/albe194e/albz/client/ui"
	"github.com/albe194e/albz/client/ui/components/base"
)

func ChatPanel(r *ui.Router) fyne.CanvasObject {
	messagesContainer := container.NewVBox()
	if len(r.Controller.State.Messages) == 0 {
		emptyState := container.NewCenter(base.Text("Choose a conversation and start talking.", base.TextStyle{
			Color: ui.CurrentTheme.MutedText,
		}))
		messagesContainer.Add(emptyState)
	} else {
		for _, msg := range r.Controller.State.Messages {
			messagesContainer.Add(MessageCard(r, msg))
		}
	}

	scroll := container.NewVScroll(messagesContainer)

	return base.Panel(scroll, base.PanelStyle{
		Fill:    ui.CurrentTheme.PanelFill,
		Stroke:  ui.CurrentTheme.PanelStroke,
		Radius:  18,
		MinSize: fyne.NewSize(0, 0),
		Padding: &base.Padding{
			Top:    16,
			Bottom: 16,
			Left:   16,
			Right:  16,
		},
	})
}

func MessageCard(r *ui.Router, msg sql.Message) fyne.CanvasObject {
	isOwn := r.Controller.State.CurrentUser != nil && msg.SenderID == r.Controller.State.CurrentUser.ID

	fill := ui.CurrentTheme.Surface
	stroke := ui.CurrentTheme.Border
	if isOwn {
		fill = ui.CurrentTheme.PrimaryAccent
		stroke = ui.CurrentTheme.AccentHover
	}

	body := widget.NewLabel(msg.Body)
	body.Wrapping = fyne.TextWrapWord

	statusColor := ui.CurrentTheme.MutedText
	switch msg.DeliveryState {
	case "delivered":
		statusColor = ui.CurrentTheme.SuccessOnline
	case "recipient_offline", "pending":
		statusColor = ui.CurrentTheme.WarningDelayed
	case "failed", "rejected":
		statusColor = ui.CurrentTheme.ErrorFailed
	}

	status := base.Text(msg.DeliveryState, base.TextStyle{
		Size:  11,
		Color: statusColor,
		Bold:  true,
	})
	status.Alignment = fyne.TextAlignTrailing

	bubble := base.Panel(container.NewVBox(body, status), base.PanelStyle{
		Fill:   fill,
		Stroke: stroke,
		Radius: 18,
		Padding: &base.Padding{
			Top:    12,
			Bottom: 12,
			Left:   14,
			Right:  14,
		},
		MinSize: fyne.NewSize(180, 0),
		Margin: &base.Margin{
			Bottom: 10,
		},
	})

	if isOwn {
		return container.NewHBox(layout.NewSpacer(), bubble)
	}

	return container.NewHBox(bubble, layout.NewSpacer())
}
