package domain

import (
	"errors"

	"stockseer.ai/blueksy-firehose/internal/apis/mlclassifier"
	"stockseer.ai/blueksy-firehose/internal/config"
)

type TextProcessor struct {
	Description string
	Evaluate    func(text string) (string, error)
}

// helps create and manage processors.
type TextProcessorFactory struct {
	processors []TextProcessor
}

// creates a new factory.
func NewTextProcessorFactory() *TextProcessorFactory {
	return &TextProcessorFactory{}
}

// AddProcessor adds a new rule to the factory.
func (tpf *TextProcessorFactory) AddProcessor(description string, evaluate func(text string) (string, error)) {
	tpf.processors = append(tpf.processors, TextProcessor{Description: description, Evaluate: evaluate})
}

// and a map of rule descriptions to results.
func (tpf *TextProcessorFactory) ProcessAll(text string) (map[string]string, error) {
	results := make(map[string]string)
	for _, tp := range tpf.processors {
		val, _ := tp.Evaluate(text)
		results[tp.Description] = val
	}
	return results, nil
}

func InitProcessors(cfg *config.AppConfig) *TextProcessorFactory {
	tpf := NewTextProcessorFactory()

	if cfg.TextCategoryClassifier {
		tpf.AddProcessor("TextCategoryClassifier", func(text string) (string, error) {
			// api call to model
			data := mlclassifier.DataRequest{
				Items: []mlclassifier.DataRequestItem{
					{Text: text},
				},
			}
			client := mlclassifier.NewClient(cfg.TextCategoryClassifierURL)
			resp, err := client.Classify(data)

			if err != nil {
				return "", err
			}

			if resp == nil || (*resp)[0].Label == "" {
				return "", errors.New("response is nil in API call")
			}

			return (*resp)[0].Label, nil
		})
	}

	if cfg.TextFinSentimentClassifier {
		tpf.AddProcessor("TextFinSentimentClassifier", func(text string) (string, error) {
			// api call to model
			data := mlclassifier.DataRequest{
				Items: []mlclassifier.DataRequestItem{
					{Text: text},
				},
			}
			client := mlclassifier.NewClient(cfg.TextFinSentimentClassifierURL)
			resp, err := client.Classify(data)

			if err != nil {
				return "", err
			}

			if resp == nil {
				return "", errors.New("response is nil in API call")
			}

			return (*resp)[0].Label, nil
		})
	}

	return tpf
}
