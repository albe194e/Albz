package pages

import (
	"context"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"

	ui "github.com/albe194e/albz/client/ui"
	"github.com/albe194e/albz/client/ui/components/base"
)

func init() {
	ui.RegisterPageRenderer(ui.Login, LoginPage)
}

func LoginPage(router *ui.Router) fyne.CanvasObject {
	usernameInput := base.TextField("Username", "Enter username")
	passwordInput := base.PasswordField("Password")

	loginButton := base.Button("Login", func() {
		if err := router.Controller.Login(
			context.Background(),
			usernameInput.Entry.Text,
			passwordInput.Entry.Text,
		); err != nil {
			fmt.Println("Login failed:", err)
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
		base.H1("Welcome back", base.TextStyle{
			Color: ui.CurrentTheme.PrimaryText,
			Bold:  true,
		}),
		base.Text("Sign in to reconnect with your local conversations.", base.TextStyle{
			Color: ui.CurrentTheme.SecondaryText,
		}),
		usernameInput.View,
		passwordInput.View,
		loginButton,
		backButton,
	)

	return base.Centered(base.Card(content), 24)
}
