package avatar

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"slices"
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
	ext, err := validateImagePayload(data, contentType)
	if err != nil {
		return "", err
	}
	dir := Dir(configDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	if err := archiveCurrentSlot(dir, memberID); err != nil {
		return "", err
	}
	filename := fmt.Sprintf("%d%s", memberID, ext)
	if err := os.WriteFile(filepath.Join(dir, filename), data, 0o644); err != nil {
		return "", err
	}
	return filename, nil
}

func validateImagePayload(data []byte, contentType string) (ext string, err error) {
	if len(data) == 0 || len(data) > MaxBytes {
		return "", fmt.Errorf("image must be between 1 byte and 2 MB")
	}
	contentType = resolveContentType(data, contentType)
	ext, ok := allowedTypes[contentType]
	if !ok {
		return "", fmt.Errorf("unsupported image type %q (use JPEG, PNG, or WebP)", contentType)
	}
	if _, err := decodeConfig(data); err != nil {
		return "", fmt.Errorf("invalid image data: %w", err)
	}
	return ext, nil
}

func resolveContentType(data []byte, contentType string) string {
	contentType = normalizeContentType(contentType)
	if _, ok := allowedTypes[contentType]; ok {
		return contentType
	}
	if detected := detectContentType(data); detected != "" {
		return detected
	}
	return contentType
}

func Select(configDir string, memberID int, filename string) (relativePath string, err error) {
	base := filepath.Base(strings.TrimSpace(filename))
	if base == "" || !BelongsToMember(memberID, base) {
		return "", fmt.Errorf("invalid avatar filename")
	}
	dir := Dir(configDir)
	src := filepath.Join(dir, base)
	if _, err = os.Stat(src); err != nil {
		return "", fmt.Errorf("avatar not found")
	}
	if isCurrentSlot(memberID, base) {
		return base, nil
	}
	return promoteAvatar(dir, memberID, src, base)
}

func promoteAvatar(dir string, memberID int, src, base string) (string, error) {
	if err := archiveCurrentSlot(dir, memberID); err != nil {
		return "", err
	}
	data, err := os.ReadFile(src)
	if err != nil {
		return "", err
	}
	destName := fmt.Sprintf("%d%s", memberID, filepath.Ext(base))
	if err := os.WriteFile(filepath.Join(dir, destName), data, 0o644); err != nil {
		return "", err
	}
	return destName, nil
}

func List(configDir string, memberID int, currentPath string) ([]Entry, error) {
	entries, err := readAvatarDir(Dir(configDir))
	if err != nil || entries == nil {
		return nil, err
	}
	out := memberAvatarEntries(entries, memberID, filepath.Base(currentPath))
	sortAvatarEntries(out)
	return out, nil
}

func readAvatarDir(dir string) ([]os.DirEntry, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	return entries, nil
}

func memberAvatarEntries(entries []os.DirEntry, memberID int, currentBase string) []Entry {
	out := make([]Entry, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !BelongsToMember(memberID, name) {
			continue
		}
		out = append(out, Entry{Filename: name, IsCurrent: name == currentBase})
	}
	return out
}

func sortAvatarEntries(out []Entry) {
	sort.Slice(out, func(i, j int) bool {
		if out[i].IsCurrent != out[j].IsCurrent {
			return out[i].IsCurrent
		}
		return out[i].Filename > out[j].Filename
	})
}

func DeleteCurrent(configDir string, memberID int) error {
	dir := Dir(configDir)
	for _, ext := range allowedTypes {
		_ = os.Remove(filepath.Join(dir, fmt.Sprintf("%d%s", memberID, ext)))
	}
	return nil
}

func DeleteAll(configDir string, memberID int) error {
	entries, err := readAvatarDir(Dir(configDir))
	if err != nil || entries == nil {
		return err
	}
	prefix := fmt.Sprintf("%d", memberID)
	for _, entry := range entries {
		if shouldDeleteMemberFile(entry, prefix) {
			_ = os.Remove(filepath.Join(Dir(configDir), entry.Name()))
		}
	}
	return nil
}

func shouldDeleteMemberFile(entry os.DirEntry, prefix string) bool {
	if entry.IsDir() || !strings.HasPrefix(entry.Name(), prefix) {
		return false
	}
	return memberOwnedAvatarName(entry.Name(), prefix)
}

func memberOwnedAvatarName(name, prefix string) bool {
	suffix := name[len(prefix):]
	return suffix != "" && (suffix[0] == '.' || suffix[0] == '-') && hasAllowedExt(name)
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
	if !isSafeAvatarFilename(filename) {
		return false
	}
	base := filepath.Base(filename)
	prefix := fmt.Sprintf("%d", memberID)
	if !strings.HasPrefix(base, prefix) {
		return false
	}
	return memberFilenameSuffixOK(base[len(prefix):], base)
}

func isSafeAvatarFilename(filename string) bool {
	return !strings.Contains(filename, "/") &&
		!strings.Contains(filename, "\\") &&
		filename == filepath.Base(filename)
}

func memberFilenameSuffixOK(suffix, base string) bool {
	if suffix == "" {
		return false
	}
	switch suffix[0] {
	case '.', '-':
		return hasAllowedExt(base)
	default:
		return false
	}
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
	return slices.Contains(values(allowedTypes), ext)
}

func values(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for _, v := range m {
		out = append(out, v)
	}
	return out
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
	if isJPEG(data) {
		return "image/jpeg"
	}
	if isPNG(data) {
		return "image/png"
	}
	if isWEBP(data) {
		return "image/webp"
	}
	return ""
}

func isJPEG(data []byte) bool {
	return len(data) >= 3 && data[0] == 0xff && data[1] == 0xd8 && data[2] == 0xff
}

func isPNG(data []byte) bool {
	return len(data) >= 8 && bytes.Equal(data[:8], []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a})
}

func isWEBP(data []byte) bool {
	return len(data) >= 12 && bytes.Equal(data[:4], []byte("RIFF")) && bytes.Equal(data[8:12], []byte("WEBP"))
}

func decodeConfig(data []byte) (image.Config, error) {
	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return image.Config{}, err
	}
	return cfg, nil
}
