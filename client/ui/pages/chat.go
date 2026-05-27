package pages

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"

	ui "github.com/albe194e/albz/client/ui"
	chat "github.com/albe194e/albz/client/ui/components/chat"
)

func init() {
	ui.RegisterPageRenderer(ui.Chat, ChatPage)
}

func ChatPage(router *ui.Router) fyne.CanvasObject {
	mainArea := container.NewBorder(
		chat.TopBar(),
		chat.MessageInputPanel(router.Controller),
		nil,
		nil,
		chat.ChatPanel(router.Controller),
	)

	basePanel := container.NewBorder(
		nil,
		nil,
		chat.ConversationListPanel(router.Controller),
		nil,
		mainArea,
	)

	return basePanel
}
