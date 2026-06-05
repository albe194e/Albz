package base

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	ui "github.com/albe194e/albz/client/ui"
)

type InputField struct {
	Entry *widget.Entry
	View  fyne.CanvasObject
}

func Input(label, placeholder string, required bool) *widget.Entry {
	entry := widget.NewEntry()
	entry.SetPlaceHolder(placeholder)
	return entry
}

func InputPassword() *widget.Entry {
	entry := widget.NewEntry()
	entry.SetPlaceHolder("Enter password..")
	entry.Password = true

	return entry
}

func TextField(label, placeholder string) *InputField {
	entry := Input(label, placeholder, false)
	return newInputField(label, entry)
}

func PasswordField(label string) *InputField {
	entry := InputPassword()
	return newInputField(label, entry)
}

func FramedEntry(entry *widget.Entry) fyne.CanvasObject {
	return Panel(entry, PanelStyle{
		Fill:    ui.CurrentTheme.PanelFill,
		Stroke:  ui.CurrentTheme.PanelStroke,
		Radius:  14,
		MinSize: fyne.NewSize(0, 48),
		Padding: &Padding{
			Top:    8,
			Bottom: 8,
			Left:   12,
			Right:  12,
		},
	})
}

func newInputField(label string, entry *widget.Entry) *InputField {
	labelWidget := Text(label, TextStyle{
		Size:  12,
		Color: ui.CurrentTheme.SecondaryText,
		Bold:  true,
	})
	labelWidget.Alignment = fyne.TextAlignLeading

	return &InputField{
		Entry: entry,
		View: container.NewVBox(
			labelWidget,
			FramedEntry(entry),
		),
	}
}
