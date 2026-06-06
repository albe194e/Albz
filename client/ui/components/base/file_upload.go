package base

import (
	"fmt"
	"io"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"

	ui "github.com/albe194e/albz/client/ui"
)

type FileUploadField struct {
	SelectedFileName string
	View             fyne.CanvasObject

	fileNameLabel *TextWidget
	errorLabel    *TextWidget
}

func ImageUploadField(window fyne.Window, label string, onSelected func(data []byte, filename string) error) *FileUploadField {
	if window == nil {
		panic("base: image upload requires a window")
	}

	labelWidget := Text(label, TextStyle{
		Size:  12,
		Color: ui.CurrentTheme.SecondaryText,
		Bold:  true,
	})
	labelWidget.Alignment = fyne.TextAlignLeading

	fileNameLabel := Text("No image selected", TextStyle{
		Color: ui.CurrentTheme.SecondaryText,
	})
	fileNameLabel.Alignment = fyne.TextAlignLeading

	errorLabel := Text("", TextStyle{
		Size:  12,
		Color: ui.CurrentTheme.WarningDelayed,
	})
	errorLabel.Alignment = fyne.TextAlignLeading

	field := &FileUploadField{
		fileNameLabel: fileNameLabel,
		errorLabel:    errorLabel,
	}

	openButton := Button("Choose Image", func() {
		open := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				field.setError(err)
				return
			}
			if reader == nil {
				return
			}
			defer reader.Close()

			data, err := io.ReadAll(reader)
			if err != nil {
				field.setError(err)
				return
			}

			filename := reader.URI().Name()
			if filename == "" {
				field.setError(fmt.Errorf("selected image is missing a filename"))
				return
			}

			if onSelected != nil {
				if err := onSelected(data, filename); err != nil {
					field.setError(err)
					return
				}
			}

			field.SelectedFileName = filename
			field.fileNameLabel.SetText(filename)
			field.errorLabel.SetText("")
		}, window)
		open.SetFilter(storage.NewExtensionFileFilter([]string{".png", ".jpg", ".jpeg", ".gif"}))
		open.Show()
	}, &BStyle{
		Fill:   ui.CurrentTheme.Surface,
		Border: ui.CurrentTheme.Border,
		Text:   ui.CurrentTheme.PrimaryText,
	})

	body := Panel(container.NewBorder(
		nil,
		nil,
		nil,
		openButton,
		fileNameLabel,
	), PanelStyle{
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

	field.View = container.NewVBox(
		labelWidget,
		body,
		errorLabel,
	)

	return field
}

func (f *FileUploadField) setError(err error) {
	if f == nil || err == nil {
		return
	}

	f.errorLabel.SetText(err.Error())
}
