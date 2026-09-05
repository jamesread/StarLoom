package apprise

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPersonTag(t *testing.T) {
	assert.Equal(t, "starloom_uid_42", PersonTag(42))
	assert.Equal(t, "starloom_uid_1,starloom_uid_3", JoinPersonTags([]int{1, 3}))
	assert.Equal(t, "", JoinPersonTags(nil))
}

func TestApprovalURL(t *testing.T) {
	assert.Equal(t, "/family/rewards?redemption=9", ApprovalURL("", 9))
	assert.Equal(t, "https://stars.example/family/rewards?redemption=9", ApprovalURL("https://stars.example/", 9))
}

func TestRenderRedemptionMessage(t *testing.T) {
	body := RenderRedemptionMessage(
		"{{requestor_name}} wants {{reward_name}} ({{stars}}). {{approval_url}} #{{redemption_id}}",
		RedemptionPlaceholders{
			ApprovalURL:   "https://app/family/rewards?redemption=3",
			RequestorName: "Alex",
			RewardName:    "Screen time",
			Stars:         5,
			RedemptionID:  3,
			RequestorID:   12,
		},
	)
	assert.Equal(t, "Alex wants Screen time (5). https://app/family/rewards?redemption=3 #3", body)

	fallback := RenderRedemptionMessage("", RedemptionPlaceholders{
		ApprovalURL:   "/family/rewards?redemption=1",
		RequestorName: "Sam",
		RewardName:    "Ice cream",
		Stars:         2,
	})
	assert.Contains(t, fallback, "Sam")
	assert.Contains(t, fallback, "Ice cream")
	assert.Contains(t, fallback, "/family/rewards?redemption=1")
}
