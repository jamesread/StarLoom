package avatar

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const png1x1 = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg=="

func decodePNG(t *testing.T) []byte {
	t.Helper()
	raw, err := base64.StdEncoding.DecodeString(png1x1)
	require.NoError(t, err)
	return raw
}

func TestSavePNG(t *testing.T) {
	dir := t.TempDir()
	raw := decodePNG(t)

	path, err := Save(dir, 7, raw, "image/png")
	require.NoError(t, err)
	require.Equal(t, "7.png", path)

	full := filepath.Join(Dir(dir), path)
	info, err := os.Stat(full)
	require.NoError(t, err)
	require.Positive(t, info.Size())
}

func TestSaveDetectsPNGWithoutContentType(t *testing.T) {
	dir := t.TempDir()
	raw := decodePNG(t)

	path, err := Save(dir, 8, raw, "")
	require.NoError(t, err)
	require.Equal(t, "8.png", path)
}

func TestSaveRejectsInvalidData(t *testing.T) {
	dir := t.TempDir()
	_, err := Save(dir, 9, []byte("not-an-image"), "image/png")
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid image data")
}

func TestSaveArchivesPreviousAvatar(t *testing.T) {
	dir := t.TempDir()
	raw := decodePNG(t)

	first, err := Save(dir, 3, raw, "image/png")
	require.NoError(t, err)
	require.Equal(t, "3.png", first)

	time.Sleep(time.Second)

	second, err := Save(dir, 3, raw, "image/png")
	require.NoError(t, err)
	require.Equal(t, "3.png", second)

	entries, err := os.ReadDir(Dir(dir))
	require.NoError(t, err)
	require.Len(t, entries, 2)
}

func TestListAndSelect(t *testing.T) {
	dir := t.TempDir()
	raw := decodePNG(t)

	_, err := Save(dir, 5, raw, "image/png")
	require.NoError(t, err)
	time.Sleep(time.Second)
	_, err = Save(dir, 5, raw, "image/png")
	require.NoError(t, err)

	listed, err := List(dir, 5, "5.png")
	require.NoError(t, err)
	require.Len(t, listed, 2)
	require.True(t, listed[0].IsCurrent)

	var archived string
	for _, entry := range listed {
		if !entry.IsCurrent {
			archived = entry.Filename
		}
	}
	require.NotEmpty(t, archived)

	path, err := Select(dir, 5, archived)
	require.NoError(t, err)
	require.Equal(t, "5.png", path)
}

func TestBelongsToMember(t *testing.T) {
	require.True(t, BelongsToMember(4, "4.png"))
	require.True(t, BelongsToMember(4, "4-1234567890.webp"))
	require.False(t, BelongsToMember(4, "40.png"))
	require.False(t, BelongsToMember(4, "../4.png"))
}

func TestDeleteCurrentKeepsArchive(t *testing.T) {
	dir := t.TempDir()
	raw := decodePNG(t)

	_, err := Save(dir, 6, raw, "image/png")
	require.NoError(t, err)
	time.Sleep(time.Second)
	_, err = Save(dir, 6, raw, "image/png")
	require.NoError(t, err)

	require.NoError(t, DeleteCurrent(dir, 6))

	entries, err := os.ReadDir(Dir(dir))
	require.NoError(t, err)
	require.Len(t, entries, 1)
	require.NotEqual(t, "6.png", entries[0].Name())
}

func TestDeleteAllRemovesMemberAvatars(t *testing.T) {
	dir := t.TempDir()
	raw := decodePNG(t)

	_, err := Save(dir, 11, raw, "image/png")
	require.NoError(t, err)
	time.Sleep(time.Second)
	_, err = Save(dir, 11, raw, "image/png")
	require.NoError(t, err)

	require.NoError(t, DeleteAll(dir, 11))
	entries, err := os.ReadDir(Dir(dir))
	require.NoError(t, err)
	require.Empty(t, entries)
}
