package domain

import (
	"fmt"
	"time"

	"stockseer.ai/blueksy-firehose/internal/logger"
)

// DataCollector struct to hold data.
type DataCollector struct {
	data      []Message
	lastIdx   int // last metric idx observed
	callCount int // number of calls to log metrics
}

// AddData adds a new data point to the collector.
func (dc *DataCollector) Add(message Message) {
	dc.data = append(dc.data, message)
}

// Clears data from DataCollector.
func (dc *DataCollector) Clear() {
	dc.data = dc.data[:0]
}

// Returns the number of messages that are new posts.
func (dc *DataCollector) GetPostFrequency() (total int, sinceLastCall int) {
	sinceLastCall = len(dc.data) - dc.lastIdx
	total = len(dc.data)
	return
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

	if dc.callCount == 60 {
		fmt.Println(tokens)
	}
	return
}

func (dc *DataCollector) LogMetrics() {
	totalPosts, postsSec := dc.GetPostFrequency()
	totalTokens, currentTokens := dc.GetPostTokenCount()
	fmt.Printf("\033[2K\rPosts(sec): %d      Current Tokens: %d     Total Posts: %d      Total Tokens: %d       Calls: %d", postsSec, currentTokens, totalPosts, totalTokens, dc.callCount)

	dc.lastIdx = len(dc.data)
	dc.callCount += 1
}

func (dc *DataCollector) StartMetrics() error {
	logger.Info("Metrics collection started")
	ticker := time.NewTicker(1 * time.Second)
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
