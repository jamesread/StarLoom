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
