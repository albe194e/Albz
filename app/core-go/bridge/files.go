package bridge

import "fmt"

func (b *Bridge) SaveProfileImage(data []byte, filename string) (string, error) {
	if b == nil || b.service == nil || b.service.FileHandler == nil {
		return "", fmt.Errorf("file handler is not initialized")
	}

	return b.service.FileHandler.SaveImageToStorage(data, filename)
}
