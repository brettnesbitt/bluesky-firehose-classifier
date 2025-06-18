package mqtt

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"stockseer.ai/blueksy-firehose/internal/appcontext"
	"stockseer.ai/blueksy-firehose/internal/config" // Assuming config.AppConfig is defined here
	"stockseer.ai/blueksy-firehose/internal/domain"
	"stockseer.ai/blueksy-firehose/internal/interfaces" // Assuming interfaces.MqttClient is defined here
	"stockseer.ai/blueksy-firehose/internal/models"
)

// MqttClient encapsulates the MQTT client and its associated state.
type MqttClient struct {
	client     mqtt.Client
	appCtx     appcontext.AppContext
	processors *domain.TextProcessorFactory
	config     *config.AppConfig // Corrected type to direct pointer if it's already a pointer in AppContext
	clientID   string

	desiredSubscriptions map[string]byte
	mu                   sync.Mutex // Mutex to protect access to client and desiredSubscriptions
	isConnected          bool
	dataCollector        *domain.DataCollector // Moved data collector to the MqttClient

	// Channel to signal that the initial connection has been established.
	// This helps synchronize the main application with the MQTT connection.
	connectOnce sync.Once
	connectedCh chan struct{}
}

// Ensure MqttClient implements the interfaces.MqttClient interface.
var _ interfaces.MqttClient = (*MqttClient)(nil)

// NewMqttClient creates and initializes a new MqttClient instance.
func NewMqttClient(
	appCtx appcontext.AppContext,
	clientID string,
	processors *domain.TextProcessorFactory,
) (*MqttClient, error) {
	if !appCtx.Config.MQTTEnabled {
		appCtx.Log.Info("MQTT is disabled in configuration")
		// Consider returning an error here if MQTT being disabled is a critical misconfiguration
		return nil, nil
	}

	mc := &MqttClient{
		appCtx:               appCtx,
		processors:           processors,
		config:               &appCtx.Config, // Assuming appCtx.Config is already *config.AppConfig
		clientID:             clientID,
		desiredSubscriptions: make(map[string]byte),
		dataCollector:        domain.NewDataCollector(appCtx),
		connectedCh:          make(chan struct{}), // Initialize the channel
	}

	broker := mc.config.MQTTBrokerURL
	options := mqtt.NewClientOptions().AddBroker(broker)
	options.SetClientID("bluesky-client-" + mc.config.MQTTUsername + "-" + clientID)
	options.SetUsername(mc.config.MQTTUsername)
	options.SetPassword(mc.config.MQTTPassword)

	options.SetKeepAlive(60 * time.Second)
	options.SetConnectTimeout(30 * time.Second)

	// Set up MQTT client callbacks
	options.OnConnect = mc.onConnect
	options.OnConnectionLost = mc.onConnectionLost
	options.SetAutoReconnect(true) // Enable auto-reconnect

	mc.client = mqtt.NewClient(options)

	// Start metrics collection once when the client is initialized
	if metricsErr := mc.dataCollector.StartMetrics(appCtx, true); metricsErr != nil {
		appCtx.Log.Error("failed to start metrics collection...", metricsErr)
		// Decide if this is a fatal error or just a warning for your application
	}

	appCtx.Log.Info("Connecting to MQTT broker...")
	appCtx.Log.Info("Broker URL: %s", broker)
	appCtx.Log.Info("Username: %s", mc.config.MQTTUsername)
	appCtx.Log.Info("Metrics Topic: %s", mc.config.MQTTMetricsTopic)
	appCtx.Log.Info("Messages Topic: %s", mc.config.MQTTMessagesTopic)

	// Connect asynchronously to avoid blocking NewMqttClient
	// The `onConnect` callback will handle the initial subscriptions.
	go func() {
		if token := mc.client.Connect(); token.Wait() && token.Error() != nil {
			mc.appCtx.Log.Error("Initial MQTT connection failed: %v", token.Error())
			// This goroutine could potentially block or retry here if you want to
			// make connection a blocking prerequisite for NewMqttClient success.
			// For now, it just logs and relies on AutoReconnect.
		}
	}()

	return mc, nil
}

