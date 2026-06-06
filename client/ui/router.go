package ui

import (
	"fmt"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"

	clientapp "github.com/albe194e/albz/client/app"
)

type pageRenderer func(*Router) fyne.CanvasObject

var pageRenderers = map[Page]pageRenderer{}

type Router struct {
	Window     fyne.Window
	Controller *clientapp.Controller
	State      *UIState

	rootShell    fyne.CanvasObject
	pageHost     *fyne.Container
	builtPages   map[Page]fyne.CanvasObject
	invalidPages map[Page]bool
	layoutMode   string
	mounted      bool
}

func NewRouter(window fyne.Window, controller *clientapp.Controller, state *UIState) *Router {
	return &Router{
		Window:       window,
		Controller:   controller,
		State:        state,
		builtPages:   make(map[Page]fyne.CanvasObject),
		invalidPages: make(map[Page]bool),
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
	if !r.mounted || r.layoutMode != r.currentLayoutMode() {
		r.InvalidateAllPages()
		r.mountRoot()
	}

	previousPage := r.State.Page
	r.State.Page = page
	rebuild := !r.mounted || page == previousPage || r.invalidPages[page]

	content := r.resolvePage(page, rebuild)
	r.pageHost.Objects = []fyne.CanvasObject{content}
	r.pageHost.Refresh()
}

func (r *Router) InvalidatePage(page Page) {
	r.invalidPages[page] = true
}

func (r *Router) InvalidateAllPages() {
	for page := range pageRenderers {
		r.invalidPages[page] = true
	}
}

func (r *Router) resolvePage(page Page, rebuild bool) fyne.CanvasObject {
	if !rebuild {
		if existing, ok := r.builtPages[page]; ok && existing != nil {
			return existing
		}
	}

	renderer, ok := pageRenderers[page]
	if !ok {
		panic(fmt.Sprintf("ui: no renderer registered for page %d", page))
	}

	content := renderer(r)
	r.builtPages[page] = content
	delete(r.invalidPages, page)
	return content
}

func (r *Router) mountRoot() {
	r.layoutMode = r.currentLayoutMode()
	r.pageHost = container.NewStack()
	r.rootShell = pageWithBackground(r, r.pageHost)
	r.Window.SetContent(r.rootShell)
	r.mounted = true
}

func (r *Router) currentLayoutMode() string {
	if r.IsMobileLayout() {
		return "mobile"
	}
	if r.IsCompactLayout() {
		return "compact"
	}
	return "desktop"
}

func pageWithBackground(r *Router, content fyne.CanvasObject) fyne.CanvasObject {
	bg := canvas.NewRectangle(CurrentTheme.AppBackground)

	return container.NewStack(
		bg,
		appShell(r, content),
	)
}

func appShell(r *Router, content fyne.CanvasObject) fyne.CanvasObject {
	outerMargin := float32(18)
	innerPadding := float32(18)
	radius := float32(30)

	if r.IsCompactLayout() {
		outerMargin = 12
		innerPadding = 12
		radius = 22
	}
	if r.IsMobileLayout() {
		outerMargin = 0
		innerPadding = 0
		radius = 0
	}

	shellBg := canvas.NewRectangle(CurrentTheme.AppShellFill)
	shellBg.StrokeColor = CurrentTheme.AppShellStroke
	shellBg.StrokeWidth = 1
	shellBg.CornerRadius = radius

	shellContent := content
	if innerPadding > 0 {
		shellContent = container.New(
			layout.NewCustomPaddedLayout(innerPadding, innerPadding, innerPadding, innerPadding),
			content,
		)
	}

	shell := container.NewStack(
		shellBg,
		shellContent,
	)

	if outerMargin <= 0 {
		return shell
	}

	return container.New(
		layout.NewCustomPaddedLayout(outerMargin, outerMargin, outerMargin, outerMargin),
		shell,
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
