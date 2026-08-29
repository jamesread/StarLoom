package avatar

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "golang.org/x/image/webp"
)

const MaxBytes = 2 << 20

var allowedTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

type Entry struct {
	Filename  string
	IsCurrent bool
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
	if err := archiveCurrentSlot(dir, memberID); err != nil {
		return "", err
	}
	filename := fmt.Sprintf("%d%s", memberID, ext)
	full := filepath.Join(dir, filename)
	if err := os.WriteFile(full, data, 0o644); err != nil {
		return "", err
	}
	return filename, nil
}

func Select(configDir string, memberID int, filename string) (relativePath string, err error) {
	base := filepath.Base(strings.TrimSpace(filename))
	if base == "" || !BelongsToMember(memberID, base) {
		return "", fmt.Errorf("invalid avatar filename")
	}
	dir := Dir(configDir)
	src := filepath.Join(dir, base)
	if _, err := os.Stat(src); err != nil {
		return "", fmt.Errorf("avatar not found")
	}
	if isCurrentSlot(memberID, base) {
		return base, nil
	}
	if err := archiveCurrentSlot(dir, memberID); err != nil {
		return "", err
	}
	data, err := os.ReadFile(src)
	if err != nil {
		return "", err
	}
	ext := filepath.Ext(base)
	destName := fmt.Sprintf("%d%s", memberID, ext)
	dest := filepath.Join(dir, destName)
	if err := os.WriteFile(dest, data, 0o644); err != nil {
		return "", err
	}
	return destName, nil
}

func List(configDir string, memberID int, currentPath string) ([]Entry, error) {
	dir := Dir(configDir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	currentBase := filepath.Base(currentPath)
	out := make([]Entry, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !BelongsToMember(memberID, name) {
			continue
		}
		out = append(out, Entry{
			Filename:  name,
			IsCurrent: name == currentBase,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].IsCurrent != out[j].IsCurrent {
			return out[i].IsCurrent
		}
		return out[i].Filename > out[j].Filename
	})
	return out, nil
}

func DeleteCurrent(configDir string, memberID int) error {
	dir := Dir(configDir)
	for _, ext := range allowedTypes {
		_ = os.Remove(filepath.Join(dir, fmt.Sprintf("%d%s", memberID, ext)))
	}
	return nil
}

func DeleteAll(configDir string, memberID int) error {
	dir := Dir(configDir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	prefix := fmt.Sprintf("%d", memberID)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasPrefix(name, prefix) {
			continue
		}
		suffix := name[len(prefix):]
		if suffix == "" {
			continue
		}
		if suffix[0] != '.' && suffix[0] != '-' {
			continue
		}
		if !hasAllowedExt(name) {
			continue
		}
		_ = os.Remove(filepath.Join(dir, name))
	}
	return nil
}

// Delete removes the current avatar slot for a member. Prefer DeleteCurrent.
func Delete(configDir string, memberID int) error {
	return DeleteCurrent(configDir, memberID)
}

func Path(configDir, relativePath string) string {
	if relativePath == "" {
		return ""
	}
	return filepath.Join(Dir(configDir), filepath.Base(relativePath))
}

func BelongsToMember(memberID int, filename string) bool {
	if strings.Contains(filename, "/") || strings.Contains(filename, "\\") || filename != filepath.Base(filename) {
		return false
	}
	base := filepath.Base(filename)
	prefix := fmt.Sprintf("%d", memberID)
	if !strings.HasPrefix(base, prefix) {
		return false
	}
	suffix := base[len(prefix):]
	if suffix == "" {
		return false
	}
	if suffix[0] == '.' {
		return hasAllowedExt(base)
	}
	if suffix[0] != '-' {
		return false
	}
	return hasAllowedExt(base)
}

func archiveCurrentSlot(dir string, memberID int) error {
	for _, ext := range allowedTypes {
		current := filepath.Join(dir, fmt.Sprintf("%d%s", memberID, ext))
		if _, err := os.Stat(current); err != nil {
			continue
		}
		archive := filepath.Join(dir, fmt.Sprintf("%d-%d%s", memberID, time.Now().Unix(), ext))
		if err := os.Rename(current, archive); err != nil {
			return err
		}
	}
	return nil
}

func isCurrentSlot(memberID int, filename string) bool {
	base := filepath.Base(filename)
	for _, ext := range allowedTypes {
		if base == fmt.Sprintf("%d%s", memberID, ext) {
			return true
		}
	}
	return false
}

func hasAllowedExt(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	for _, allowed := range allowedTypes {
		if ext == allowed {
			return true
		}
	}
	return false
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
