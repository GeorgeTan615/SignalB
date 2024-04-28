package strategy

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	_bitcoinFNGApiURL = "https://api.alternative.me/fng/"
	_extremeFear      = "Extreme Fear"
	_fear             = "Fear"
	_neutral          = "Neutral"
	_greed            = "Greed"
	_extremeGreed     = "Extreme Greed"
)

var _fngWhitelistedTickerSymbols = []string{"BITCOIN"}

type FearNGreedIdx struct {
	httpClient *http.Client
}

func newFearNGreedIdx() *FearNGreedIdx {
	return &FearNGreedIdx{
		httpClient: &http.Client{},
	}
}

func (s *FearNGreedIdx) GetName() string {
	return fmt.Sprintf("fng")
}

func (s *FearNGreedIdx) GetWhitelistedTickerSymbols() []string {
	return _fngWhitelistedTickerSymbols
}

func (s *FearNGreedIdx) Evaluate(_ []float64) *EvaluationResult {
	resp, err := s.httpClient.Get(_bitcoinFNGApiURL)
	if err != nil {
		return NewEvaluationResult(false, fmt.Sprintf("get fng api: %v", err))
	}

	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return NewEvaluationResult(false, fmt.Sprintf("read fng api resp body: %v", err))
	}

	var fngData struct {
		Data []struct {
			Value               string `json:"value"`
			ValueClassification string `json:"value_classification"`
		} `json:"data"`
	}

	if err := json.Unmarshal(bodyBytes, &fngData); err != nil {
		return NewEvaluationResult(false, fmt.Sprintf("unmarshal fng api resp body: %v", err))
	}

	if len(fngData.Data) < 1 {
		return NewEvaluationResult(false, fmt.Sprintf("expected data at least of length 1, got: %v", fngData.Data))
	}

	var (
		fngValue               = fngData.Data[0].Value
		fngValueClassification = fngData.Data[0].ValueClassification
	)

	strength, err := s.getStrength(fngValueClassification)
	if err != nil {
		return NewEvaluationResult(false, err.Error())
	}

	typ, err := s.getTyp(fngValueClassification)
	if err != nil {
		return NewEvaluationResult(false, err.Error())
	}

	isActionNeeded := typ == Buy || typ == Sell

	return NewEvaluationResult(
		isActionNeeded,
		s.getEvaluationMessage(
			fngValue,
			fngValueClassification,
			strength,
			typ,
		),
	)
}

func (s *FearNGreedIdx) getStrength(valueClassification string) (Strength, error) {
	switch valueClassification {
	case _extremeFear, _extremeGreed:
		return VeryStrong, nil
	case _fear, _greed:
		return Strong, nil
	case _neutral:
		return Neutral, nil
	default:
		return "", fmt.Errorf("unexpected value classification: %v", valueClassification)
	}
}

func (s *FearNGreedIdx) getTyp(valueClassification string) (Type, error) {
	switch valueClassification {
	case _extremeFear, _fear:
		return Buy, nil
	case _greed, _extremeGreed:
		return Sell, nil
	case _neutral:
		return Notify, nil
	default:
		return "", fmt.Errorf("unexpected value classification: %v", valueClassification)
	}
}

func (s *FearNGreedIdx) getEvaluationMessage(
	fngValue,
	fngValueClassification string,
	strength Strength,
	typ Type,
) string {
	return fmt.Sprintf("%s %s! %s(%s)", strength, typ, fngValueClassification, fngValue)
}
