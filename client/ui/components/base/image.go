package base

import (
	"bytes"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"

	ui "github.com/albe194e/albz/client/ui"
)

type ImageStyle struct {
	SizeWidth, SizeHeight int
	FillMode              canvas.ImageFill
	PanelStyle            PanelStyle
}

func Image(r *ui.Router, path string, style ImageStyle) fyne.CanvasObject {
	var img *canvas.Image
	if r != nil &&
		r.Controller != nil &&
		r.Controller.FileHandler != nil &&
		strings.TrimSpace(path) != "" {
		loadedImage, err := r.Controller.FileHandler.LoadImageFromPath(path)
		if err == nil {
			img = canvas.NewImageFromImage(loadedImage.Image)
		}
	}
	if img == nil {
		img = canvas.NewImageFromResource(theme.FyneLogo())
	}

	if style.FillMode != 0 {
		img.FillMode = style.FillMode
	} else {
		img.FillMode = canvas.ImageFillContain
	}

	if style.SizeWidth > 0 && style.SizeHeight > 0 {
		img.SetMinSize(fyne.NewSize(
			float32(style.SizeWidth),
			float32(style.SizeHeight),
		))
	}

	return Panel(img, style.PanelStyle)
}

func ProfilePicture(r *ui.Router, path string) fyne.CanvasObject {
	var handlerImage image.Image
	if r != nil &&
		r.Controller != nil &&
		r.Controller.FileHandler != nil &&
		strings.TrimSpace(path) != "" {
		loadedImage, err := r.Controller.FileHandler.LoadCircularImageFromPath(path)
		if err == nil {
			handlerImage = loadedImage
		}
	}
	if handlerImage == nil {
		resource := theme.FyneLogo()
		decoded, _, err := image.Decode(bytes.NewReader(resource.Content()))
		if err != nil {
			return Image(r, "", ImageStyle{
				SizeWidth:  40,
				SizeHeight: 40,
				FillMode:   canvas.ImageFillContain,
			})
		}
		if r != nil && r.Controller != nil && r.Controller.FileHandler != nil {
			handlerImage = r.Controller.FileHandler.CircularImage(decoded)
		} else {
			handlerImage = decoded
		}
	}

	img := canvas.NewImageFromImage(handlerImage)
	img.FillMode = canvas.ImageFillContain
	img.SetMinSize(fyne.NewSize(40, 40))

	return img
}
