package pages

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	ui "github.com/albe194e/albz/client/ui"
	base "github.com/albe194e/albz/client/ui/components/base"
)

func init() {
	ui.RegisterPageRenderer(ui.Landing, LandingPage)
}

func LandingPage(router *ui.Router) fyne.CanvasObject {

	loginButton := base.Button("Login", func() {
		router.NavigateTo(ui.Login)
	})

	registerButton := base.Button("Register", func() {
		router.NavigateTo(ui.Register)
	})

	return container.NewVBox(
		container.NewGridWithColumns(2,
			loginButton,
			registerButton,
		),
	)
}
