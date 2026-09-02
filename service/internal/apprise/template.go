package apprise

import (
	"fmt"
	"strconv"
	"strings"
)

// DefaultRedemptionMessage is used when the message template cvar is empty.
const DefaultRedemptionMessage = `{{requestor_name}} requested "{{reward_name}}" ({{stars}} stars).

Approve here: {{approval_url}}`

// RedemptionPlaceholders holds values substituted into the message template.
type RedemptionPlaceholders struct {
	ApprovalURL   string
	RequestorName string
	RewardName    string
	Stars         int
	RedemptionID  int
	RequestorID   int
}

// ApprovalURL builds an absolute (or path-only) deep link to the pending rewards page.
func ApprovalURL(externalBaseURL string, redemptionID int) string {
	path := fmt.Sprintf("/family/rewards?redemption=%d", redemptionID)
	base := strings.TrimRight(strings.TrimSpace(externalBaseURL), "/")
	if base == "" {
		return path
	}
	return base + path
}

// RenderTemplate replaces {{placeholders}} in tmpl. Unknown placeholders are left unchanged.
func RenderTemplate(tmpl string, vals map[string]string) string {
	if tmpl == "" {
		return ""
	}
	var b strings.Builder
	b.Grow(len(tmpl))
	for i := 0; i < len(tmpl); {
		if tmpl[i] == '{' && i+1 < len(tmpl) && tmpl[i+1] == '{' {
			end := strings.Index(tmpl[i+2:], "}}")
			if end >= 0 {
				key := strings.TrimSpace(tmpl[i+2 : i+2+end])
				if v, ok := vals[key]; ok {
					b.WriteString(v)
					i += 2 + end + 2
					continue
				}
			}
		}
		b.WriteByte(tmpl[i])
		i++
	}
	return b.String()
}

// RenderRedemptionMessage fills the redemption approval template.
func RenderRedemptionMessage(tmpl string, p RedemptionPlaceholders) string {
	if strings.TrimSpace(tmpl) == "" {
		tmpl = DefaultRedemptionMessage
	}
	return RenderTemplate(tmpl, map[string]string{
		"approval_url":   p.ApprovalURL,
		"requestor_name": p.RequestorName,
		"reward_name":    p.RewardName,
		"stars":          strconv.Itoa(p.Stars),
		"redemption_id":  strconv.Itoa(p.RedemptionID),
		"requestor_id":   strconv.Itoa(p.RequestorID),
	})
}
