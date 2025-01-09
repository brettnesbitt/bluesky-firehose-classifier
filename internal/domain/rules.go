package domain

import (
	"fmt"
	"regexp"
	"strings"

	"stockseer.ai/blueksy-firehose/internal/config"
)

// Rule represents a single rule with a description and a function to evaluate it.
type Rule struct {
	Description string
	Evaluate    func(text string) bool
}

// RuleFactory helps create and manage rules.
type RuleFactory struct {
	rules []Rule
}

// NewRuleFactory creates a new RuleFactory.
func NewRuleFactory() *RuleFactory {
	return &RuleFactory{}
}

// AddRule adds a new rule to the factory.
func (rf *RuleFactory) AddRule(description string, evaluate func(text string) bool) {
	rf.rules = append(rf.rules, Rule{Description: description, Evaluate: evaluate})
}

// EvaluateAll evaluates all rules against the given text and returns bool indicating if all passed
// and a map of rule descriptions to results.
func (rf *RuleFactory) EvaluateAll(text string) (bool, map[string]bool) {
	results := make(map[string]bool)
	allPass := true
	for _, rule := range rf.rules {
		valid := rule.Evaluate(text)
		results[rule.Description] = rule.Evaluate(text)
		if !valid {
			allPass = false
		}
	}
	return allPass, results
}

/*
* 	Our application has a set of rules for basic filtering of data that we consume
* 	Not all posts are valuable for our models or analysis and we can minimize the amount of
* 	data that lands in our downstream ML models.
 */
func InitRules(cfg *config.AppConfig) *RuleFactory {

	rf := NewRuleFactory()

	if cfg.RuleEnglishOnly {
		rf.AddRule("English posts only", func(text string) bool {
			return isLikelyEnglish(text)
		})
	}

	// string meets minimum length
	if cfg.RuleMinLength {
		minLength := cfg.RuleMinLengthValue
		description := fmt.Sprintf("Length greater than %d characters", minLength)
		rf.AddRule(description, func(text string) bool {
			return len(text) > minLength
		})
	}

	// string with a url
	if cfg.RuleContainsURL {
		urlRegex := regexp.MustCompile(`(https?://)?([\w\.]+)\.([a-z]{2,6}\.?)?/?.*`)
		rf.AddRule("Contains a URL", func(text string) bool {
			return urlRegex.MatchString(text)
		})
	}

	// string contains specific keywords
	if cfg.RuleContainsKeywords {
		keywords := strings.Split(cfg.RuleContainsKeywordsValues, ",")
		rf.AddRule("Contains relevant keywords", func(text string) bool {
			textLower := strings.ToLower(text)
			for _, keyword := range keywords {
				if strings.Contains(textLower, keyword) {
					return true
				}
			}
			return false
		})
	}

	// string contains specific hashtags
	if cfg.RuleContainsHashtag {
		hashtags := strings.Split(cfg.RuleContainsHashtagValues, ",")
		rf.AddRule("Contains relevant hashtags", func(text string) bool {
			textLower := strings.ToLower(text)
			for _, hashtag := range hashtags {
				if strings.Contains(textLower, fmt.Sprintf("#%s", hashtag)) {
					return true
				}
			}
			return false
		})
	}

	return rf
}
