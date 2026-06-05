package ui

import (
	"fmt"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"

	clientapp "github.com/albe194e/albz/client/app"
)

type pageRenderer func(*Router) fyne.CanvasObject

var pageRenderers = map[Page]pageRenderer{}

type Router struct {
	Window     fyne.Window
	Controller *clientapp.Controller
	State      *UIState
}

func NewRouter(window fyne.Window, controller *clientapp.Controller, state *UIState) *Router {
	return &Router{
		Window:     window,
		Controller: controller,
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
	accent := canvas.NewRectangle(CurrentTheme.Surface)
	accent.SetMinSize(fyne.NewSize(0, 140))

	return container.NewStack(
		bg,
		container.NewBorder(accent, nil, nil, nil, nil),
		content,
	)
}

func (r *Router) IsCompactLayout() bool {
	if runtime.GOOS == "android" {
		return true
	}

	size := r.Window.Canvas().Size()
	return size.Width > 0 && size.Width < 760
}

func (r *Router) IsMobileLayout() bool {
	return runtime.GOOS == "android"
}
