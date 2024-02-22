package strategy

import "fmt"

const (
	tolerancePercentage float64 = 10
)

type SMA struct {
	Length   int
	Strength StrategyStrength
}

func newSMA(length int, strength StrategyStrength) *SMA {
	return &SMA{
		Length:   length,
		Strength: strength,
	}
}

func (s *SMA) GetName() string {
	return fmt.Sprintf("sma%d", s.Length)
}

func (s *SMA) Evaluate(data []float64) *EvaluationResult {
	if len(data) < s.Length {
		return NewEvaluationResult(false, fmt.Sprintf("lack %d data", s.Length))
	}

	var sum float64

	for i := 0; i < s.Length; i++ {
		idx := len(data) - 1 - i
		sum += data[idx]
	}

	sma := sum / float64(s.Length)
	latestPrice := data[len(data)-1]
	upperZone := sma * ((100 + tolerancePercentage) / 100)
	lowerZone := sma * ((100 - tolerancePercentage) / 100)
	isPriceInZone := upperZone >= latestPrice && latestPrice >= lowerZone

	return NewEvaluationResult(isPriceInZone, s.getEvaluationMessage(sma, isPriceInZone))
}

func (s *SMA) getEvaluationMessage(sma float64, isSuccess bool) string {
	if !isSuccess {
		return fmt.Sprintf("Price not at %s levels(%0.2f)", s.GetName(), sma)
	}

	return fmt.Sprintf("%s zone! Price at %s levels(%0.2f)", s.Strength, s.GetName(), sma)
}
