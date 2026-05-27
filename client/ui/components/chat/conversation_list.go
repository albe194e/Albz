package components

import (
	"context"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"

	clientapp "github.com/albe194e/albz/client/app"
	"github.com/albe194e/albz/client/db/sqlc/sql"
	ui "github.com/albe194e/albz/client/ui"
	"github.com/albe194e/albz/client/ui/components/base"
)

func ConversationListPanel(c *clientapp.Controller) fyne.CanvasObject {
	panelRect := canvas.NewRectangle(ui.CurrentTheme.SidebarFill)
	panelRect.SetMinSize(fyne.NewSize(260, 0))

	title := base.H1("Albz")
	createConvBtn := base.Button("+ New",
		func() {
			err := c.CreateConversation(context.Background(), "New Conversation")
			if err != nil {
				fmt.Printf("Error creating conv: %v", err)
			}
		},
	)

	header := container.NewBorder(nil, nil, nil, createConvBtn, title)

	// Conversation list
	conversationList := container.NewVBox()
	for _, conv := range c.State.Conversations {
		convItem := ConversationCard(conv)
		conversationList.Add(convItem)
	}

	content := container.NewVBox(
		header,
		conversationList,
	)

	paddedContent := container.New(
		layout.NewCustomPaddedLayout(16, 16, 24, 24), // top, bottom, left, right
		content,
	)

	return container.NewStack(
		panelRect,
		paddedContent,
	)
}

func ConversationCard(conv sql.Conversation) fyne.CanvasObject {
	cardRect := canvas.NewRectangle(ui.CurrentTheme.ConversationCardFill)
	cardRect.StrokeColor = ui.CurrentTheme.ConversationCardStroke
	cardRect.StrokeWidth = 1
	cardRect.SetMinSize(fyne.NewSize(0, 60))

	title := base.H2(conv.Name)
	title.Alignment = fyne.TextAlignCenter

	content := container.NewCenter(title)

	return container.NewStack(
		cardRect,
		content,
	)
}
