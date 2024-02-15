package strategy

import (
	"fmt"
	"log"

	"github.com/signalb/internal/marketprice"
)

type RSIStrategy struct {
	Level    int
	Strength StrategyStrength
	Type     StrategyType
}

func NewRSIStrategy(level int, strength StrategyStrength, typ StrategyType) *RSIStrategy {
	return &RSIStrategy{
		Level:    level,
		Strength: strength,
		Type:     typ,
	}
}

func (s *RSIStrategy) GetName() string {
	return fmt.Sprintf("%s%d", "rsi", s.Level)
}

func (s *RSIStrategy) Evaluate(data []*marketprice.TickerData) bool {
	// TODO ensure all data is ascending order
	if len(data) != 200 {
		log.Println("Number of data should be 200")
	}

	// TODO continue
	rsi := calculateRSI(data, len(data))

	return true
}

func (s *RSIStrategy) GetStrength() StrategyStrength {
	return s.Strength
}

func (s *RSIStrategy) GetType() StrategyType {
	return s.Type
}

// Calculate RSI for a given set of prices and a specified period
func calculateRSI(data []*marketprice.TickerData, period int) float64 {
	gain := 0.0
	loss := 0.0

	for i := 1; i <= period; i++ {
		prevDataPrice, currDataPrice := data[i-1].Price, data[i].Price
		diff := currDataPrice - prevDataPrice

		if diff > 0 {
			gain += diff
		} else {
			loss += -diff
		}
	}

	avgGain := gain / float64(period)
	avgLoss := loss / float64(period)
	rs := avgGain / avgLoss
	rsi := 100 - (100 / (1 + rs))

	return rsi
}
