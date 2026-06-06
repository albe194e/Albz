package components

import (
	"context"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	app "github.com/albe194e/albz/client/app"
	"github.com/albe194e/albz/client/db/sqlc/sql"
	ui "github.com/albe194e/albz/client/ui"
	"github.com/albe194e/albz/client/ui/components/base"
)

func ConversationListPanel(r *ui.Router) fyne.CanvasObject {
	title := base.H1("Albz", base.TextStyle{
		Size:  28,
		Color: ui.CurrentTheme.PrimaryText,
		Bold:  true})

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
	}

	createConvBtn := base.Button("+", func() {
		openConversationModal(r, conversationList)
	}, nil)

	headerContent := container.NewHBox(
		title,
		layout.NewSpacer(),
		createConvBtn,
	)

	profileBtn := base.Button("Profile", func() {
		r.NavigateTo(ui.Profile)
	}, &base.BStyle{
		Fill:   ui.CurrentTheme.Surface,
		Border: ui.CurrentTheme.Border,
		Text:   ui.CurrentTheme.PrimaryText,
	})

	footerContent := container.NewHBox(
		profileBtn,
		layout.NewSpacer(),
		base.ProfilePicture(r, currentUserProfilePicturePath(r)),
	)

	minWidth := float32(300)
	rightMargin := float32(12)
	panelRadius := float32(18)
	if r.IsMobileLayout() {
		minWidth = 300
		rightMargin = 0
		panelRadius = 0
	}

	content := base.PanelWithHeaderFooter(
		container.NewVScroll(conversationList),
		base.PanelStyle{
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
		},
		container.NewVBox(headerContent, base.Seperator(nil)),
		container.NewVBox(base.Seperator(nil), footerContent),
	)

	return content
}

func currentUserProfilePicturePath(r *ui.Router) string {
	if r == nil || r.Controller == nil || r.Controller.State == nil || r.Controller.State.CurrentUser == nil {
		return ""
	}

	return r.Controller.State.CurrentUser.ProfilePictureUrl
}

func openConversationModal(r *ui.Router, conversationList *widget.List) {
	if r == nil || r.Window == nil || r.Controller == nil || r.Controller.State == nil {
		return
	}

	selectedFriends := make(map[string]sql.Friend)
	nameInput := base.Input("", "Conversation name (optional)", false)

	errorLabel := base.Text("", base.TextStyle{
		Size:  12,
		Color: ui.CurrentTheme.WarningDelayed,
	})

	friendRows := make([]fyne.CanvasObject, 0, len(r.Controller.State.Friends))
	for _, friend := range r.Controller.State.Friends {
		friendCopy := friend
		friendMeta := container.NewVBox(
			base.Text(app.FriendDisplayName(friend), base.TextStyle{
				Color: ui.CurrentTheme.PrimaryText,
				Bold:  true,
			}),
		)
		if friend.Username != "" {
			friendMeta.Add(base.Text("@"+friend.Username, base.TextStyle{
				Size:  12,
				Color: ui.CurrentTheme.SecondaryText,
			}))
		}

		check := widget.NewCheck("", func(checked bool) {
			if checked {
				selectedFriends[friendCopy.UserID] = friendCopy
				return
			}
			delete(selectedFriends, friendCopy.UserID)
		})

		friendRows = append(friendRows, container.NewBorder(nil, nil, check, nil, friendMeta))
	}

	if len(friendRows) == 0 {
		friendRows = append(friendRows, base.Text("Add a friend before creating a conversation.", base.TextStyle{
			Color: ui.CurrentTheme.MutedText,
		}))
	}

	friendList := container.NewVBox(friendRows...)
	friendScroll := container.NewVScroll(friendList)
	friendScroll.SetMinSize(fyne.NewSize(320, 220))

	var modal *widget.PopUp
	closeModal := func() {
		if modal != nil {
			modal.Hide()
		}
	}

	createConversation := func() {
		friends := make([]sql.Friend, 0, len(selectedFriends))
		for _, friend := range r.Controller.State.Friends {
			if selectedFriend, ok := selectedFriends[friend.UserID]; ok {
				friends = append(friends, selectedFriend)
			}
		}

		conversation, err := r.Controller.CreateConversationWithFriends(context.Background(), nameInput.Text, friends...)
		if err != nil {
			errorLabel.SetText(err.Error())
			return
		}

		r.State.ActiveConversationID = conversation.ID
		_ = r.State.ActiveConversationName.Set(conversation.Name)
		if _, err := r.Controller.LoadConversationMessages(context.Background(), conversation.ID); err != nil {
			fmt.Printf("Error loading messages: %v\n", err)
		}

		conversationList.Refresh()
		closeModal()

		if r.IsMobileLayout() {
			r.State.SidebarOpen = false
		}
		r.NavigateTo(r.State.Page)
	}

	content := container.NewVBox(
		base.H2("New Conversation", base.TextStyle{
			Color: ui.CurrentTheme.PrimaryText,
			Bold:  true,
		}),
		base.Text("Choose one or more friends to start a conversation.", base.TextStyle{
			Color: ui.CurrentTheme.SecondaryText,
		}),
		base.FramedEntry(nameInput),
		friendScroll,
		errorLabel,
		container.NewGridWithColumns(
			2,
			base.Button("Cancel", closeModal, &base.BStyle{
				Fill:   ui.CurrentTheme.Surface,
				Border: ui.CurrentTheme.Border,
				Text:   ui.CurrentTheme.PrimaryText,
			}),
			base.Button("Create", createConversation, nil),
		),
	)

	modal = base.Modal(r.Window, content, &base.ModalStyle{
		Focus: true,
	})
}

func ConversationCard(conv sql.Conversation) fyne.CanvasObject {
	title := base.H2(conv.Name, base.TextStyle{
		Color: ui.CurrentTheme.PrimaryText,
		Bold:  true})
	title.Alignment = fyne.TextAlignLeading

	return base.Panel(title, base.PanelStyle{
		Fill:    ui.CurrentTheme.ConversationCardFill,
		Stroke:  ui.CurrentTheme.ConversationCardStroke,
		Radius:  26,
		MinSize: fyne.NewSize(0, 72),
		Padding: &base.Padding{
			Top:    12,
			Bottom: 12,
			Left:   22,
			Right:  22,
		},
	})
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
