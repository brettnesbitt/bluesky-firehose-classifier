package domain

import "stockseer.ai/blueksy-firehose/internal/models"

// shouldStoreMessage checks if a message has a positive or negative financial sentiment.
func ShouldStoreMessage(message *models.ProtoMessage) bool {
	return message.FinSentiment != nil &&
		(*message.FinSentiment == "positive" || *message.FinSentiment == "negative")
}
