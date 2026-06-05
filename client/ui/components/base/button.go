package base

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	ui "github.com/albe194e/albz/client/ui"
)

type BStyle struct {
	Fill    color.Color
	Border  color.Color
	Text    color.Color
	Radius  float32
	MinSize fyne.Size
}

type StyledButton struct {
	widget.BaseWidget

	Label    string
	OnTapped func()
	Style    *BStyle
}

func Button(label string, onTapped func(), style *BStyle) *StyledButton {
	if style == nil {
		style = &BStyle{}
	}

	if style.Fill == nil {
		style.Fill = ui.CurrentTheme.PrimaryAccent
	}
	if style.Border == nil {
		style.Border = ui.CurrentTheme.PrimaryAccent
	}
	if style.Text == nil {
		style.Text = ui.CurrentTheme.PrimaryText
	}
	if style.Radius == 0 {
		style.Radius = ui.BtnRadius
	}
	if style.MinSize.Width == 0 && style.MinSize.Height == 0 {
		style.MinSize = fyne.NewSize(0, 44)
	}

	btn := &StyledButton{
		Label:    label,
		OnTapped: onTapped,
		Style:    style,
	}

	btn.ExtendBaseWidget(btn)
	return btn
}

func (b *StyledButton) Tapped(_ *fyne.PointEvent) {
	if b.OnTapped != nil {
		b.OnTapped()
	}
}

func (b *StyledButton) CreateRenderer() fyne.WidgetRenderer {
	bg := canvas.NewRectangle(b.Style.Fill)
	bg.StrokeColor = b.Style.Border
	bg.StrokeWidth = 1
	bg.CornerRadius = b.Style.Radius
	bg.SetMinSize(b.Style.MinSize)

	label := canvas.NewText(b.Label, b.Style.Text)
	label.TextSize = 14
	label.TextStyle = fyne.TextStyle{Bold: true}
	label.Alignment = fyne.TextAlignCenter

	content := container.NewStack(
		bg,
		container.NewCenter(label),
	)

	return widget.NewSimpleRenderer(content)
}
