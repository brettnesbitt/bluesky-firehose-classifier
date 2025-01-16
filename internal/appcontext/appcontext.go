package appcontext

import (
	"context"

	"stockseer.ai/blueksy-firehose/internal/config"
	"stockseer.ai/blueksy-firehose/internal/logger"
)

// AppContext holds the application config and logger.
type AppContext struct {
	Config config.AppConfig
	Log    logger.Logger
}

// NewAppContext creates a new AppContext.
func NewAppContext(config *config.AppConfig) AppContext {
	log := logger.NewLogger()
	return AppContext{
		Config: *config,
		Log:    log,
	}
}

// ContextWithAppContext adds the AppContext to the context.
func ContextWithAppContext(ctx context.Context, appContext AppContext) context.Context {
	return context.WithValue(ctx, appContextKey{}, appContext)
}

// AppContextFromContext retrieves the AppContext from the context.
func AppContextFromContext(ctx context.Context) (AppContext, bool) {
	appContext, ok := ctx.Value(appContextKey{}).(AppContext)
	return appContext, ok
}

type appContextKey struct{}
