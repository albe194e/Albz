package base

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"

	ui "github.com/albe194e/albz/client/ui"
)

type SeperatorStyle struct {
	Width     float32 // Percentage
	Thickness float32
	Margin    *Margin
}

func Seperator(style *SeperatorStyle) fyne.CanvasObject {
	if style == nil {
		style = &SeperatorStyle{}
	}

	setSeperatorDefaults(style)

	line := canvas.NewRectangle(ui.CurrentTheme.Divider)
	line.SetMinSize(fyne.NewSize(0, style.Thickness))

	return container.New(
		layout.NewCustomPaddedLayout(
			style.Margin.Top,
			style.Margin.Bottom,
			style.Margin.Left,
			style.Margin.Right,
		),
		container.New(
			&separatorLayout{
				widthRatio: style.Width,
				thickness:  style.Thickness,
			},
			line,
		),
	)
}

func setSeperatorDefaults(style *SeperatorStyle) {
	if style.Thickness <= 0 {
		style.Thickness = 1
	}

	if style.Width <= 0 {
		style.Width = 1
	}
	if style.Width > 1 {
		style.Width = 1
	}

	if style.Margin == nil {
		style.Margin = &Margin{}
	}
	if style.Margin.Top <= 0 {
		style.Margin.Top = 10
	}
	if style.Margin.Bottom <= 0 {
		style.Margin.Bottom = 10
	}
}

type separatorLayout struct {
	widthRatio float32
	thickness  float32
}

func (l *separatorLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) == 0 {
		return
	}

	lineWidth := size.Width * l.widthRatio
	lineHeight := l.thickness
	x := (size.Width - lineWidth) / 2
	y := (size.Height - lineHeight) / 2

	for _, object := range objects {
		object.Move(fyne.NewPos(x, y))
		object.Resize(fyne.NewSize(lineWidth, lineHeight))
	}
}

func (l *separatorLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) == 0 {
		return fyne.NewSize(0, l.thickness)
	}

	return fyne.NewSize(0, l.thickness)
}
