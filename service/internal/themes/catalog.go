package themes

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type indexEntry struct {
	ID string `json:"id"`
}

var bundledThemeNames = []string{"ancient-greece", "aztecs", "egypt", "space"}

var bundledSupplementalThemeNames = []string{
	"catppuccin-latte-frappe",
	"dracula-alucard",
	"gruvbox-dark-light",
	"waffles",
}

// AvailableNames returns drop-in theme folder names from webui/themes and
// webui/supplemental-themes, or bundled fallbacks when webui is unavailable.
func AvailableNames(webuiDir string) []string {
	if webuiDir != "" {
		names := themeNamesFromWebUIDir(webuiDir)
		if len(names) > 0 {
			return names
		}
	}
	out := append(append([]string{}, bundledThemeNames...), bundledSupplementalThemeNames...)
	return uniqueSorted(out)
}

func themeNamesFromWebUIDir(webuiDir string) []string {
	var names []string

	themesDir := filepath.Join(webuiDir, "themes")
	if fromIndex := namesFromIndex(filepath.Join(themesDir, "index.json")); len(fromIndex) > 0 {
		names = append(names, fromIndex...)
	} else if fromDir := namesFromDir(themesDir); len(fromDir) > 0 {
		names = append(names, fromDir...)
	}

	supplementalDir := filepath.Join(webuiDir, "supplemental-themes")
	if fromIndex := namesFromIndex(filepath.Join(supplementalDir, "index.json")); len(fromIndex) > 0 {
		names = append(names, fromIndex...)
	} else if fromDir := namesFromDir(supplementalDir); len(fromDir) > 0 {
		names = append(names, fromDir...)
	}

	return uniqueSorted(names)
}

func namesFromIndex(path string) []string {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var payload struct {
		Themes []json.RawMessage `json:"themes"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil
	}
	names := make([]string, 0, len(payload.Themes))
	for _, raw := range payload.Themes {
		var asString string
		if err := json.Unmarshal(raw, &asString); err == nil && asString != "" {
			names = append(names, asString)
			continue
		}
		var asObject indexEntry
		if err := json.Unmarshal(raw, &asObject); err == nil && asObject.ID != "" {
			names = append(names, asObject.ID)
		}
	}
	return uniqueSorted(names)
}

func namesFromDir(themesDir string) []string {
	entries, err := os.ReadDir(themesDir)
	if err != nil {
		return nil
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if _, err := os.Stat(filepath.Join(themesDir, entry.Name(), "theme.css")); err != nil {
			continue
		}
		names = append(names, entry.Name())
	}
	return uniqueSorted(names)
}

func IsValidName(name string, allowed []string) bool {
	name = strings.TrimSpace(name)
	if name == "" {
		return true
	}
	for _, candidate := range allowed {
		if candidate == name {
			return true
		}
	}
	return false
}

func uniqueSorted(names []string) []string {
	seen := make(map[string]struct{}, len(names))
	out := make([]string, 0, len(names))
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}
