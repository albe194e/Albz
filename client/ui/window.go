package ui

import (
	"context"
	"runtime"

	"fyne.io/fyne/v2"

	clientapp "github.com/albe194e/albz/client/app"
)

func Run(a fyne.App, c *clientapp.Controller, uiState *UIState) {
	w := a.NewWindow("Albz")
	router := NewRouter(w, c, uiState)
	c.OnStateChanged = func() {
		fyne.Do(func() {
			router.NavigateTo(router.State.Page)
		})
	}

	uiState.Page = Landing
	if err := c.VerifySession(context.Background()); err == nil {
		uiState.Page = Chat
	}

	router.NavigateTo(uiState.Page)

	if runtime.GOOS != "android" {
		w.Resize(fyne.NewSize(1120, 760))
	}
	w.ShowAndRun()

	_ = c
	_ = uiState
}
