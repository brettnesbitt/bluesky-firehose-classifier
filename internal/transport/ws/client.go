package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

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
	for {
		conn, resp, err := websocket.DefaultDialer.Dial(uri, http.Header{})
		if err != nil {
			logger.Error("failed to connect to WebSocket", err)
			time.Sleep(5 * time.Second) // Wait before retrying
			continue
		}

		defer resp.Body.Close()
		defer conn.Close()

		logger.Info("Connected to WebSocket: %s", uri)

		ConsumeMessages(appCtx, conn, rules, processors)

		// If ConsumeMessages returns, it means the connection was closed
		logger.Info("Connection closed, attempting to reconnect...")
		time.Sleep(5 * time.Second) // Wait before attempting to reconnect
	}
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
			fmt.Printf("error reading message: %v\n", err)
			break
		}

		_, v, _ := conn.ReadMessage()

		var m domain.Message
		jsonerr := json.Unmarshal(v, &m)

		if jsonerr != nil {
			fmt.Println("failed to parse message as json !!!!!")
			//logger.Error("failed to parse message as json", jsonerr)
			continue
		}

		passed, _ := rules.EvaluateAll(m.Commit.Record.Text)
		if passed {
			processed, _ := processors.ProcessAll(m.Commit.Record.Text)

			// Clean up categories by removing special characters and splitting on " and "
			cleaned := regexp.MustCompile(`[\W_]+`).ReplaceAllString(processed["TextCategoryClassifier"], " ")
			m.Categories = strings.Fields(cleaned)
			m.FinSentiment = processed["TextFinSentimentClassifier"]
			if err := dc.Add(m); err != nil {
				logger.Error("failed to add message", err)
			}

			logger.Debug("Record: \n %s \n\n", m)
		}

	}
}
