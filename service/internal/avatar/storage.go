package avatar

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"

	_ "golang.org/x/image/webp"
)

const MaxBytes = 2 << 20

var allowedTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

func Dir(configDir string) string {
	return filepath.Join(configDir, "avatars")
}

func Save(configDir string, memberID int, data []byte, contentType string) (relativePath string, err error) {
	if len(data) == 0 || len(data) > MaxBytes {
		return "", fmt.Errorf("image must be between 1 byte and 2 MB")
	}

	contentType = normalizeContentType(contentType)
	if _, ok := allowedTypes[contentType]; !ok {
		if detected := detectContentType(data); detected != "" {
			contentType = detected
		}
	}
	ext, ok := allowedTypes[contentType]
	if !ok {
		return "", fmt.Errorf("unsupported image type %q (use JPEG, PNG, or WebP)", contentType)
	}
	if _, err := decodeConfig(data); err != nil {
		return "", fmt.Errorf("invalid image data: %w", err)
	}

	dir := Dir(configDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	for _, e := range allowedTypes {
		_ = os.Remove(filepath.Join(dir, fmt.Sprintf("%d%s", memberID, e)))
	}
	filename := fmt.Sprintf("%d%s", memberID, ext)
	full := filepath.Join(dir, filename)
	if err := os.WriteFile(full, data, 0o644); err != nil {
		return "", err
	}
	return filename, nil
}

func Delete(configDir string, memberID int) error {
	dir := Dir(configDir)
	for _, ext := range allowedTypes {
		_ = os.Remove(filepath.Join(dir, fmt.Sprintf("%d%s", memberID, ext)))
	}
	return nil
}

func Path(configDir, relativePath string) string {
	if relativePath == "" {
		return ""
	}
	return filepath.Join(Dir(configDir), filepath.Base(relativePath))
}

func normalizeContentType(contentType string) string {
	contentType = strings.ToLower(strings.TrimSpace(contentType))
	switch contentType {
	case "image/jpg", "image/pjpeg", "image/x-citrix-jpeg":
		return "image/jpeg"
	case "image/x-png":
		return "image/png"
	default:
		return contentType
	}
}

func detectContentType(data []byte) string {
	if len(data) >= 3 && data[0] == 0xff && data[1] == 0xd8 && data[2] == 0xff {
		return "image/jpeg"
	}
	if len(data) >= 8 && bytes.Equal(data[:8], []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a}) {
		return "image/png"
	}
	if len(data) >= 12 && bytes.Equal(data[:4], []byte("RIFF")) && bytes.Equal(data[8:12], []byte("WEBP")) {
		return "image/webp"
	}
	return ""
}

func decodeConfig(data []byte) (image.Config, error) {
	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return image.Config{}, err
	}
	return cfg, nil
}
