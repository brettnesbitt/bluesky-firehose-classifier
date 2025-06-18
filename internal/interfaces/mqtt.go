package interfaces

import "stockseer.ai/blueksy-firehose/internal/models"

// MqttClient defines the interface for MQTT client operations that AppContext needs to expose.
type MqttClient interface {
	IsConnected() bool
	PublishToMQTT(topic string, msg *models.ProtoMessage) error
	PublishJSONToMQTT(topic string, msg string) error
	ConsumeMessages() error
}
