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
	server "stockseer.ai/blueksy-firehose/internal/transport/http"
	"stockseer.ai/blueksy-firehose/internal/transport/mqtt"
)

func main() {
	// Load our configuration on start up.
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(fmt.Sprintf("Failed to load configuration: %s ", err))
	}

	// initialize our processors that govern what data to consume
	processors := domain.InitProcessors(cfg)

	// Create full app context with MongoDB connection first
	appContext := appcontext.NewAppContext(cfg, false, nil) // Create context with MongoDB first
	log := appContext.Log

	// Initialize MQTT client if enabled, ensures that we can connect...
	mqttClient, err := mqtt.NewMqttClient(
		appContext,
		"server",
		processors,
	) // Pass full app context with MongoDB
	if err != nil {
		log.Error("Failed to initialize MQTT client: %v", err)
		os.Exit(1)
	}
	defer mqttClient.Disconnect() // Ensure client disconnects on exit

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

	// Start our server in a separate goroutine.
	go func() {
		servererr := server.StartServer(ctx)

		if servererr != nil {
			log.Error("server failed to start", servererr)
		}
	}()

	// consume messages from MQTT
	if err := mqttClient.ConsumeMessages(); err != nil { // Call directly on the concrete instance
		appContext.Log.Error("Failed to start MQTT message consumption: %v", err)
		os.Exit(1)
	}

	// Wait forever to keep the main goroutine running
	select {}
}
