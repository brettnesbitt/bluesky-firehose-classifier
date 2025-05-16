package mqtt

import (
	"os"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"stockseer.ai/blueksy-firehose/internal/appcontext"
)

var mqttClient mqtt.Client

func InitMQTTClient(appCtx appcontext.AppContext) {
	cfg := appCtx.Config

	if !cfg.MQTTEnabled {
		appCtx.Log.Info("MQTT is disabled in configuration")
		return
	}

	broker := cfg.MQTTBrokerURL
	options := mqtt.NewClientOptions().AddBroker(broker)
	options.SetClientID("bluesky-client")
	options.SetUsername(cfg.MQTTUsername)
	options.SetPassword(cfg.MQTTPassword)

	// Set up MQTT client callbacks
	options.OnConnect = func(c mqtt.Client) {
		appCtx.Log.Info("Connected to MQTT broker")
	}
	options.OnConnectionLost = func(c mqtt.Client, err error) {
		appCtx.Log.Info("Connection lost: %v\n", err)
	}

	appCtx.Log.Info("Connecting to MQTT broker...")
	appCtx.Log.Info("Broker URL: %s", broker)
	appCtx.Log.Info("Username: %s", cfg.MQTTUsername)
	appCtx.Log.Info("Metrics Topic: %s", cfg.MQTTMetricsTopic)
	appCtx.Log.Info("Messages Topic: %s", cfg.MQTTMessagesTopic)

	mqttClient = mqtt.NewClient(options)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		appCtx.Log.Error("Error connecting to MQTT broker: %v\n", token.Error())
		os.Exit(1)
	}
}

func PublishToMQTT(appCtx appcontext.AppContext, topic, message string) {
	if mqttClient == nil {
		InitMQTTClient(appCtx)
	}
	if token := mqttClient.Publish(topic, 0, false, message); token.Wait() && token.Error() != nil {
		appCtx.Log.Error("Error publishing to MQTT: %v\n", token.Error())
	}
}
