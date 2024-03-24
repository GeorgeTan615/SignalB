package strategy

import (
	"fmt"
	"log"

	"github.com/signalb/internal/marketprice"
)

const (
	zoneTolerance = 2
	length        = 14
)

type RSI struct {
	Level    float64
	Strength Strength
	Type     Type
}

func NewRSI(level float64, strength Strength, typ Type) *RSI {
	return &RSI{
		Level:    level,
		Strength: strength,
		Type:     typ,
	}
}

func (s *RSI) GetName() string {
	return fmt.Sprintf("rsi%0.f", s.Level)
}

func (s *RSI) Evaluate(data []float64) *EvaluationResult {
	if len(data) != marketprice.RefreshAllDataLength {
		log.Printf("Number of data should be %d", marketprice.RefreshAllDataLength)
	}

	rsi := calculateRSI(data, length)

	isSuccess := s.isRSIReachedLevel(rsi)

	return NewEvaluationResult(s.isRSIReachedLevel(rsi), s.getEvaluationMessage(rsi, isSuccess))
}

func (s *RSI) getEvaluationMessage(rsi float64, isSuccess bool) string {
	if !isSuccess {
		return fmt.Sprintf("RSI of %0.2f not at %s levels", rsi, s.GetName())
	}

	var msg string
	if s.Type == Notify {
		msg = fmt.Sprintf("RSI of %0.2f reached %s levels", rsi, s.GetName())
	} else {
		msg = fmt.Sprintf("%s %s! RSI of %0.2f in %s zone", s.Strength, s.Type, rsi, s.GetName())
	}

	return msg
}

func (s *RSI) isRSIReachedLevel(rsi float64) bool {
	var res bool

	upperZone := s.Level + zoneTolerance
	lowerZone := s.Level - zoneTolerance

	switch s.Type {
	case Sell:
		res = rsi >= lowerZone
	case Buy:
		res = rsi <= upperZone
	case Notify:
		fallthrough
	default:
		res = upperZone >= rsi && rsi >= lowerZone
	}

	return res
}

// Reference https://blog.quantinsti.com/rsi-indicator/
func calculateRSI(data []float64, length int) float64 {
	prices := data
	period := length
	var averageGain, averageLoss, rs float64

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
