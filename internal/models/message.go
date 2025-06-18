package models

import (
	"encoding/json"
	"fmt"
	"time"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// JSONMessage is an interface that provides JSON serialization/deserialization methods
// for protobuf messages.
type JSONMessage interface {
	ToJSON() (string, error)
	FromJSON(jsonStr string) error
	ToJSONMap() (map[string]interface{}, error)
	FromJSONMap(data map[string]interface{}) error
}

// ToJSON converts a protobuf message to a JSON string.
func ToJSON(msg proto.Message) (string, error) {
	if msg == nil {
		return "", fmt.Errorf("message is nil")
	}

	marshaler := protojson.MarshalOptions{
		UseProtoNames:   true,
		UseEnumNumbers:  false,
		EmitUnpopulated: true, // Changed to true to ensure all fields are included
	}

	jsonBytes, err := marshaler.Marshal(msg)
	if err != nil {
		return "", fmt.Errorf("failed to marshal to JSON: %w", err)
	}

	return string(jsonBytes), nil
}

// FromJSON converts a JSON string to a protobuf message.
func FromJSON(msg proto.Message, jsonStr string) error {
	if msg == nil {
		return fmt.Errorf("message is nil")
	}

	unmarshaler := protojson.UnmarshalOptions{
		DiscardUnknown: false,
	}

	err := unmarshaler.Unmarshal([]byte(jsonStr), msg)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return nil
}

// ToJSONMap converts a protobuf message to a JSON map.
func ToJSONMap(msg proto.Message) (map[string]interface{}, error) {
	if msg == nil {
		return nil, fmt.Errorf("message is nil")
	}

	jsonStr, err := ToJSON(msg)
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	err = json.Unmarshal([]byte(jsonStr), &data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal to map: %w", err)
	}

	return data, nil
}

// FromJSONMap converts a JSON map to a protobuf message.
func FromJSONMap(msg proto.Message, data map[string]interface{}) error {
	if msg == nil {
		return fmt.Errorf("message is nil")
	}

	jsonStr, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal map to JSON: %w", err)
	}

	return FromJSON(msg, string(jsonStr))
}

// WithDateTime adds a DateTime field to the map if TimeUs is present.
// TimeUs is in microseconds since the epoch.
func (m *ProtoMessage) WithDateTime() (*map[string]interface{}, error) {
	// Convert the message to a map first
	data, err := ToJSONMap(m)
	if err != nil {
		return nil, fmt.Errorf("failed to convert message to map: %w", err)
	}

	// Add DateTime field if TimeUs is present
	data["datetime"] = time.UnixMicro(m.TimeUs).String()

	return &data, nil
}

// Implement JSONMessage interface for all protobuf messages.
func (m *ProtoMessage) ToJSON() (string, error) {
	return ToJSON(m)
}

func (m *ProtoMessage) FromJSON(jsonStr string) error {
	return FromJSON(m, jsonStr)
}

func (m *ProtoMessage) ToJSONMap() (map[string]interface{}, error) {
	return ToJSONMap(m)
}

func (m *ProtoMessage) FromJSONMap(data map[string]interface{}) error {
	return FromJSONMap(m, data)
}
