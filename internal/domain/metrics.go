package domain

import (
	"time"

	"stockseer.ai/blueksy-firehose/internal/appcontext"
)

// Categories we want to track sentiment for.
var trackedCategories = []string{"labour", "politics", "economy", "conflict"}

// DataCollector struct to hold data.
type DataCollector struct {
	AppCtx    appcontext.AppContext
	data      []Message
	lastIdx   int // last metric idx observed
	callCount int // number of calls to log metrics
}

// shouldStoreMessage checks if a message has a positive or negative financial sentiment.
func shouldStoreMessage(message Message) bool {
	return message.FinSentiment == "positive" || message.FinSentiment == "negative"
}

// AddData adds a new data point to the collector.
func (dc *DataCollector) Add(message Message) error {
	if shouldStoreMessage(message) {
		err := dc.AppCtx.MessageRepo.Insert(message)
		if err != nil {
			return err
		}
	}
	// Check the length and pop the oldest entry if needed
	if len(dc.data) >= 100 {
		dc.data = dc.data[1:] // Remove the first element
	}
	dc.data = append(dc.data, message)

	return nil
}

// Clears data from DataCollector.
func (dc *DataCollector) Clear() {
	dc.data = dc.data[:0]
}

// Returns the number of messages that are new posts.
type PostMetrics struct {
	Total     int
	SinceLast int
	Timestamp int64
}

// Print outputs the contents of PostMetrics.
func (pm *PostMetrics) Print(appCtx appcontext.AppContext) {
	appCtx.Log.Info("Posts(sec): %d      Total Posts: %d", pm.SinceLast, pm.Total)
}

func (dc *DataCollector) GetPostFrequency() PostMetrics {
	sinceLast := len(dc.data) - dc.lastIdx
	total := len(dc.data)
	return PostMetrics{
		Total:     total,
		SinceLast: sinceLast,
		Timestamp: time.Now().Unix(),
	}
}

// Returns the number of messages that are new posts.
func (dc *DataCollector) GetPostTokenCount() (total int, sinceLastCall int) {
	total = 0
	sinceLastCall = 0

	tokens := ""

	for idx, data := range dc.data {
		total += len(data.Commit.Record.Text)
		if idx >= dc.lastIdx {
			sinceLastCall += len(data.Commit.Record.Text)
		}
		tokens += "\n\n****POST****\n"
		tokens += data.Commit.Record.Text
	}

	return
}

// Helper function to check if a category exists in the categories slice.
func containsCategory(categories []string, category string) bool {
	for _, cat := range categories {
		if cat == category {
			return true
		}
	}
	return false
}

type CategoryMetrics struct {
	Negative  int
	Positive  int
	Category  string
	Timestamp int64
}

// Print outputs the contents of CategoryMetrics.
func (cm *CategoryMetrics) Print(appCtx appcontext.AppContext) {
	appCtx.Log.Info("%-12s: Negative: %v      Positive: %v", cm.Category, cm.Negative, cm.Positive)
}

func (dc *DataCollector) GetSentimentFrequency(category string) CategoryMetrics {
	sinceLastCall := make(map[string]int)

	for idx, data := range dc.data {
		if data.FinSentiment != "" && containsCategory(data.Categories, category) {
			if idx >= dc.lastIdx {
				sinceLastCall[data.FinSentiment]++
			}
		}
	}

	return CategoryMetrics{
		Negative:  sinceLastCall["negative"],
		Positive:  sinceLastCall["positive"],
		Category:  category,
		Timestamp: time.Now().Unix(),
	}
}

func (dc *DataCollector) LogMetrics() {
	postMetrics := dc.GetPostFrequency()
	totalTokens, currentTokens := dc.GetPostTokenCount()

	// Print basic metrics
	postMetrics.Print(dc.AppCtx)
	const logFormat = "Current Tokens: %d     Total Tokens: %d       Calls: %d"
	dc.AppCtx.Log.Info(logFormat, currentTokens, totalTokens, dc.callCount)

	// Print sentiment frequencies for each tracked category
	for _, category := range trackedCategories {
		metrics := dc.GetSentimentFrequency(category)

		metrics.Print(dc.AppCtx)

		err := dc.AppCtx.MetricsRepo.Insert(metrics)
		if err != nil {
			dc.AppCtx.Log.Error("Error storing metrics", err)
		}
	}

	dc.lastIdx = len(dc.data)
	dc.callCount += 1
}

func (dc *DataCollector) StartMetrics(appCtx appcontext.AppContext) error {
	dc.AppCtx = appCtx // Ensure AppCtx is assigned
	dc.AppCtx.Log.Info("Metrics collection started")

	ticker := time.NewTicker(60 * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				dc.LogMetrics()
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

	return nil
}
