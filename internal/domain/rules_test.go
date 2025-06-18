package domain

import (
	"fmt"
	"strings"
	"testing"

	"stockseer.ai/blueksy-firehose/internal/config"
	"stockseer.ai/blueksy-firehose/internal/models"
)

func TestRuleFactory(t *testing.T) {
	mockCfg := &config.AppConfig{
		RuleMinLength:              true,
		RuleMinLengthValue:         10,
		RuleContainsURL:            true,
		RuleContainsKeywords:       true,
		RuleContainsHashtag:        true,
		RuleContainsKeywordsValues: "tutorial,insight,solution,guide",
		RuleContainsHashtagValues:  "golang",
	}

	rf := InitRules(mockCfg)

	testCases := []struct {
		name            string
		text            string
		expectedResults map[string]bool
	}{
		{
			name: "Post with keywords, URL, hashtag, and long length",
			text: "This is a great tutorial on how to use Go! https://example.com #golang",
			expectedResults: map[string]bool{
				"Contains relevant keywords":        true,
				"Length greater than 10 characters": true,
				"Contains a URL":                    true,
				"Contains relevant hashtags":        true,
			},
		},
		{
			name: "Short post without keywords, URL, or hashtag",
			text: "short",
			expectedResults: map[string]bool{
				"Contains relevant keywords":        false,
				"Length greater than 10 characters": false,
				"Contains a URL":                    false,
				"Contains relevant hashtags":        false,
			},
		},
		{
			name: "Post with keyword but short length",
			text: "tutorial",
			expectedResults: map[string]bool{
				"Contains relevant keywords":        true,
				"Length greater than 10 characters": false,
				"Contains a URL":                    false,
				"Contains relevant hashtags":        false,
			},
		},
		{
			name: "Post with URL but no hashtag",
			text: "Check this out: http://test.com",
			expectedResults: map[string]bool{
				"Contains relevant keywords":        false,
				"Length greater than 10 characters": true,
				"Contains a URL":                    true,
				"Contains relevant hashtags":        false,
			},
		},
		{
			name: "Empty string",
			text: "",
			expectedResults: map[string]bool{
				"Contains relevant keywords":        false,
				"Length greater than 10 characters": false,
				"Contains a URL":                    false,
				"Contains relevant hashtags":        false,
			},
		},
		{
			name: "Post with mixed case keyword",
			text: "TuToRiAl",
			expectedResults: map[string]bool{
				"Contains relevant keywords":        true,
				"Length greater than 10 characters": false,
				"Contains a URL":                    false,
				"Contains relevant hashtags":        false,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a mock message
			mockMessage := &models.ProtoMessage{
				Commit: &models.Commit{
					Record: &models.Record{
						Text: tc.text,
					},
				},
			}
			_, results := rf.EvaluateAll(tc.text, mockMessage)
			if len(results) != len(tc.expectedResults) {
				t.Errorf("Expected %d results, got %d", len(tc.expectedResults), len(results))
			}
			for ruleName, expected := range tc.expectedResults {
				actual, ok := results[ruleName]
				if !ok {
					t.Errorf("Rule '%s' not found in results", ruleName)
					continue
				}
				if actual != expected {
					t.Errorf(
						"Rule '%s': expected %t, got %t for text: %s",
						ruleName,
						expected,
						actual,
						tc.text,
					)
				}
			}
		})
	}
}

func ExampleRuleFactory() {
	rf := NewRuleFactory()

	// Create a mock message
	mockMessage := &models.ProtoMessage{
		Commit: &models.Commit{
			Record: &models.Record{
				Text: "",
			},
		},
	}

	rf.AddRule("Contains 'example'", func(text string, message *models.ProtoMessage) bool {
		return strings.Contains(text, "example")
	})

	_, results := rf.EvaluateAll("This is an example text.", mockMessage)
	fmt.Println(results["Contains 'example'"])
	// Output: true
}
