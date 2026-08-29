package store

import "regexp"

// MemberStarColorPalette is the default set of star/avatar colors for children.
var MemberStarColorPalette = []string{
	"#e74c3c", "#3498db", "#2ecc71", "#f39c12", "#9b59b6",
	"#1abc9c", "#e67e22", "#ff6b6b", "#4ecdc4", "#6c5ce7",
}

var hexColorRE = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)

func DefaultMemberStarColor(memberID int) string {
	if len(MemberStarColorPalette) == 0 {
		return "#3498db"
	}
	idx := memberID % len(MemberStarColorPalette)
	if idx < 0 {
		idx += len(MemberStarColorPalette)
	}
	return MemberStarColorPalette[idx]
}

func NormalizeMemberStarColor(color string) string {
	if HexColorRE.MatchString(color) {
		return color
	}
	return ""
}

func NextChildStarColor(existingChildCount int) string {
	if len(MemberStarColorPalette) == 0 {
		return "#3498db"
	}
	return MemberStarColorPalette[existingChildCount%len(MemberStarColorPalette)]
}

// HexColorRE validates #RRGGBB colors.
var HexColorRE = hexColorRE
