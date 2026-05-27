package base

import "fyne.io/fyne/v2/widget"

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
