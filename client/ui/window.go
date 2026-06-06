package ui

import (
	"context"
	"runtime"

	"fyne.io/fyne/v2"

	clientapp "github.com/albe194e/albz/client/app"
)

func Run(a fyne.App, c *clientapp.Controller, uiState *UIState, windowTitle string) {
	if windowTitle == "" {
		windowTitle = "Albz"
	}

	w := a.NewWindow(windowTitle)
	router := NewRouter(w, c, uiState)
	c.OnStateChanged = func() {
		fyne.Do(func() {
			router.InvalidateAllPages()
			router.NavigateTo(router.State.Page)
		})
	}

	uiState.Page = Landing
	if err := c.VerifySession(context.Background()); err == nil {
		uiState.Page = Chat
	}

	if runtime.GOOS != "android" {
		w.Resize(fyne.NewSize(1120, 760))
	}

	router.NavigateTo(uiState.Page)
	w.ShowAndRun()

	_ = c
	_ = uiState
}
