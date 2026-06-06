package file

import (
	"image"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Handler struct {
	LocalDataDir       string
	cacheMu            sync.RWMutex
	imageCache         map[string]*LoadedImage
	circularImageCache map[string]image.Image
}

type LoadedFile struct {
	Data        []byte
	Path        string
	ContentType string
	Size        int64
}

type UploadFileToStorage func(bytes []byte, filename string) (string, error)

func NewHandler(localDataDir string) (*Handler, error) {
	trimmedPath := strings.TrimSpace(localDataDir)
	if trimmedPath == "" {
		return nil, os.ErrInvalid
	}

	absolutePath, err := filepath.Abs(trimmedPath)
	if err != nil {
		return nil, err
	}

	return &Handler{
		LocalDataDir:       absolutePath,
		imageCache:         make(map[string]*LoadedImage),
		circularImageCache: make(map[string]image.Image),
	}, nil
}

// Should not be exposed
func loadFileFromPath(path string) (*LoadedFile, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	contentType := http.DetectContentType(data)

	return &LoadedFile{
		Data:        data,
		Path:        path,
		ContentType: contentType,
		Size:        info.Size(),
	}, nil
}
