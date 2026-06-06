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

func ButtonLogout(r *ui.Router) *StyledButton {
	return LogoutConfirmButton(r)
}

func LogoutConfirmButton(r *ui.Router) *StyledButton {
	return Button("Logout", func() {
		if r == nil || r.Window == nil {
			return
		}

		var modal *widget.PopUp
		closeModal := func() {
			if modal != nil {
				modal.Hide()
			}
		}

		modalContent := container.NewVBox(
			H2("Log out?", TextStyle{
				Color: ui.CurrentTheme.PrimaryText,
				Bold:  true,
			}),
			Text("Are you sure you want to log out?", TextStyle{
				Color: ui.CurrentTheme.SecondaryText,
			}),
			container.NewGridWithColumns(
				2,
				Button("Cancel", func() {
					closeModal()
				}, &BStyle{
					Fill:   ui.CurrentTheme.Surface,
					Border: ui.CurrentTheme.Border,
					Text:   ui.CurrentTheme.PrimaryText,
				}),
				Button("Logout", func() {
					closeModal()
				}, nil),
			),
		)

		modal = Modal(r.Window, modalContent, &ModalStyle{
			Focus: true,
		})
	}, &BStyle{
		Fill:   ui.CurrentTheme.Surface,
		Border: ui.CurrentTheme.Border,
		Text:   ui.CurrentTheme.PrimaryText,
	})
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

	paddedLabel := container.New(
		layout.NewCustomPaddedLayout(10, 10, 14, 14),
		container.NewCenter(label),
	)

	content := container.NewStack(
		bg,
		paddedLabel,
	)

	return widget.NewSimpleRenderer(content)
}
