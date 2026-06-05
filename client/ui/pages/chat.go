package pages

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"

	ui "github.com/albe194e/albz/client/ui"
	chat "github.com/albe194e/albz/client/ui/components/chat"
)

func init() {
	ui.RegisterPageRenderer(ui.Chat, ChatPage)
}

func ChatPage(router *ui.Router) fyne.CanvasObject {
	chatBody := container.NewBorder(
		nil,
		chat.MessageInputPanel(router),
		nil,
		nil,
		chat.ChatPanel(router),
	)

	body := chatBody
	if router.IsMobileLayout() && router.State.SidebarOpen {
		body = container.NewStack(
			chatBody,
			mobileSidebarOverlay(router),
		)
	} else if !router.IsMobileLayout() {
		body = container.NewBorder(
			nil,
			nil,
			chat.ConversationListPanel(router),
			nil,
			chatBody,
		)
	}

	return container.NewBorder(
		chat.TopBar(router),
		nil,
		nil,
		nil,
		body,
	)
}

func mobileSidebarOverlay(router *ui.Router) fyne.CanvasObject {
	dimmer := canvas.NewRectangle(color.NRGBA{A: 150})

	sidebar := container.NewBorder(
		nil,
		nil,
		chat.ConversationListPanel(router),
		nil,
		nil,
	)

	return container.NewStack(
		dimmer,
		container.NewHBox(
			sidebar,
			layout.NewSpacer(),
		),
	)
}
