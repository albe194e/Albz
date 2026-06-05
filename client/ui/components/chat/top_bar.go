package components

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"

	ui "github.com/albe194e/albz/client/ui"
	"github.com/albe194e/albz/client/ui/components/base"
)

func TopBar(r *ui.Router) fyne.CanvasObject {
	chatName := base.H1Binding(r.State.ActiveConversationName, base.TextStyle{
		Color: ui.CurrentTheme.PrimaryText,
		Bold:  true})

	statusText := "Offline"
	statusColor := ui.CurrentTheme.ErrorFailed
	if r.Controller.State.ServerConnected {
		statusText = "Online"
		statusColor = ui.CurrentTheme.SuccessOnline
	}

	status := base.Text(statusText, base.TextStyle{
		Size:  12,
		Color: statusColor,
		Bold:  true,
	})

	right := fyne.CanvasObject(status)
	if r.IsMobileLayout() {
		toggleLabel := "Conversations"
		if r.State.SidebarOpen {
			toggleLabel = "Hide"
		}

		right = container.NewHBox(
			status,
			base.Button(toggleLabel, func() {
				r.State.SidebarOpen = !r.State.SidebarOpen
				r.NavigateTo(r.State.Page)
			}, &base.BStyle{
				Fill:    ui.CurrentTheme.Surface,
				Border:  ui.CurrentTheme.Border,
				Text:    ui.CurrentTheme.PrimaryText,
				MinSize: fyne.NewSize(126, 40),
			}),
		)
	}

	header := container.NewBorder(nil, nil, nil, right, chatName)

	return base.Panel(header, base.PanelStyle{
		Fill:    ui.CurrentTheme.Surface,
		Stroke:  ui.CurrentTheme.Border,
		Radius:  18,
		MinSize: fyne.NewSize(0, 64),
		Padding: &base.Padding{
			Top:    14,
			Bottom: 14,
			Left:   16,
			Right:  16,
		},
		Margin: &base.Margin{
			Bottom: 12,
		},
	})
}
