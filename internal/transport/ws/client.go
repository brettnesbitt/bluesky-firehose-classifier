package ws

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"

	"stockseer.ai/blueksy-firehose/internal/config"
	"stockseer.ai/blueksy-firehose/internal/domain"
	"stockseer.ai/blueksy-firehose/internal/logger"
)

func StartWebSocketClient(cfg *config.AppConfig, rules *domain.RuleFactory, processors *domain.TextProcessorFactory) {
	uri := cfg.JetstreamURL
	conn, resp, _ := websocket.DefaultDialer.Dial(uri, http.Header{})
	defer conn.Close()
	defer resp.Body.Close()

	logger.Info("Connected to WebSocket", "url", uri)

	ConsumeMessages(conn, rules, processors)
}

func ConsumeMessages(conn *websocket.Conn, rules *domain.RuleFactory, processors *domain.TextProcessorFactory) {

	// add data collector for metrics collection
	dc := domain.DataCollector{}
	metricsErr := dc.StartMetrics()

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

			fmt.Println(processed, "........$")

			m.Categories = strings.Split(processed["TextCategoryClassifier"], " and ")
			m.FinSentiment = processed["TextFinSentimentClassifier"]
			dc.Add(m)

			fmt.Println("........")
			fmt.Println(m)
		}

	}
}
