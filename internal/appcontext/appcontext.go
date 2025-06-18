package appcontext

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"stockseer.ai/blueksy-firehose/internal/config"
	"stockseer.ai/blueksy-firehose/internal/interfaces"
	"stockseer.ai/blueksy-firehose/internal/logger"
	"stockseer.ai/blueksy-firehose/internal/repositories"
)

// AppContext holds the application config and logger.

// AppContext holds the application config, logger, and repositories.
type AppContext struct {
	Config      config.AppConfig
	Log         logger.Logger
	MongoClient *mongo.Client
	MQTTClient  interfaces.MqttClient
	MetricsRepo repositories.Repository
	MessageRepo repositories.Repository
}

// NewAppContext creates a new AppContext.
func NewAppContext(
	config *config.AppConfig,
	wssReader bool,
	mqttClient interfaces.MqttClient,
) AppContext {
	log := logger.NewLogger()

	var metricsRepo repositories.Repository
	var messageRepo repositories.Repository
	var client *mongo.Client
	// Initialize repositories
	if !wssReader {
		// Connect to MongoDB
		clientOptions := options.Client().ApplyURI(config.MongoURI)
		client, err := mongo.Connect(context.Background(), clientOptions)
		if err != nil {
			panic(fmt.Sprintf("Failed to connect to MongoDB: %s", err))
		}

		metricsRepo = repositories.NewMongoRepository(client, "blueskyfh", "metrics")
		messageRepo = repositories.NewMongoRepository(client, "blueskyfh", "messages")
	}

	return AppContext{
		Config:      *config,
		Log:         log,
		MongoClient: client,
		MetricsRepo: metricsRepo,
		MessageRepo: messageRepo,
		MQTTClient:  mqttClient,
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
