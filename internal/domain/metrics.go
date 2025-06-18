package domain

import (
	"strings"
	"sync"
	"time"

	"stockseer.ai/blueksy-firehose/internal/appcontext"
	"stockseer.ai/blueksy-firehose/internal/models"
)

// Categories we want to track sentiment for.

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
	sentimentSinceLog   map[string]*models.CategoryMetrics
	periodicCallCounter int
}

// NewDataCollector creates and initializes a new DataCollector.
func NewDataCollector(appCtx appcontext.AppContext) *DataCollector {
	return &DataCollector{
		AppCtx:              appCtx,
		totalPosts:          0,
		totalTokens:         0,
		postsSinceLog:       0,
		tokensSinceLog:      0,
		sentimentSinceLog:   make(map[string]*models.CategoryMetrics),
		periodicCallCounter: 0,
	}
}

// Add safely adds a new data point, updating aggregated metrics.
func (dc *DataCollector) Add(message *models.ProtoMessage) error {
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

	if message.FinSentiment != nil {
		// Update sentiment counts for relevant categories
		for _, category := range message.Categories {
			// Only track categories we care about
			if !containsCategory(trackedCategories, category) {
				continue
			}

			if _, ok := dc.sentimentSinceLog[category]; !ok {
				dc.sentimentSinceLog[category] = &models.CategoryMetrics{Category: category}
			}

			sentiment := *message.FinSentiment
			switch sentiment {
			case "positive":
				dc.sentimentSinceLog[category].Positive++
			case "negative":
				dc.sentimentSinceLog[category].Negative++
			default:
				dc.AppCtx.Log.Warn("Unknown sentiment value: %s", sentiment)
			}
		}
	}

	return nil
}

// LogMetrics logs the currently aggregated metrics and resets the periodic counters.
func (dc *DataCollector) LogMetrics(server bool) {
	// Lock for the entire duration to ensure a consistent snapshot of metrics is logged and reset
	dc.mu.Lock()
	defer dc.mu.Unlock()

	dc.periodicCallCounter++

	// --- Log Post and Token Metrics ---
	dc.AppCtx.Log.Info(
		"Posts(sec): %d      Total Posts: %d",
		dc.postsSinceLog/intervalSecs,
		dc.totalPosts,
	)
	dc.AppCtx.Log.Info(
		"Tokens(sec): %d     Total Tokens: %d       Calls: %d",
		dc.tokensSinceLog/intervalSecs,
		dc.totalTokens,
		dc.periodicCallCounter,
	)

	// --- Log Sentiment Metrics for each tracked category ---
	if server {
		for _, category := range trackedCategories {
			metrics, ok := dc.sentimentSinceLog[category]
			if !ok {
				// If no posts for this category were seen, log zeroes to indicate it's still being tracked.
				metrics = &models.CategoryMetrics{
					Category: category,
					Negative: 0,
					Positive: 0,
				}
				dc.sentimentSinceLog[category] = metrics
			}

			metrics.Timestamp = time.Now().Unix()
			metrics.Print(dc.AppCtx.Log)

			// Persist and publish metrics
			if err := dc.AppCtx.MetricsRepo.Insert(*metrics); err != nil {
				dc.AppCtx.Log.Error("Failed to insert metrics", err)
				return
			}
		}
	}

	// --- Reset the "since last log" counters for the next interval ---
	dc.postsSinceLog = 0
	dc.tokensSinceLog = 0
	// Reset sentiment metrics to zero but keep the map structure
	for category := range dc.sentimentSinceLog {
		dc.sentimentSinceLog[category] = &models.CategoryMetrics{
			Category: category,
			Negative: 0,
			Positive: 0,
		}
	}
}

// StartMetrics begins the periodic logging of collected metrics.
func (dc *DataCollector) StartMetrics(appCtx appcontext.AppContext, server bool) error {
	dc.AppCtx = appCtx
	dc.AppCtx.Log.Info("Metrics collection started")

	// Initialize sentimentSinceLog with empty metrics for tracked categories
	for _, category := range trackedCategories {
		dc.sentimentSinceLog[category] = &models.CategoryMetrics{
			Category: category,
			Negative: 0,
			Positive: 0,
		}
	}

	ticker := time.NewTicker(time.Duration(intervalSecs) * time.Second)
	// Immediately log once on startup without waiting for the first tick
	go dc.LogMetrics(server)

	go func() {
		for {
			// This select statement is blocking, so it's safe to not have a quit channel
			// if this goroutine is meant to run for the lifetime of the application.
			<-ticker.C
			dc.LogMetrics(server)
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
