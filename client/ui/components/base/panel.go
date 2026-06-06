package base

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
)

type PanelStyle struct {
	Fill    color.Color
	Stroke  color.Color
	Radius  float32
	MinSize fyne.Size
	Padding *Padding
	Margin  *Margin
}

func Panel(content fyne.CanvasObject, style PanelStyle) fyne.CanvasObject {
	if style.Padding != nil {
		content = container.New(
			layout.NewCustomPaddedLayout(
				style.Padding.Top,
				style.Padding.Bottom,
				style.Padding.Left,
				style.Padding.Right,
			),
			content,
		)
	}

	bg := canvas.NewRectangle(style.Fill)
	bg.StrokeColor = style.Stroke
	if hasVisibleStroke(style.Stroke) {
		bg.StrokeWidth = 1
	}
	bg.CornerRadius = style.Radius
	bg.SetMinSize(style.MinSize)

	panel := container.NewStack(bg, content)

	if style.Margin != nil {
		panel = container.New(
			layout.NewCustomPaddedLayout(
				style.Margin.Top,
				style.Margin.Bottom,
				style.Margin.Left,
				style.Margin.Right,
			),
			panel,
		)
	}

	return panel
}

func PanelWithHeaderFooter(content fyne.CanvasObject, style PanelStyle, headerContent, footerContent fyne.CanvasObject) fyne.CanvasObject {
	mainContent := container.NewBorder(
		headerContent,
		footerContent,
		nil,
		nil,
		content,
	)
	return Panel(mainContent, style)
}

func PanelWithHeader(content fyne.CanvasObject, style PanelStyle, headerContent fyne.CanvasObject) fyne.CanvasObject {
	mainContent := container.NewBorder(
		headerContent,
		nil,
		nil,
		nil,
		content,
	)
	return Panel(mainContent, style)
}

func PanelWithFooter(content fyne.CanvasObject, style PanelStyle, footerContent fyne.CanvasObject) fyne.CanvasObject {
	mainContent := container.NewBorder(
		nil,
		footerContent,
		nil,
		nil,
		content,
	)
	return Panel(mainContent, style)
}

func hasVisibleStroke(stroke color.Color) bool {
	if stroke == nil {
		return false
	}

	_, _, _, alpha := stroke.RGBA()
	return alpha > 0
}