// onConnect is the callback executed when the MQTT client connects or reconnects.
func (mc *MqttClient) onConnect(c mqtt.Client) {
	mc.appCtx.Log.Info("Connected to MQTT broker")
	mc.mu.Lock()
	mc.isConnected = true
	// Resubscribe to all desired topics
	for topic, qos := range mc.desiredSubscriptions {
		mc.appCtx.Log.Info("Resubscribing to topic: %s", topic)
		// It's crucial that subscribeToTopic internally handles the actual subscription.
		// We pass the already-bound messageHandler for the topic.
		if err := mc.subscribeInternal(topic, qos, mc.messageHandler(topic)); err != nil {
			msg := fmt.Sprintf("Failed to resubscribe to topic %s: %v", topic, err)
			mc.appCtx.Log.Error(msg, err)
		}
	}
	mc.mu.Unlock()

	// Signal that the initial connection has been made
	// This ensures `connectedCh` is closed exactly once.
	mc.connectOnce.Do(func() {
		close(mc.connectedCh)
	})
}

// onConnectionLost is the callback executed when the MQTT client loses its connection.
func (mc *MqttClient) onConnectionLost(c mqtt.Client, err error) {
	mc.appCtx.Log.Error("Connection lost: %v", err)
	mc.mu.Lock()
	mc.isConnected = false
	mc.mu.Unlock()
	// Auto-reconnect will handle re-establishing the connection.
	// No explicit reconnect loop needed here.
}

// IsConnected returns the connection status of the MQTT client.
func (mc *MqttClient) IsConnected() bool {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	return mc.isConnected && mc.client.IsConnected()
}

// Disconnect disconnects the MQTT client.
func (mc *MqttClient) Disconnect() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if mc.client != nil && mc.client.IsConnected() {
		mc.appCtx.Log.Info("Disconnecting MQTT client...")
		mc.client.Disconnect(250)
		mc.isConnected = false
	}
}

// PublishToMQTT publishes a protobuf message to the specified topic.
func (mc *MqttClient) PublishToMQTT(topic string, msg *models.ProtoMessage) error {
	if !mc.IsConnected() {
		return fmt.Errorf("MQTT client not connected, cannot publish to topic %s", topic)
	}

	data, err := proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("error marshaling protobuf message: %w", err)
	}

	if token := mc.client.Publish(topic, byte(0), false, data); token.Wait() &&
		token.Error() != nil {
		return fmt.Errorf("error publishing to MQTT topic %s: %w", topic, token.Error())
	}
	return nil
}

// PublishJSONToMQTT publishes a JSON string to the specified topic.
func (mc *MqttClient) PublishJSONToMQTT(topic string, msg string) error {
	if !mc.IsConnected() {
		return fmt.Errorf("MQTT client not connected, cannot publish to topic %s", topic)
	}

	if token := mc.client.Publish(topic, 0, false, []byte(msg)); token.Wait() &&
		token.Error() != nil {
		return fmt.Errorf("error publishing JSON to MQTT topic %s: %w", topic, token.Error())
	}
	return nil
}

