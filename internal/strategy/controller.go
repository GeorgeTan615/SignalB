package strategy

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/signalb/internal/errors"
	"github.com/signalb/internal/telegram"
	"github.com/signalb/internal/timeframe"
	"github.com/signalb/utils"
)

func GetStrategiesController(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"strategies": AllowedStrategies,
	})
}

func EvaluateTickerStrategiesByTimeframeController(c *gin.Context) {
	tf := c.Param("timeframe")

	if !utils.SliceContains[string](timeframe.AllowedTimeframes[:], tf) {
		c.JSON(http.StatusBadRequest, errors.NewErrorResp(fmt.Sprintf("Timeframe must be of %v", timeframe.AllowedTimeframes)))
		return
	}

	res, err := evaluateTickersStrategiesByTimeframe(tf)

	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorRespWithErr("Error evaluating strategies for each ticker in the given timeframe", err))
		return
	}

	formattedOutput := formatTickersStrategiesOutput(tf, res)
	err = telegram.Bot.SendMessageByHTML(telegram.Bot.DefaultChatId, formattedOutput)

	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorRespWithErr("Error sending updates to Telegram", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"results": res,
	})
}

func formatTickersStrategiesOutput(timeframe string, results map[string][]*StrategyResp) string {
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
				fmt.Sprintf("<code>%s %s: %s</code>\n", resultLogo, strategyResp.Strategy.GetName(), strategyResp.EvaluationMessage))
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
