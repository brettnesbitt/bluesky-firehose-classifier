package ws

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"

	"stockseer.ai/blueksy-firehose/internal/appcontext"
	"stockseer.ai/blueksy-firehose/internal/domain"
)

func StartWebSocketClient(ctx context.Context, rules *domain.RuleFactory, processors *domain.TextProcessorFactory) {
	// get our app config and logger
	appCtx, _ := appcontext.AppContextFromContext(ctx)
	cfg := appCtx.Config
	logger := appCtx.Log

	uri := cfg.JetstreamURL
	conn, resp, _ := websocket.DefaultDialer.Dial(uri, http.Header{})
	defer conn.Close()
	defer resp.Body.Close()

	logger.Info("Connected to WebSocket", "url", uri)

	ConsumeMessages(appCtx, conn, rules, processors)
}

func ConsumeMessages(appCtx appcontext.AppContext, conn *websocket.Conn, rules *domain.RuleFactory, processors *domain.TextProcessorFactory) {

	logger := appCtx.Log

	// add data collector for metrics collection
	dc := domain.DataCollector{}
	metricsErr := dc.StartMetrics(appCtx)

	if metricsErr != nil {
		logger.Error("failed to start metrics collection...", metricsErr)
	}

	for {
		if _, _, err := conn.NextReader(); err != nil {
			conn.Close()
			break
		}

		_, v, _ := conn.ReadMessage()

		var m domain.Message
		jsonerr := json.Unmarshal(v, &m)

		if jsonerr != nil {
			logger.Error("failed to parse message as json", jsonerr)
		}

		passed, _ := rules.EvaluateAll(m.Commit.Record.Text)
		if passed {
			processed, _ := processors.ProcessAll(m.Commit.Record.Text)

			m.Categories = strings.Split(processed["TextCategoryClassifier"], " and ")
			m.FinSentiment = processed["TextFinSentimentClassifier"]
			dc.Add(m)

			logger.Debug("Record: \n %s \n\n", m)
		}

	}
}
