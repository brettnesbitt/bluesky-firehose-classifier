package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/rs/zerolog"
	"stockseer.ai/blueksy-firehose/internal/appcontext"
	"stockseer.ai/blueksy-firehose/internal/config"
	"stockseer.ai/blueksy-firehose/internal/domain"
	"stockseer.ai/blueksy-firehose/internal/transport/mqtt"
	"stockseer.ai/blueksy-firehose/internal/transport/ws"
)

func main() {
	// Load our configuration on start up.
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(fmt.Sprintf("Failed to load configuration: %s ", err))
	}

	// Create full app context with MongoDB connection first
	appContext := appcontext.NewAppContext(cfg, false, nil) // Create context with MongoDB first
	log := appContext.Log

	// Initialize MQTT client if enabled, ensures that we can connect...
	mqttClient, err := mqtt.NewMqttClient(
		appContext,
		"server",
		nil,
	) // Pass full app context with MongoDB
	if err != nil {
		log.Error("Failed to initialize MQTT client: %v", err)
		os.Exit(1)
	}
	defer mqttClient.Disconnect() // Ensure client disconnects on exit

	// add mqtt client to app context
	appContext.MQTTClient = mqttClient

	if cfg.DevMode {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	const devModeFormat = "DEV MODE: %s"
	log.Info(devModeFormat, strconv.FormatBool(cfg.DevMode))

	// Create a context with cancellation.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctx = appcontext.ContextWithAppContext(ctx, appContext)

	// initialize our rules that govern what data to consume
	rules := domain.InitRules(cfg)

	log.Info("Starting WebSocket client...")
	// start our web socket client to receive messages with our rules
	ws.StartWebSocketClient(ctx, rules)
}
