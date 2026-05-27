package pages

import (
	"context"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	ui "github.com/albe194e/albz/client/ui"
	base "github.com/albe194e/albz/client/ui/components/base"
)

func init() {
	ui.RegisterPageRenderer(ui.Register, RegisterPage)
}

func RegisterPage(router *ui.Router) fyne.CanvasObject {
	nameInput := base.Input("Name", "Name", true)
	usernameInput := base.Input("Username", "Username", true)
	passwordInput := base.InputPassword()

	registerButton := widget.NewButton("Register", func() {
		if err := router.Controller.Register(
			context.Background(),
			nameInput.Text,
			usernameInput.Text,
			passwordInput.Text,
		); err != nil {
			// TODO: show error to user
			return
		}

		router.NavigateTo(ui.Chat)
	})

	return container.NewVBox(
		widget.NewLabel("Register"),
		nameInput,
		usernameInput,
		passwordInput,
		registerButton,
	)
}
