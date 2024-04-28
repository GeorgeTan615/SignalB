package strategy

import (
	"fmt"
)

type (
	Strength string
	Type     string
)

const (
	Key Strength = "Key"

	VeryWeak   Strength = "Very Weak"
	Weak       Strength = "Weak"
	Neutral    Strength = "Neutral"
	Strong     Strength = "Strong"
	VeryStrong Strength = "Very Strong"

	Sell   Type = "Sell"
	Buy    Type = "Buy"
	Notify Type = "Notify"
)

type EvaluationResult struct {
	IsFulfilled       bool
	EvaluationMessage string
}

func NewEvaluationResult(isFulfilled bool, message string) *EvaluationResult {
	return &EvaluationResult{
		IsFulfilled:       isFulfilled,
		EvaluationMessage: message,
	}
}

type Strategy interface {
	GetName() string
	GetWhitelistedTickerSymbols() []string
	Evaluate(data []float64) *EvaluationResult
}

var (
	StrategyManager *Manager
)

type Manager struct {
	NameToStrategyMap map[string]Strategy
}

func NewStrategyManager(strategies ...Strategy) *Manager {
	nameToStrategyMap := make(map[string]Strategy, len(strategies))

	for _, strategy := range strategies {
		nameToStrategyMap[strategy.GetName()] = strategy
	}

	return &Manager{
		NameToStrategyMap: nameToStrategyMap,
	}
}

func (sm *Manager) GetStrategyByName(strategyName string) (Strategy, error) {
	strategy, ok := sm.NameToStrategyMap[strategyName]

	if !ok {
		return nil, fmt.Errorf("strategy %s not found, check if strategy is registered", strategyName)
	}

	return strategy, nil
}

func (sm *Manager) GetStrategies() []string {
	res := make([]string, 0, len(sm.NameToStrategyMap))

	for name := range sm.NameToStrategyMap {
		res = append(res, name)
	}

	return res
}

func InitStrategies() {
	// RSI
	rsi20, rsi30, rsi40, rsi70, rsi80 := NewRSI(20, VeryStrong, Buy),
		NewRSI(30, Strong, Buy),
		NewRSI(40, Key, Buy),
		NewRSI(70, Strong, Sell),
		NewRSI(80, VeryStrong, Sell)

	// SMA
	sma200 := newSMA(200, VeryStrong)

	// FNG
	fng := newFearNGreedIdx()

	StrategyManager = NewStrategyManager(
		rsi20, rsi30, rsi40, rsi70, rsi80,
		sma200,
		fng,
	)
}
