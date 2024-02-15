package strategy

import (
	"fmt"
	"log"

	"github.com/signalb/internal/marketprice"
)

type RSIStrategy struct {
	Level    float64
	Strength StrategyStrength
	Type     StrategyType
}

func NewRSIStrategy(level float64, strength StrategyStrength, typ StrategyType) *RSIStrategy {
	return &RSIStrategy{
		Level:    level,
		Strength: strength,
		Type:     typ,
	}
}

func (s *RSIStrategy) GetName() string {
	return fmt.Sprintf("%s%0f", "rsi", s.Level)
}

// Data must be in ascending order
func (s *RSIStrategy) Evaluate(data []*marketprice.TickerData) (string, bool) {
	if len(data) != 200 {
		log.Println("Number of data should be 200")
	}

	rsi := calculateRSI(data, len(data))

	return s.getSuccessMessage(rsi), s.isRSIReachedLevel(rsi)
}

func (s *RSIStrategy) getSuccessMessage(rsi float64) string {
	if s.Type == Notify {
		return fmt.Sprintf("RSI(%0.2f) reached %d levels", rsi, int(s.Level))
	} else {
		return fmt.Sprintf("%s %s, RSI(%0.2f) in %0f zone", s.Strength, s.Type, rsi, s.Level)
	}
}

func (s *RSIStrategy) isRSIReachedLevel(rsi float64) bool {
	if s.Type == Sell {
		return rsi >= s.Level
	} else if s.Type == Buy {
		return rsi <= s.Level
	} else {
		return rsi >= s.Level || rsi <= s.Level
	}
}

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