// processMessage handles the processing of an MQTT message and returns the processed message.
func (mc *MqttClient) processMessage(msg mqtt.Message) *models.ProtoMessage {
	jsonStr := string(msg.Payload())

	var protoMessage models.ProtoMessage
	if err := protojson.Unmarshal([]byte(jsonStr), &protoMessage); err != nil {
		mc.appCtx.Log.Error("Error parsing JSON message: %v", err)
		return nil
	}

	processed, err := mc.processors.ProcessAll(protoMessage.Commit.Record.Text)
	if err != nil {
		mc.appCtx.Log.Error("Failed to process message: %v", err)
		return nil
	}

	cleaned := regexp.MustCompile(`[\W_]+`).
		ReplaceAllString(processed["TextCategoryClassifier"], " ")
	protoMessage.Categories = strings.Fields(cleaned)

	finSentiment := processed["TextFinSentimentClassifier"]
	protoMessage.FinSentiment = &finSentiment

	if domain.ShouldStoreMessage(&protoMessage) {
		if err := mc.appCtx.MessageRepo.Insert(&protoMessage); err != nil {
			mc.appCtx.Log.Error("Failed to insert message: %v", err)
		}
	}
	return &protoMessage
}

// messageHandler creates a message handler for a specific topic.
func (mc *MqttClient) messageHandler(topic string) mqtt.MessageHandler {
	return func(client mqtt.Client, msg mqtt.Message) {
		protoMessage := mc.processMessage(msg)
		if protoMessage != nil {
			if err := mc.dataCollector.Add(protoMessage); err != nil {
				mc.appCtx.Log.Error("Failed to add message to data collector: %v", err)
			}
		}
	}
}

// subscribeInternal is the internal function that performs the MQTT subscription.
// It assumes the mutex is already locked if called from onConnect, or locks it if called directly.
func (mc *MqttClient) subscribeInternal(topic string, qos byte, handler mqtt.MessageHandler) error {
	if token := mc.client.Subscribe(topic, qos, handler); token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to subscribe to topic %s: %w", topic, token.Error())
	}
	mc.appCtx.Log.Info("Successfully subscribed to topic: %s", topic)
	return nil
}

// Subscribe adds a topic to the desired subscriptions and attempts to subscribe if connected.
// This is the public method for external calls to subscribe.
func (mc *MqttClient) Subscribe(topic string, qos byte) error {
	mc.mu.Lock()
	// Always add to desired subscriptions, so it's picked up on reconnect
	mc.desiredSubscriptions[topic] = qos
	mc.mu.Unlock()

	// Create handler specific to this topic
	handler := mc.messageHandler(topic)

	// Attempt to subscribe immediately if already connected.
	// Otherwise, onConnect will handle it.
	if mc.IsConnected() { // Uses IsConnected to safely check connection status
		mc.mu.Lock() // Lock before calling internal subscribe method
		defer mc.mu.Unlock()
		return mc.subscribeInternal(topic, qos, handler)
	} else {
		mc.appCtx.Log.Warn("Not currently connected. Topic %s will be subscribed on successful MQTT connection.", topic)
		return nil // Not an error, just means subscription will happen later
	}
}

// ConsumeMessages sets up the subscription for consuming messages.
// This function should ideally just *register* the topic for consumption.
// The actual subscription happens in onConnect.
func (mc *MqttClient) ConsumeMessages() error {
	topic := mc.config.MQTTMessagesTopic
	qos := byte(1) // QoS 1 for "at least once" delivery

	// Register the subscription. The Subscribe method will handle the actual logic.
	// It will add to desiredSubscriptions and attempt immediate subscription if connected.
	if err := mc.Subscribe(topic, qos); err != nil {
		return fmt.Errorf("failed to register message consumption for topic %s: %w", topic, err)
	}

	mc.appCtx.Log.Info("MQTT message consumption setup initiated for topic: %s", topic)

	// Wait for the initial connection to be established before proceeding.
	// This helps ensure subscriptions are registered before the main loop starts.
	select {
	case <-mc.connectedCh:
		mc.appCtx.Log.Info("Initial MQTT connection established, consumer is active.")
	case <-time.After(30 * time.Second): // Add a timeout for the initial connection
		return fmt.Errorf("timed out waiting for initial MQTT connection for consumer")
	}

	// Keep the function running to listen for messages.
	// This will block forever, which is typical for a long-running service.
	// It's important that this is called in a goroutine if your main function needs to do other things.
	select {}
}
