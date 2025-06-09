package domain

import (
	"encoding/json"
	"strings"
	"sync"
	"time"

	"stockseer.ai/blueksy-firehose/internal/appcontext"
	"stockseer.ai/blueksy-firehose/internal/transport/mqtt"
)

// Categories we want to track sentiment for.
var trackedCategories = []string{"labour", "politics", "economy", "conflict"}

var intervalSecs = 60

// DataCollector safely collects and aggregates metrics from concurrent sources.
type DataCollector struct {
	AppCtx appcontext.AppContext

	// Mutex to protect concurrent access to the fields below
	mu sync.RWMutex

	// --- Overall Metrics ---
	totalPosts  int64
	totalTokens int64

	// --- "Since Last Log" Metrics ---
	// These are aggregated on Add and reset on LogMetrics
	postsSinceLog       int
	tokensSinceLog      int
	sentimentSinceLog   map[string]*CategoryMetrics
	periodicCallCounter int
}

// NewDataCollector creates and initializes a new DataCollector.
func NewDataCollector(appCtx appcontext.AppContext) *DataCollector {
	return &DataCollector{
		AppCtx:            appCtx,
		sentimentSinceLog: make(map[string]*CategoryMetrics),
	}
}

// shouldStoreMessage checks if a message has a positive or negative financial sentiment.
func shouldStoreMessage(message Message) bool {
	return message.FinSentiment == "positive" || message.FinSentiment == "negative"
}

// Add safely adds a new data point, updating aggregated metrics.
func (dc *DataCollector) Add(message Message) error {
	// Store the message first if it meets criteria
	if shouldStoreMessage(message) {
		err := dc.AppCtx.MessageRepo.Insert(message)
		if err != nil {
			return err // Return early on storage error
		}
		if dc.AppCtx.Config.MQTTEnabled {
			jsonString, err := message.ToJSON()
			if err != nil {
				return err // Return early on JSON error
			}
			mqtt.PublishToMQTT(dc.AppCtx, dc.AppCtx.Config.MQTTMessagesTopic, jsonString)
		}
	}

	// --- Lock for the remainder of the function to update metrics safely ---
	dc.mu.Lock()
	defer dc.mu.Unlock()

	// Update overall and periodic post counts
	dc.totalPosts++
	dc.postsSinceLog++

	// Update overall and periodic token counts
	// strings.Fields is a simple, effective way to count words/tokens.
	messageTokens := len(strings.Fields(message.Commit.Record.Text))
	dc.totalTokens += int64(messageTokens)
	dc.tokensSinceLog += messageTokens

	// Update sentiment counts for relevant categories
	for _, category := range message.Categories {
		// Only track categories we care about
		if !containsCategory(trackedCategories, category) {
			continue
		}

		if _, ok := dc.sentimentSinceLog[category]; !ok {
			dc.sentimentSinceLog[category] = &CategoryMetrics{Category: category}
		}

		switch message.FinSentiment {
		case "positive":
			dc.sentimentSinceLog[category].Positive++
		case "negative":
			dc.sentimentSinceLog[category].Negative++
		}
	}

	return nil
}

// LogMetrics logs the currently aggregated metrics and resets the periodic counters.
func (dc *DataCollector) LogMetrics() {
	// Lock for the entire duration to ensure a consistent snapshot of metrics is logged and reset
	dc.mu.Lock()
	defer dc.mu.Unlock()

	dc.periodicCallCounter++

	// --- Log Post and Token Metrics ---
	dc.AppCtx.Log.Info("Posts(sec): %d      Total Posts: %d", dc.postsSinceLog/intervalSecs, dc.totalPosts)
	dc.AppCtx.Log.Info("Tokens(sec): %d     Total Tokens: %d       Calls: %d", dc.tokensSinceLog/intervalSecs, dc.totalTokens, dc.periodicCallCounter)

	// --- Log Sentiment Metrics for each tracked category ---
	for _, category := range trackedCategories {
		metrics, ok := dc.sentimentSinceLog[category]
		if !ok {
			// If no posts for this category were seen, log zeroes to indicate it's still being tracked.
			metrics = &CategoryMetrics{Category: category}
		}

		metrics.Timestamp = time.Now().Unix()
		metrics.Print(dc.AppCtx)

		// Persist and publish metrics
		err := dc.AppCtx.MetricsRepo.Insert(*metrics)
		if err != nil {
			dc.AppCtx.Log.Error("Error storing metrics", err)
		}
		if dc.AppCtx.Config.MQTTEnabled {
			jsonString, err := metrics.ToJSON()
			if err != nil {
				dc.AppCtx.Log.Error("Error converting metrics to JSON", err)
			} else {
				mqtt.PublishToMQTT(dc.AppCtx, dc.AppCtx.Config.MQTTMetricsTopic, jsonString)
			}
		}
	}

	// --- CRITICAL: Reset the "since last log" counters for the next interval ---
	dc.postsSinceLog = 0
	dc.tokensSinceLog = 0
	// Clear the map for the next batch of metrics
	dc.sentimentSinceLog = make(map[string]*CategoryMetrics)
}

// StartMetrics begins the periodic logging of collected metrics.
func (dc *DataCollector) StartMetrics(appCtx appcontext.AppContext) error {
	dc.AppCtx = appCtx
	dc.AppCtx.Log.Info("Metrics collection started")

	ticker := time.NewTicker(time.Duration(intervalSecs) * time.Second)
	// Immediately log once on startup without waiting for the first tick
	go dc.LogMetrics()

	go func() {
		for {
			// This select statement is blocking, so it's safe to not have a quit channel
			// if this goroutine is meant to run for the lifetime of the application.
			<-ticker.C
			dc.LogMetrics()
		}
	}()

	return nil
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

// Print outputs the contents of CategoryMetrics.
func (cm *CategoryMetrics) Print(appCtx appcontext.AppContext) {
	appCtx.Log.Info("%-12s: Negative: %v      Positive: %v", cm.Category, cm.Negative, cm.Positive)
}
