package server

import (
	"time"

	"github.com/jamesread/starapp/service/internal/rewards"
	"github.com/jamesread/starapp/service/internal/store"
)

func rewardAvailableNow(reward *store.RewardRow, balance int, now time.Time) bool {
	if reward == nil {
		return false
	}
	ok, err := rewards.EvaluateAvailabilityExpression(
		reward.AvailabilityExpression,
		rewards.AvailabilityEnvAt(now, balance, reward.CostStars),
	)
	return err == nil && ok
}

// rewardUnavailableDueToSchedule is true when the availability expression fails for
// reasons other than insufficient balance (e.g. time-of-day rules).
func rewardUnavailableDueToSchedule(reward *store.RewardRow, balance int, now time.Time) bool {
	if reward == nil || !reward.Active {
		return false
	}
	listBalance := balance
	if listBalance < reward.CostStars {
		listBalance = reward.CostStars
	}
	ok, err := rewards.EvaluateAvailabilityExpression(
		reward.AvailabilityExpression,
		rewards.AvailabilityEnvAt(now, listBalance, reward.CostStars),
	)
	return err != nil || !ok
}
