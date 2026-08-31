package rewards

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
)

// AvailabilityEnv is the variable set exposed to reward availability expressions.
// Times use the server local timezone.
type AvailabilityEnv struct {
	Hour      int
	Minute    int
	DayName   string
	Day       int
	Month     int
	Year      int
	Balance   int
	CostStars int
}

var programCache sync.Map

var availabilityEnvTypes = map[string]any{
	"hour":      0,
	"minute":    0,
	"dayName":   "",
	"day":       0,
	"month":     0,
	"year":      0,
	"balance":   0,
	"costStars": 0,
}

func availabilityEnvMap(env AvailabilityEnv) map[string]any {
	return map[string]any{
		"hour":      env.Hour,
		"minute":    env.Minute,
		"dayName":   env.DayName,
		"day":       env.Day,
		"month":     env.Month,
		"year":      env.Year,
		"balance":   env.Balance,
		"costStars": env.CostStars,
	}
}

func AvailabilityEnvAt(now time.Time, balance, costStars int) AvailabilityEnv {
	return AvailabilityEnv{
		Hour:      now.Hour(),
		Minute:    now.Minute(),
		DayName:   now.Format("Mon"),
		Day:       now.Day(),
		Month:     int(now.Month()),
		Year:      now.Year(),
		Balance:   balance,
		CostStars: costStars,
	}
}

func ValidateAvailabilityExpression(exprStr string) error {
	_, err := compileAvailability(exprStr)
	return err
}

func EvaluateAvailabilityExpression(exprStr string, env AvailabilityEnv) (bool, error) {
	exprStr = strings.TrimSpace(exprStr)
	if exprStr == "" {
		return true, nil
	}
	program, err := compileAvailability(exprStr)
	if err != nil {
		return false, err
	}
	out, err := expr.Run(program, availabilityEnvMap(env))
	if err != nil {
		return false, err
	}
	ok, boolOK := out.(bool)
	if !boolOK {
		return false, fmt.Errorf("expression did not evaluate to bool")
	}
	return ok, nil
}

func compileAvailability(exprStr string) (*vm.Program, error) {
	exprStr = strings.TrimSpace(exprStr)
	if exprStr == "" {
		return nil, nil
	}
	if cached, ok := programCache.Load(exprStr); ok {
		return cached.(*vm.Program), nil
	}
	program, err := expr.Compile(exprStr, expr.AsBool(), expr.Env(availabilityEnvTypes))
	if err != nil {
		return nil, err
	}
	programCache.Store(exprStr, program)
	return program, nil
}
