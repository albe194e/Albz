package base

import "fyne.io/fyne/v2/widget"

func Button(label string, onTapped func()) *widget.Button {
	return widget.NewButton(label, onTapped)
}
