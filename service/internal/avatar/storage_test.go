package avatar

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

const png1x1 = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg=="

func TestSavePNG(t *testing.T) {
	dir := t.TempDir()
	raw, err := base64.StdEncoding.DecodeString(png1x1)
	require.NoError(t, err)

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
	raw, err := base64.StdEncoding.DecodeString(png1x1)
	require.NoError(t, err)

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
