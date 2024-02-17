package strategy

import (
	"fmt"
	"log"

	"github.com/signalb/internal/marketprice"
)

const (
	zoneTolerancePercentage = 10
	length                  = 14
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
	return fmt.Sprintf("%s%0.f", "rsi", s.Level)
}

func (s *RSIStrategy) Evaluate(data []float64) *EvaluationResult {
	if len(data) != marketprice.RefreshAllDataLength {
		log.Printf("Number of data should be %d", marketprice.RefreshAllDataLength)
	}

	rsi := calculateRSI(data, length)

	isSuccess := s.isRSIReachedLevel(rsi)

	return &EvaluationResult{
		IsFulfilled:       s.isRSIReachedLevel(rsi),
		EvaluationMessage: s.getEvaluationMessage(rsi, isSuccess),
	}
}

func (s *RSIStrategy) getEvaluationMessage(rsi float64, isSuccess bool) string {
	if !isSuccess {
		return fmt.Sprintf("RSI(%0.9f) did not reach %0.f levels", rsi, s.Level)
	}

	if s.Type == Notify {
		return fmt.Sprintf("RSI(%0.2f) reached %d levels", rsi, int(s.Level))
	} else {
		return fmt.Sprintf("%s %s, RSI(%0.2f) in %0.f zone", s.Strength, s.Type, rsi, s.Level)
	}
}

func (s *RSIStrategy) isRSIReachedLevel(rsi float64) bool {
	if s.Type == Sell {
		return rsi >= s.Level
	} else if s.Type == Buy {
		return rsi <= s.Level
	} else {
		extra10Percent := (100.00 + zoneTolerancePercentage) / 100
		less10Percent := (100.00 - zoneTolerancePercentage) / 100
		return s.Level*extra10Percent >= rsi && rsi >= s.Level*less10Percent
	}
}

// Reference https://blog.quantinsti.com/rsi-indicator/
func calculateRSI(data []float64, length int) float64 {
	prices := data
	period := length
	var averageGain, averageLoss, rs float64

	// Calculate initial average gain and loss
	for i := 1; i <= period; i++ {
		diff := prices[i] - prices[i-1]
		if diff > 0 {
			averageGain += diff
		} else {
			averageLoss -= diff
		}
	}

	averageGain /= float64(period)
	averageLoss /= float64(period)

	for i := period; i < len(prices); i++ {
		currentGain := 0.0
		currentLoss := 0.0

		diff := prices[i] - prices[i-1]
		if diff > 0 {
			currentGain = diff
		} else {
			currentLoss = -diff
		}

		averageGain = (averageGain*(float64(period-1)) + currentGain) / float64(period)
		averageLoss = (averageLoss*(float64(period-1)) + currentLoss) / float64(period)
	}

	rs = averageGain / averageLoss
	return 100 - (100 / (1 + rs))
}
