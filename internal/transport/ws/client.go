package ws

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gorilla/websocket"
	"stockseer.ai/blueksy-firehose/internal/appcontext"
	"stockseer.ai/blueksy-firehose/internal/domain"
	"stockseer.ai/blueksy-firehose/internal/models"
)

func StartWebSocketClient(
	ctx context.Context,
	rules *domain.RuleFactory,
) {
	// get our app config and logger
	appCtx, _ := appcontext.AppContextFromContext(ctx)
	cfg := appCtx.Config
	logger := appCtx.Log

	uri := cfg.JetstreamURL
	for {
		logger.Info("Connecting to WebSocket...")
		conn, resp, err := websocket.DefaultDialer.Dial(uri, http.Header{})
		if resp != nil {
			if err := resp.Body.Close(); err != nil {
				logger.Error("Failed to close response body", err)
			}
		}
		if err != nil {
			logger.Error("failed to connect to WebSocket", err)
			time.Sleep(5 * time.Second) // Wait before retrying
			continue
		}

		logger.Info("Connected to WebSocket: %s", uri)

		ConsumeMessages(appCtx, conn, rules)

		// If ConsumeMessages returns, it means the connection was closed
		logger.Info("Connection closed, attempting to reconnect...")
		time.Sleep(5 * time.Second) // Wait before attempting to reconnect
	}
}

func ConsumeMessages(
	appCtx appcontext.AppContext,
	conn *websocket.Conn,
	rules *domain.RuleFactory,
) {
	logger := appCtx.Log

	// add data collector for metrics collection
	dc := domain.DataCollector{}
	metricsErr := dc.StartMetrics(appCtx, false)

	if metricsErr != nil {
		logger.Error("failed to start metrics collection...", metricsErr)
	}

	for {
		if _, _, err := conn.NextReader(); err != nil {
			logger.Error("error reading message: %v", err)
			if err := conn.Close(); err != nil {
				logger.Error("failed to close connection: %v", err)
			}
			return
		}

		_, v, _ := conn.ReadMessage()

		var m models.ProtoMessage
		// Create a new Unmarshaler from the jsonpb package
		unmarshaler := jsonpb.Unmarshaler{}

		// Use the Unmarshaler's Unmarshal method. It takes an io.Reader.
		err := unmarshaler.Unmarshal(strings.NewReader(string(v)), &m)
		if err != nil {
			logger.Debug("failed to unmarshal JSON with jsonpb: %v", string(v))
			logger.Error("failed to unmarshal JSON with jsonpb: %v", err)
		}

		text := ""
		if m.Commit != nil && m.Commit.Record != nil {
			text = m.Commit.Record.Text
		}
		passed, _ := rules.EvaluateAll(text, &m)
		if passed {
			if err := dc.Add(&m); err != nil {
				logger.Error("failed to add message", err)
			}
			json, _ := m.ToJSON()
			err := appCtx.MQTTClient.PublishJSONToMQTT(appCtx.Config.MQTTMessagesTopic, json)
			if err != nil {
				logger.Error("failed to publish message", err)
			}
		}
	}
}
