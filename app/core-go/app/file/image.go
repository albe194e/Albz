package file

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

type LoadedImage struct {
	File   LoadedFile
	Image  image.Image
	Format string
	Width  int
	Height int
}

func (h *Handler) SaveImageToStorage(data []byte, filename string) (string, error) {
	if len(data) == 0 {
		return "", fmt.Errorf("image data is required")
	}

	_, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("decode image: %w", err)
	}

	imageDirectory, err := h.imageDirectory()
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(imageDirectory, 0755); err != nil {
		return "", fmt.Errorf("create image directory: %w", err)
	}

	extension := strings.ToLower(filepath.Ext(strings.TrimSpace(filename)))
	if extension == "" {
		extension = "." + strings.ToLower(format)
		if extension == ".jpeg" {
			extension = ".jpg"
		}
	}

	path := filepath.Join(imageDirectory, uuid.NewString()+extension)
	err = os.WriteFile(path, data, 0644)
	if err != nil {
		return "", err
	}
	return path, nil
}

func (h *Handler) LoadImageFromPath(path string) (*LoadedImage, error) {
	trimmedPath := strings.TrimSpace(path)
	if trimmedPath == "" {
		return nil, fmt.Errorf("image path is required")
	}

	if h != nil {
		h.cacheMu.RLock()
		cachedImage, ok := h.imageCache[trimmedPath]
		h.cacheMu.RUnlock()
		if ok {
			return cachedImage, nil
		}
	}

	file, err := loadFileFromPath(path)
	if err != nil {
		return nil, err
	}

	img, format, err := image.Decode(bytes.NewReader(file.Data))
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}

	bounds := img.Bounds()

	loadedImage := &LoadedImage{
		File:   *file,
		Image:  img,
		Format: format,
		Width:  bounds.Dx(),
		Height: bounds.Dy(),
	}

	if h != nil {
		h.cacheMu.Lock()
		h.imageCache[trimmedPath] = loadedImage
		h.cacheMu.Unlock()
	}

	return loadedImage, nil
}

func (h *Handler) CircularImage(src image.Image) image.Image {
	return circularImage(src)
}

func (h *Handler) LoadCircularImageFromPath(path string) (image.Image, error) {
	trimmedPath := strings.TrimSpace(path)
	if trimmedPath == "" {
		return nil, fmt.Errorf("image path is required")
	}

	if h != nil {
		h.cacheMu.RLock()
		cachedImage, ok := h.circularImageCache[trimmedPath]
		h.cacheMu.RUnlock()
		if ok {
			return cachedImage, nil
		}
	}

	loadedImage, err := h.LoadImageFromPath(trimmedPath)
	if err != nil {
		return nil, err
	}

	circular := circularImage(loadedImage.Image)

	if h != nil {
		h.cacheMu.Lock()
		h.circularImageCache[trimmedPath] = circular
		h.cacheMu.Unlock()
	}

	return circular, nil
}

func cropSquare(src image.Image) image.Image {
	bounds := src.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	size := min(width, height)

	offsetX := bounds.Min.X + (width-size)/2
	offsetY := bounds.Min.Y + (height-size)/2

	dst := image.NewNRGBA(image.Rect(0, 0, size, size))
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dst.Set(x, y, src.At(offsetX+x, offsetY+y))
		}
	}

	return dst
}

func MaskCircle(src image.Image) image.Image {
	bounds := src.Bounds()
	size := min(bounds.Dx(), bounds.Dy())
	dst := image.NewNRGBA(image.Rect(0, 0, size, size))

	radius := float64(size) / 2
	center := radius - 0.5

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x) - center
			dy := float64(y) - center
			if dx*dx+dy*dy > radius*radius {
				dst.Set(x, y, color.NRGBA{})
				continue
			}

			dst.Set(x, y, src.At(bounds.Min.X+x, bounds.Min.Y+y))
		}
	}

	return dst
}

func circularImage(src image.Image) image.Image {
	return MaskCircle(cropSquare(src))
}

func (h *Handler) imageDirectory() (string, error) {
	if h == nil {
		return "", fmt.Errorf("file handler is required")
	}

	if strings.TrimSpace(h.LocalDataDir) == "" {
		return "", fmt.Errorf("local data directory is required")
	}

	return filepath.Join(h.LocalDataDir, "images"), nil
}
