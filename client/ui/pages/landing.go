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
	loginButton := base.Button("Login",
		func() {
			router.NavigateTo(ui.Login)
		},
		nil,
	)

	registerButton := base.Button("Register",
		func() {
			router.NavigateTo(ui.Register)
		},
		&base.BStyle{
			Fill:   ui.CurrentTheme.Surface,
			Border: ui.CurrentTheme.Border,
			Text:   ui.CurrentTheme.PrimaryText,
		},
	)

	hero := container.NewVBox(
		base.H1("Albz", base.TextStyle{
			Size:  38,
			Color: ui.CurrentTheme.PrimaryText,
			Bold:  true,
		}),
		base.Text("A local-first chat that keeps your history on your own device.", base.TextStyle{
			Size:  16,
			Color: ui.CurrentTheme.SecondaryText,
		}),
		container.NewGridWithColumns(
			1,
			loginButton,
			registerButton,
		),
	)

	return base.Centered(base.Card(hero), 24)
}
