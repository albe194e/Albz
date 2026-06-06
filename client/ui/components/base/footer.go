package base

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

type FooterPanelStyle struct {
	SizeHeight float32 // Percentage
	SizeWidth  float32 // Percentage
}

func FooterPanel(content fyne.CanvasObject, style FooterPanelStyle) fyne.CanvasObject {
	// Normalize style values
	if style.SizeHeight <= 0 {
		style.SizeHeight = 0.1
	}
	if style.SizeHeight > 1 {
		style.SizeHeight = 1
	}

	if style.SizeWidth <= 0 {
		style.SizeWidth = 1
	}
	if style.SizeWidth > 1 {
		style.SizeWidth = 1
	}

	return container.New(
		&footerPanelLayout{
			widthRatio:  style.SizeWidth,
			heightRatio: style.SizeHeight,
		},
		content,
	)
}

type footerPanelLayout struct {
	widthRatio  float32
	heightRatio float32
}

func (l *footerPanelLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) == 0 {
		return
	}

	footerSize := fyne.NewSize(size.Width*l.widthRatio, size.Height*l.heightRatio)

	for _, object := range objects {
		object.Move(fyne.NewPos(0, size.Height-footerSize.Height))
		object.Resize(footerSize)
	}
}

func (l *footerPanelLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) == 0 {
		return fyne.NewSize(0, 0)
	}

	contentMinSize := objects[0].MinSize()
	return fyne.NewSize(
		contentMinSize.Width/l.widthRatio,
		contentMinSize.Height/l.heightRatio,
	)
}
