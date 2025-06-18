package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"stockseer.ai/blueksy-firehose/internal/appcontext"
	httphandlers "stockseer.ai/blueksy-firehose/internal/handlers/http"
)

// StartServer starts the Gin server.
func StartServer(ctx context.Context) error {
	// get our app config and logger
	appCtx, _ := appcontext.AppContextFromContext(ctx)
	cfg := appCtx.Config
	logger := appCtx.Log

	// Create a default gin router
	router := gin.Default()

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.ServerPort)

	httphandlers.RegisterHealthzRoutes(router)

	srv := &http.Server{
		Addr:              addr,
		Handler:           router.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	// Start the server
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("listener: ", err)
		}
	}()

	<-ctx.Done()
	logger.Info("Shutting down server...")

	// Graceful shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Fatal("Server forced to shutdown:", err)
	}

	return nil
}
