package rewards

import (
	"testing"
	"time"
)

func TestValidateAvailabilityExpression(t *testing.T) {
	if err := ValidateAvailabilityExpression(""); err != nil {
		t.Fatalf("empty: %v", err)
	}
	if err := ValidateAvailabilityExpression(`hour > 9 && hour < 18`); err != nil {
		t.Fatalf("valid: %v", err)
	}
	if err := ValidateAvailabilityExpression(`countPerDay < 2 && countPerWeek < 5`); err != nil {
		t.Fatalf("count vars: %v", err)
	}
	if err := ValidateAvailabilityExpression(`unknownVar == 1`); err == nil {
		t.Fatal("unknown var should fail")
	}
	if err := ValidateAvailabilityExpression(`hour + 1`); err == nil {
		t.Fatal("non-bool should fail")
	}
}

func TestEvaluateAvailabilityExpression(t *testing.T) {
	env := AvailabilityEnv{
		Hour: 10, Minute: 30, DayName: "Sat", Day: 15, Month: 8, Year: 2026,
		Balance: 20, CostStars: 5,
	}
	ok, err := EvaluateAvailabilityExpression(`(hour > 9 && hour < 18) && (dayName == "Sat" || dayName == "Sun")`, env)
	if err != nil || !ok {
		t.Fatalf("weekend daytime: ok=%v err=%v", ok, err)
	}
	env.DayName = "Mon"
	ok, err = EvaluateAvailabilityExpression(`(hour > 9 && hour < 18) && (dayName == "Sat" || dayName == "Sun")`, env)
	if err != nil || ok {
		t.Fatalf("weekday should be unavailable: ok=%v err=%v", ok, err)
	}
	ok, err = EvaluateAvailabilityExpression("", env)
	if err != nil || !ok {
		t.Fatalf("empty means always available: ok=%v err=%v", ok, err)
	}
	env.CountPerDay = 2
	ok, err = EvaluateAvailabilityExpression(`countPerDay < 2`, env)
	if err != nil || ok {
		t.Fatalf("countPerDay limit: ok=%v err=%v", ok, err)
	}
	env.CountPerWeek = 4
	ok, err = EvaluateAvailabilityExpression(`countPerWeek < 5`, env)
	if err != nil || !ok {
		t.Fatalf("countPerWeek under limit: ok=%v err=%v", ok, err)
	}
}

func TestAvailabilityEnvAt(t *testing.T) {
	when := time.Date(2026, 8, 30, 14, 45, 0, 0, time.Local)
	env := AvailabilityEnvAt(when, 12, 3)
	if env.Hour != 14 || env.Minute != 45 || env.DayName != "Sun" || env.Balance != 12 || env.CostStars != 3 {
		t.Fatalf("env=%+v", env)
	}
}
