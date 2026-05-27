package ui

import (
	"context"

	"fyne.io/fyne/v2"
	fyneapp "fyne.io/fyne/v2/app"

	clientapp "github.com/albe194e/albz/client/app"
	"github.com/albe194e/albz/client/input"
)

func Run(c *clientapp.Controller, userInput *input.Input, uiState *UIState) {
	a := fyneapp.New()
	w := a.NewWindow("Albz")
	router := NewRouter(w, c, userInput, uiState)

	uiState.Page = Landing
	if err := c.VerifySession(context.Background()); err == nil {
		uiState.Page = Chat
	}

	router.NavigateTo(uiState.Page)

	w.Resize(fyne.NewSize(900, 600))
	w.ShowAndRun()

	_ = c
	_ = userInput
}
