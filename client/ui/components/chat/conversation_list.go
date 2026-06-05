package components

import (
	"context"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/albe194e/albz/client/db/sqlc/sql"
	ui "github.com/albe194e/albz/client/ui"
	"github.com/albe194e/albz/client/ui/components/base"
)

func ConversationListPanel(r *ui.Router) fyne.CanvasObject {
	title := base.H1("Albz", base.TextStyle{
		Size:  28,
		Color: ui.CurrentTheme.PrimaryText,
		Bold:  true})

	subtitle := base.Text("Local-first conversations", base.TextStyle{
		Size:  13,
		Color: ui.CurrentTheme.SecondaryText,
	})

	conversationList := widget.NewList(
		func() int {
			return len(r.Controller.State.Conversations)
		},
		func() fyne.CanvasObject {
			return ConversationCard(sql.Conversation{})
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			conv := r.Controller.State.Conversations[id]
			updateConversationCard(item, conv, r.State.ActiveConversationID == conv.ID)
		},
	)
	conversationList.HideSeparators = true
	conversationList.OnSelected = func(id widget.ListItemID) {
		conv := r.Controller.State.Conversations[id]
		r.State.ActiveConversationID = conv.ID
		_ = r.State.ActiveConversationName.Set(conv.Name)
		if _, err := r.Controller.LoadConversationMessages(context.Background(), conv.ID); err != nil {
			fmt.Printf("Error loading messages: %v\n", err)
		}
		if r.IsMobileLayout() {
			r.State.SidebarOpen = false
			r.NavigateTo(r.State.Page)
		}
		conversationList.Refresh()
	}

	// TODO: Create modal for creating new conversation
	createConvBtnFunc := func() {
		err := r.Controller.CreateConversation(context.Background(), "New Conversation")
		if err != nil {
			fmt.Printf("Error creating conv: %v", err)
			return
		}

		conversationList.Refresh()
	}

	createConvBtn := base.Button("+ New",
		createConvBtnFunc,
		nil)

	headerText := container.NewVBox(title, subtitle)

	header := container.NewBorder(nil, nil, nil, createConvBtn, headerText)
	footer := container.NewBorder(nil, nil, nil, nil, base.Button("Profile", func() {
		r.NavigateTo(ui.Profile)
	}, &base.BStyle{
		Fill:   ui.CurrentTheme.Surface,
		Border: ui.CurrentTheme.Border,
		Text:   ui.CurrentTheme.PrimaryText,
	}))

	content := container.NewBorder(
		header,
		footer,
		nil,
		nil,
		conversationList,
	)

	minWidth := float32(300)
	rightMargin := float32(12)
	panelRadius := float32(18)
	if r.IsMobileLayout() {
		minWidth = 300
		rightMargin = 0
		panelRadius = 0
	}

	return base.Panel(content, base.PanelStyle{
		Fill:    ui.CurrentTheme.SidebarFill,
		Stroke:  ui.CurrentTheme.SidebarStroke,
		Radius:  panelRadius,
		MinSize: fyne.NewSize(minWidth, 0),
		Padding: &base.Padding{
			Top:    16,
			Bottom: 16,
			Left:   18,
			Right:  18,
		},
		Margin: &base.Margin{
			Right: rightMargin,
		},
	})
}

func ConversationCard(conv sql.Conversation) fyne.CanvasObject {
	cardRect := canvas.NewRectangle(ui.CurrentTheme.ConversationCardFill)
	cardRect.StrokeColor = ui.CurrentTheme.ConversationCardStroke
	cardRect.StrokeWidth = 1
	cardRect.CornerRadius = 16
	cardRect.SetMinSize(fyne.NewSize(0, 72))

	title := base.H2(conv.Name, base.TextStyle{
		Color: ui.CurrentTheme.PrimaryText,
		Bold:  true})

	title.Alignment = fyne.TextAlignLeading

	content := container.NewPadded(title)

	return container.NewStack(
		cardRect,
		content,
	)
}

func updateConversationCard(item fyne.CanvasObject, conv sql.Conversation, active bool) {
	card := item.(*fyne.Container)
	cardRect := card.Objects[0].(*canvas.Rectangle)
	content := card.Objects[1].(*fyne.Container)
	title := content.Objects[0].(*base.TextWidget)

	title.SetText(conv.Name)
	if active {
		cardRect.FillColor = ui.CurrentTheme.PrimaryAccent
		cardRect.StrokeColor = ui.CurrentTheme.AccentHover
		title.Color = ui.CurrentTheme.PrimaryText
	} else {
		cardRect.FillColor = ui.CurrentTheme.ConversationCardFill
		cardRect.StrokeColor = ui.CurrentTheme.ConversationCardStroke
		title.Color = ui.CurrentTheme.PrimaryText
	}

	title.Refresh()
	cardRect.Refresh()
}
