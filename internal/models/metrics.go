package models

import (
	"encoding/json"

	"stockseer.ai/blueksy-firehose/internal/logger"
)

// CategoryMetrics holds sentiment counts for a specific category.
type CategoryMetrics struct {
	Negative  int    `json:"negative"`
	Positive  int    `json:"positive"`
	Category  string `json:"category"`
	Timestamp int64  `json:"timestamp"`
}

// ToJSON marshals the CategoryMetrics struct to a JSON string.
func (cm *CategoryMetrics) ToJSON() (string, error) {
	jsonData, err := json.Marshal(cm)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

func (cm *CategoryMetrics) Print(logger logger.Logger) {
	logger.Info(
		"Category: %s, Negative: %d, Positive: %d, Timestamp: %d",
		cm.Category,
		cm.Negative,
		cm.Positive,
		cm.Timestamp,
	)
}
