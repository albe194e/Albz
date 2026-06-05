package pages

import (
	"context"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"

	app "github.com/albe194e/albz/client/app"
	ui "github.com/albe194e/albz/client/ui"
	"github.com/albe194e/albz/client/ui/components/base"
)

func init() {
	ui.RegisterPageRenderer(ui.Profile, ProfilePage)
}

func ProfilePage(router *ui.Router) fyne.CanvasObject {
	backBtn := base.Button("Back", func() {
		router.NavigateTo(ui.Chat)
	}, &base.BStyle{
		Fill:   ui.CurrentTheme.Surface,
		Border: ui.CurrentTheme.Border,
		Text:   ui.CurrentTheme.PrimaryText,
	})

	titleLabel := base.H1("Profile", base.TextStyle{
		Size:  28,
		Color: ui.CurrentTheme.PrimaryText,
		Bold:  true,
	})

	friendCodeLabel := base.Text("Friend code unavailable", base.TextStyle{
		Color: ui.CurrentTheme.PrimaryText,
		Bold:  true,
	})
	if router.Controller.State.CurrentUser != nil {
		friendCodeLabel.SetText("Your friend code: " + router.Controller.State.CurrentUser.FriendCode)
	}

	friendCodeInput := base.Input("Send friend request", "Enter a friend's code", true)
	friendCodeInput.SetText("ALBZ-")

	sendRequestBtn := base.Button("Send Request", func() {
		err := router.Controller.SendFriendRequest(context.Background(), friendCodeInput.Text)
		if err != nil {
			fmt.Printf("Error sending friend request: %v\n", err)
			return
		}
		friendCodeInput.SetText("")
	}, nil)

	requestCards := []fyne.CanvasObject{}
	if len(router.Controller.State.FriendRequests) == 0 {
		requestCards = append(requestCards, base.Text("No pending friend requests", base.TextStyle{
			Color: ui.CurrentTheme.MutedText,
		}))
	} else {
		for _, request := range router.Controller.State.FriendRequests {
			requestCopy := request
			requestMeta := container.NewVBox(
				base.Text(app.FriendRequestDisplayName(request), base.TextStyle{
					Color: ui.CurrentTheme.PrimaryText,
					Bold:  true,
				}),
			)
			if request.Username != "" {
				requestMeta.Add(base.Text("@"+request.Username, base.TextStyle{
					Size:  12,
					Color: ui.CurrentTheme.SecondaryText,
				}))
			}
			requestCards = append(requestCards, base.Card(container.NewVBox(
				requestMeta,
				container.NewGridWithColumns(
					2,
					base.Button("Accept", func() {
						err := router.Controller.AcceptFriendRequest(context.Background(), requestCopy.FromUserID)
						if err != nil {
							fmt.Printf("Error accepting friend request: %v\n", err)
						}
					}, nil),
					base.Button("Reject", func() {
						err := router.Controller.RejectFriendRequest(context.Background(), requestCopy.FromUserID)
						if err != nil {
							fmt.Printf("Error rejecting friend request: %v\n", err)
						}
					}, &base.BStyle{
						Fill:   ui.CurrentTheme.Surface,
						Border: ui.CurrentTheme.Border,
						Text:   ui.CurrentTheme.PrimaryText,
					}),
				),
			)))
		}
	}

	friendCards := []fyne.CanvasObject{}
	if len(router.Controller.State.Friends) == 0 {
		friendCards = append(friendCards, base.Text("No friends added yet", base.TextStyle{
			Color: ui.CurrentTheme.MutedText,
		}))
	} else {
		for _, friend := range router.Controller.State.Friends {
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
			friendCards = append(friendCards, base.Card(container.NewBorder(
				nil,
				nil,
				nil,
				base.Button("Start Chat", func() {
					conversation, err := router.Controller.CreateConversationWithFriend(context.Background(), friendCopy)
					if err != nil {
						fmt.Printf("Error creating conversation: %v\n", err)
						return
					}

					router.State.ActiveConversationID = conversation.ID
					_ = router.State.ActiveConversationName.Set(conversation.Name)
					if _, err := router.Controller.LoadConversationMessages(context.Background(), conversation.ID); err != nil {
						fmt.Printf("Error loading messages: %v\n", err)
					}
					router.NavigateTo(ui.Chat)
				}, nil),
				friendMeta,
			)))
		}
	}

	errorLabel := base.Text("", base.TextStyle{
		Size:  12,
		Color: ui.CurrentTheme.WarningDelayed,
	})
	if router.Controller.State.LastNetworkError != "" {
		errorLabel.SetText("Network: " + router.Controller.State.LastNetworkError)
	}

	header := container.NewBorder(nil, nil, backBtn, nil, titleLabel)
	shareCard := base.Card(container.NewVBox(
		base.Text("Share this code so another person can add you.", base.TextStyle{
			Color: ui.CurrentTheme.SecondaryText,
		}),
		friendCodeLabel,
		container.NewBorder(nil, nil, nil, sendRequestBtn, base.FramedEntry(friendCodeInput)),
		errorLabel,
	))

	content := container.NewVBox(
		header,
		shareCard,
		base.Section("Pending Requests", container.NewVBox(requestCards...)),
		base.Section("Friends", container.NewVBox(friendCards...)),
	)

	return base.Centered(content, 24)
}
