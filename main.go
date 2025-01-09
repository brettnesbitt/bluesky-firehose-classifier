package main

import (
	"context"

	"stockseer.ai/blueksy-firehose/internal/config"
	"stockseer.ai/blueksy-firehose/internal/domain"
	"stockseer.ai/blueksy-firehose/internal/logger"
	server "stockseer.ai/blueksy-firehose/internal/transport/http"
	"stockseer.ai/blueksy-firehose/internal/transport/ws"
)

func main() {

	// Load our configuration on start up.
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatal("Failed to load configuration: ", err)
	}

	// Create a context with cancellation.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start our server in a separate goroutine.
	go func() {
		servererr := server.StartServer(ctx, cfg)

		if servererr != nil {
			logger.Error("server failed to start", servererr)
		}
	}()

	// initialize our rules that govern what data to consume
	rules := domain.InitRules(cfg)
	processors := domain.InitProcessors(cfg)

	// start our web socket client to receive messages with our rules
	ws.StartWebSocketClient(cfg, rules, processors)
}
