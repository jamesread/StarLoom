package i18n

import "sort"

// AvailableLanguageCodes lists UI languages the app supports.
// Extend when translation files are added.
func AvailableLanguageCodes() []string {
	codes := []string{"en"}
	sort.Strings(codes)
	return codes
}

func IsSupportedLanguage(code string) bool {
	if code == "" {
		return true
	}
	for _, available := range AvailableLanguageCodes() {
		if available == code {
			return true
		}
	}
	return false
}

func LanguageLabel(code string) string {
	switch code {
	case "en":
		return "English"
	default:
		return code
	}
}
