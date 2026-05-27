package pages

import (
	"context"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	ui "github.com/albe194e/albz/client/ui"
	"github.com/albe194e/albz/client/ui/components/base"
)

func init() {
	ui.RegisterPageRenderer(ui.Login, LoginPage)
}

func LoginPage(router *ui.Router) fyne.CanvasObject {
	usernameInput := base.Input("Username", "Enter username..", true)
	passwordInput := base.InputPassword()

	loginButton := widget.NewButton("Login", func() {
		if err := router.Controller.Login(
			context.Background(),
			usernameInput.Text,
			passwordInput.Text,
		); err != nil {
			// TODO: show error to user
			fmt.Println("Login failed:", err)
			return
		}

		router.NavigateTo(ui.Chat)
	})

	return container.NewVBox(
		widget.NewLabel("Login"),
		usernameInput,
		passwordInput,
		loginButton,
	)
}
