package base

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	ui "github.com/albe194e/albz/client/ui"
)

type ModalStyle struct {
	Focus   bool
	OnClose func()
}

func Modal(window fyne.Window, content fyne.CanvasObject, style *ModalStyle) *widget.PopUp {
	if style == nil {
		style = &ModalStyle{Focus: true}
	}

	panel := Panel(content, PanelStyle{
		Fill:   ui.CurrentTheme.PanelFill,
		Stroke: ui.CurrentTheme.PanelStroke,
		Radius: 18,
		Padding: &Padding{
			Top:    18,
			Bottom: 18,
			Left:   18,
			Right:  18,
		},
	})

	var popup *widget.PopUp
	closeModal := func() {
		if popup != nil {
			popup.Hide()
		}
		if style.OnClose != nil {
			style.OnClose()
		}
	}

	overlay := newModalOverlay(style.Focus, closeModal)

	centeredPanel := container.New(
		layout.NewCustomPaddedLayout(24, 24, 24, 24),
		container.NewCenter(panel),
	)

	modalRoot := container.NewStack(
		overlay,
		centeredPanel,
	)

	popup = widget.NewPopUp(modalRoot, window.Canvas())
	popup.Resize(window.Canvas().Size())
	popup.Move(fyne.NewPos(0, 0))
	popup.Show()

	return popup
}

type modalOverlay struct {
	widget.BaseWidget
	onTapped func()
	fill     color.Color
}

func newModalOverlay(focus bool, onTapped func()) *modalOverlay {
	fill := color.NRGBA{A: 0}
	if focus {
		fill = ui.CurrentTheme.AppBackground
		fill.A = 230
	}

	overlay := &modalOverlay{
		onTapped: onTapped,
		fill:     fill,
	}
	overlay.ExtendBaseWidget(overlay)
	return overlay
}

func (o *modalOverlay) Tapped(_ *fyne.PointEvent) {
	if o.onTapped != nil {
		o.onTapped()
	}
}

func (o *modalOverlay) CreateRenderer() fyne.WidgetRenderer {
	bg := canvas.NewRectangle(o.fill)
	return widget.NewSimpleRenderer(bg)
}
