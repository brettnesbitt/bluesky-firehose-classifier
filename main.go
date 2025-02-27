package main

import (
	"context"
	"fmt"
	"strconv"

	"github.com/rs/zerolog"
	"stockseer.ai/blueksy-firehose/internal/appcontext"
	"stockseer.ai/blueksy-firehose/internal/config"
	"stockseer.ai/blueksy-firehose/internal/domain"
	server "stockseer.ai/blueksy-firehose/internal/transport/http"
	"stockseer.ai/blueksy-firehose/internal/transport/ws"
)

func main() {

	// Load our configuration on start up.
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(fmt.Sprintf("Failed to load configuration: %s ", err))
	}

	appContext := appcontext.NewAppContext(cfg)
	log := appContext.Log

	if cfg.DevMode {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	log.Info(fmt.Sprintf("DEV MODE: %s", strconv.FormatBool(cfg.DevMode)))

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

	// initialize our rules that govern what data to consume
	rules := domain.InitRules(cfg)
	processors := domain.InitProcessors(cfg)

	// start our web socket client to receive messages with our rules
	ws.StartWebSocketClient(ctx, rules, processors)
}
