package themes

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAvailableNamesFromIndex(t *testing.T) {
	dir := t.TempDir()
	themesDir := filepath.Join(dir, "themes")
	if err := os.MkdirAll(filepath.Join(themesDir, "egypt"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(themesDir, "egypt", "theme.css"), []byte("body{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	index := `{"themes":[{"id":"egypt","label":"Egypt"}]}`
	if err := os.WriteFile(filepath.Join(themesDir, "index.json"), []byte(index), 0o644); err != nil {
		t.Fatal(err)
	}

	got := AvailableNames(dir)
	if len(got) != 1 || got[0] != "egypt" {
		t.Fatalf("got %v", got)
	}
}

func TestAvailableNamesMergesSupplementalThemes(t *testing.T) {
	dir := t.TempDir()
	themesDir := filepath.Join(dir, "themes")
	supplementalDir := filepath.Join(dir, "supplemental-themes")
	if err := os.MkdirAll(filepath.Join(themesDir, "space"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(themesDir, "space", "theme.css"), []byte("body{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(themesDir, "index.json"), []byte(`{"themes":["space"]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(supplementalDir, "waffles"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(supplementalDir, "waffles", "theme.css"), []byte("body{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(supplementalDir, "index.json"), []byte(`{"themes":["waffles"]}`), 0o644); err != nil {
		t.Fatal(err)
	}

	got := AvailableNames(dir)
	want := []string{"space", "waffles"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}

func TestAvailableNamesFallbackIncludesBundledThemes(t *testing.T) {
	got := AvailableNames("")
	want := []string{
		"ancient-greece",
		"aztecs",
		"catppuccin-latte-frappe",
		"dracula-alucard",
		"egypt",
		"gruvbox-dark-light",
		"space",
		"waffles",
	}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}

func TestIsValidName(t *testing.T) {
	allowed := []string{
		"ancient-greece",
		"aztecs",
		"catppuccin-latte-frappe",
		"dracula-alucard",
		"egypt",
		"gruvbox-dark-light",
		"space",
		"waffles",
	}
	if !IsValidName("", allowed) {
		t.Fatal("empty should be valid")
	}
	if !IsValidName("ancient-greece", allowed) {
		t.Fatal("ancient-greece should be valid")
	}
	if !IsValidName("aztecs", allowed) {
		t.Fatal("aztecs should be valid")
	}
	if !IsValidName("egypt", allowed) {
		t.Fatal("egypt should be valid")
	}
	if !IsValidName("space", allowed) {
		t.Fatal("space should be valid")
	}
	if !IsValidName("waffles", allowed) {
		t.Fatal("waffles should be valid")
	}
	if IsValidName("mars", allowed) {
		t.Fatal("mars should be invalid")
	}
}
