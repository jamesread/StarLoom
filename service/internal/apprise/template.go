package apprise

import (
	"fmt"
	"strconv"
	"strings"

	aptemplate "github.com/jamesread/armature-apprise/template"
)

// DefaultRedemptionMessage is used when the message template cvar is empty.
const DefaultRedemptionMessage = `{{requestor_name}} requested "{{reward_name}}" ({{stars}} stars).

Approve here: {{approval_url}}`

const DefaultChoreCompletedMessage = `{{child_name}} completed "{{chore_title}}" (+{{stars}} stars) on {{date}}.`

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
	return aptemplate.Render(tmpl, vals)
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

// ChoreCompletedPlaceholders holds values substituted into the chore-completed template.
type ChoreCompletedPlaceholders struct {
	ChildName       string
	ChoreTitle      string
	Stars           int
	Date            string
	CompletedByName string
}

// RenderChoreCompletedMessage fills the chore completed notification template.
func RenderChoreCompletedMessage(tmpl string, p ChoreCompletedPlaceholders) string {
	if strings.TrimSpace(tmpl) == "" {
		tmpl = DefaultChoreCompletedMessage
	}
	completedBy := strings.TrimSpace(p.CompletedByName)
	if completedBy == "" {
		completedBy = p.ChildName
	}
	return RenderTemplate(tmpl, map[string]string{
		"child_name":        p.ChildName,
		"chore_title":       p.ChoreTitle,
		"stars":             strconv.Itoa(p.Stars),
		"date":              p.Date,
		"completed_by_name": completedBy,
	})
}
