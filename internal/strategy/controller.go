package strategy

import (
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/signalb/internal/errors"
	"github.com/signalb/internal/telegram"
	"github.com/signalb/internal/timeframe"
)

func GetStrategiesController(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"strategies": StrategyManager.GetStrategies(),
	})
}

func EvaluateTickerStrategiesByTimeframeController(c *gin.Context) {
	tf := c.Param("timeframe")

	if !slices.Contains(timeframe.AllowedTimeframes, tf) {
		c.JSON(http.StatusBadRequest,
			errors.NewErrorResp(fmt.Errorf("valid timeframes: %v", timeframe.AllowedTimeframes)))
		return
	}

	res, err := evaluateTickersStrategiesByTimeframe(c.Request.Context(), tf)
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			errors.NewErrorResp(fmt.Errorf("error evaluating strategies for each ticker in the given timeframe: %w", err)))
		return
	}

	formattedOutput := formatTickersStrategiesOutput(tf, res)
	err = telegram.Bot.SendMessageByHTML(telegram.Bot.DefaultChatID, formattedOutput)
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			errors.NewErrorResp(fmt.Errorf("send updates to Telegram: %w", err)))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"results": res,
	})
}

func formatTickersStrategiesOutput(timeframe string, results map[string][]*Resp) string {
	// Build out the content message
	var resultContentBuilder strings.Builder
	for ticker, strategyResps := range results {
		// Ticker
		resultContentBuilder.WriteString(fmt.Sprintf("<b>%s</b>\n", ticker))
		for _, strategyResp := range strategyResps {
			var resultLogo string
			if strategyResp.IsFulfilled {
				resultLogo = "✅"
			} else {
				resultLogo = "❌"
			}

			resultContentBuilder.WriteString(
				fmt.Sprintf("<code>%s %s: %s</code>\n",
					resultLogo,
					strategyResp.Strategy.GetName(),
					strategyResp.EvaluationMessage,
				),
			)
		}
		resultContentBuilder.WriteString("\n")
	}

	resultContent := resultContentBuilder.String()
	if resultContent == "" {
		return ""
	}

	title := fmt.Sprintf("<b><u>%s</u></b>\n", timeframe)
	return title + resultContent
}
