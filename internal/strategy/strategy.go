package strategy

import (
	"fmt"

	"github.com/signalb/internal/marketprice"
)

type StrategyStrength string
type StrategyType string

const (
	VeryWeak   StrategyStrength = "Very Weak"
	Weak       StrategyStrength = "Weak"
	Strong     StrategyStrength = "Strong"
	VeryStrong StrategyStrength = "Very Strong"

	Sell   StrategyType = "Sell"
	Buy    StrategyType = "Buy"
	Notify StrategyType = "Notify"
)

type Strategy interface {
	GetName() string
	Evaluate(data []*marketprice.TickerData) (string, bool)
}

var AllowedStrategies []string
var strategyManager *StrategyManager

type StrategyManager struct {
	Strategies        []Strategy
	NameToStrategyMap map[string]Strategy
}

func NewStrategyManager(strategies ...Strategy) *StrategyManager {
	nameToStrategyMap := make(map[string]Strategy, len(strategies))

	for _, strategy := range strategies {
		nameToStrategyMap[strategy.GetName()] = strategy
		AllowedStrategies = append(AllowedStrategies, strategy.GetName())
	}

	return &StrategyManager{
		Strategies: strategies,
	}
}

func (sm *StrategyManager) GetStrategyByName(strategyName string) (Strategy, error) {
	strategy, ok := sm.NameToStrategyMap[strategyName]

	if !ok {
		return nil, fmt.Errorf("strategy %s not found, check if strategy is registered", strategyName)
	}

	return strategy, nil
}

func init() {
	// RSI
	rsi20, rsi30, rsi70, rsi80 :=
		NewRSIStrategy(20, VeryStrong, Buy),
		NewRSIStrategy(30, Strong, Buy),
		NewRSIStrategy(70, Strong, Sell),
		NewRSIStrategy(80, VeryStrong, Sell)

	strategyManager = NewStrategyManager(
		rsi20, rsi30, rsi70, rsi80,
	)
}
