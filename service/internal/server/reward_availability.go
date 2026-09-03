package server

import (
	"context"
	"time"

	"github.com/jamesread/starapp/service/internal/rewards"
	"github.com/jamesread/starapp/service/internal/store"
)

func (s *Server) rewardAvailableNow(ctx context.Context, reward *store.RewardRow, memberID, balance int, now time.Time) bool {
	if reward == nil {
		return false
	}
	env, err := s.rewardAvailabilityEnv(ctx, reward, memberID, balance, now)
	if err != nil {
		return false
	}
	ok, err := rewards.EvaluateAvailabilityExpression(reward.AvailabilityExpression, env)
	return err == nil && ok
}

// rewardUnavailableDueToSchedule is true when the availability expression fails for
// reasons other than insufficient balance (e.g. time-of-day rules).
func (s *Server) rewardUnavailableDueToSchedule(ctx context.Context, reward *store.RewardRow, memberID, balance int, now time.Time) bool {
	if reward == nil || !reward.Active {
		return false
	}
	listBalance := balance
	if listBalance < reward.CostStars {
		listBalance = reward.CostStars
	}
	env, err := s.rewardAvailabilityEnv(ctx, reward, memberID, listBalance, now)
	if err != nil {
		return true
	}
	ok, err := rewards.EvaluateAvailabilityExpression(reward.AvailabilityExpression, env)
	return err != nil || !ok
}

func (s *Server) rewardAvailabilityEnv(ctx context.Context, reward *store.RewardRow, memberID, balance int, now time.Time) (rewards.AvailabilityEnv, error) {
	env := rewards.AvailabilityEnvAt(now, balance, reward.CostStars)
	countDay, err := s.store.CountApprovedRedemptionsForMemberRewardBetween(
		ctx, memberID, reward.ID, now.Format("2006-01-02"), now.Format("2006-01-02"),
	)
	if err != nil {
		return env, err
	}
	weekStart, weekEnd := weekRangeDates(now)
	countWeek, err := s.store.CountApprovedRedemptionsForMemberRewardBetween(
		ctx, memberID, reward.ID, weekStart, weekEnd,
	)
	if err != nil {
		return env, err
	}
	env.CountPerDay = countDay
	env.CountPerWeek = countWeek
	return env, nil
}

func weekRangeDates(now time.Time) (start, end string) {
	wd := int(now.Weekday())
	if wd == 0 {
		wd = 7
	}
	monday := now.AddDate(0, 0, -(wd - 1))
	sunday := monday.AddDate(0, 0, 6)
	return monday.Format("2006-01-02"), sunday.Format("2006-01-02")
}
