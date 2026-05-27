package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"

	clientapp "github.com/albe194e/albz/client/app"
	"github.com/albe194e/albz/client/input"
)

type pageRenderer func(*Router) fyne.CanvasObject

var pageRenderers = map[Page]pageRenderer{}

type Router struct {
	Window     fyne.Window
	Controller *clientapp.Controller
	UserInput  *input.Input
	State      *UIState
}

func NewRouter(window fyne.Window, controller *clientapp.Controller, userInput *input.Input, state *UIState) *Router {
	return &Router{
		Window:     window,
		Controller: controller,
		UserInput:  userInput,
		State:      state,
	}
}

func RegisterPageRenderer(page Page, renderer pageRenderer) {
	if renderer == nil {
		panic(fmt.Sprintf("ui: nil renderer registered for page %d", page))
	}
	if _, exists := pageRenderers[page]; exists {
		panic(fmt.Sprintf("ui: duplicate renderer registered for page %d", page))
	}

	pageRenderers[page] = renderer
}

func (r *Router) NavigateTo(page Page) {
	r.State.Page = page

	renderer, ok := pageRenderers[page]
	if !ok {
		panic(fmt.Sprintf("ui: no renderer registered for page %d", page))
	}

	content := renderer(r)
	r.Window.SetContent(pageWithBackground(content))
}

func pageWithBackground(content fyne.CanvasObject) fyne.CanvasObject {
	bg := canvas.NewRectangle(CurrentTheme.AppBackground)
	return container.NewStack(bg, content)
}
