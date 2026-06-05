package pages

import (
	"context"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"

	ui "github.com/albe194e/albz/client/ui"
	base "github.com/albe194e/albz/client/ui/components/base"
)

func init() {
	ui.RegisterPageRenderer(ui.Register, RegisterPage)
}

func RegisterPage(router *ui.Router) fyne.CanvasObject {
	nameInput := base.TextField("Name", "Display name")
	usernameInput := base.TextField("Username", "Choose a username")
	passwordInput := base.PasswordField("Password")

	registerButton := base.Button("Create Account", func() {
		if err := router.Controller.Register(
			context.Background(),
			nameInput.Entry.Text,
			usernameInput.Entry.Text,
			passwordInput.Entry.Text,
		); err != nil {
			fmt.Println("Register failed:", err)
			return
		}

		router.NavigateTo(ui.Chat)
	}, nil)

	backButton := base.Button("Back", func() {
		router.NavigateTo(ui.Landing)
	}, &base.BStyle{
		Fill:   ui.CurrentTheme.Surface,
		Border: ui.CurrentTheme.Border,
		Text:   ui.CurrentTheme.SecondaryText,
	})

	content := container.NewVBox(
		base.H1("Create account", base.TextStyle{
			Color: ui.CurrentTheme.PrimaryText,
			Bold:  true,
		}),
		base.Text("Set up a local-first identity and start building your network.", base.TextStyle{
			Color: ui.CurrentTheme.SecondaryText,
		}),
		nameInput.View,
		usernameInput.View,
		passwordInput.View,
		registerButton,
		backButton,
	)

	return base.Centered(base.Card(content), 24)
}
