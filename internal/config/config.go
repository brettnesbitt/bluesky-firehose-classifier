// config/config.go
package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type AppConfig struct {
	DevMode                    bool
	Host                       string
	ServerPort                 int
	JetstreamURL               string
	MongoURI                   string
	RuleEnglishOnly            bool
	RuleMinLength              bool
	RuleMinLengthValue         int
	RuleContainsURL            bool
	RuleContainsKeywords       bool
	RuleContainsHashtag        bool
	RuleContainsKeywordsValues string
	RuleContainsHashtagValues  string

	TextCategoryClassifier     bool
	TextFinSentimentClassifier bool

	TextCategoryClassifierURL     string
	TextFinSentimentClassifierURL string
}

func (c AppConfig) String() string {
	var sb strings.Builder

	sb.WriteString("App Configuration:\n")
	sb.WriteString(fmt.Sprintf("  Dev Mode: %t\n", c.DevMode))
	sb.WriteString(fmt.Sprintf("  Host: %s\n", c.Host))
	sb.WriteString(fmt.Sprintf("  Server Port: %d\n", c.ServerPort))
	sb.WriteString(fmt.Sprintf("  Jetstream URL: %s\n", c.JetstreamURL))

	sb.WriteString("\n  Rules:\n")
	sb.WriteString(fmt.Sprintf("    English Only: %t\n", c.RuleEnglishOnly))
	sb.WriteString(fmt.Sprintf("    Minimum Length: %t (Value: %d)\n", c.RuleMinLength, c.RuleMinLengthValue))
	sb.WriteString(fmt.Sprintf("    Contains URL: %t\n", c.RuleContainsURL))
	sb.WriteString(fmt.Sprintf("    Contains Keywords: %t (Values: %s)\n", c.RuleContainsKeywords, c.RuleContainsKeywordsValues))
	sb.WriteString(fmt.Sprintf("    Contains Hashtag: %t (Values: %s)\n", c.RuleContainsHashtag, c.RuleContainsHashtagValues))

	sb.WriteString("\n  Text Processing:\n")
	sb.WriteString(fmt.Sprintf("    Category Classifier: %t\n", c.TextCategoryClassifier))
	sb.WriteString(fmt.Sprintf("    Financial Sentiment Classifier: %t\n", c.TextFinSentimentClassifier))

	return sb.String()
}

func LoadConfig() (*AppConfig, error) {

	// Initialize Viper
	viper.SetConfigFile(".env") // Set the path to your .env file
	viper.SetConfigType("env")  // Set the file format (optional, as Viper can infer)

	// Read the configuration file
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	cfg := &AppConfig{
		DevMode:                       viper.GetBool("DEV_MODE"),
		Host:                          viper.GetString("HOST"),
		ServerPort:                    viper.GetInt("SERVER_PORT"),
		JetstreamURL:                  viper.GetString("JETSTREAM_URL"),
		MongoURI:                      viper.GetString("MONGO_URI"),
		RuleEnglishOnly:               viper.GetBool("RULE_ENGLISH_ONLY"),
		RuleMinLength:                 viper.GetBool("RULE_MIN_LENGTH"),
		RuleMinLengthValue:            viper.GetInt("RULE_MIN_LENGTH_VALUE"),
		RuleContainsKeywords:          viper.GetBool("RULE_CONTAINS_KEYWORDS"),
		RuleContainsKeywordsValues:    viper.GetString("RULE_CONTAINS_KEYWORDS_VALUE"),
		RuleContainsHashtag:           viper.GetBool("RULE_CONTAINS_HASHTAG"),
		RuleContainsHashtagValues:     viper.GetString("RULE_CONTAINS_HASHTAG_VALUE"),
		TextCategoryClassifier:        viper.GetBool("TEXT_CATEGORY_CLASSIFIER"),
		TextFinSentimentClassifier:    viper.GetBool("TEXT_FIN_SENTIMENT_CLASSIFIER"),
		TextCategoryClassifierURL:     viper.GetString("TEXT_CATEGORY_CLASSIFIER_URL"),
		TextFinSentimentClassifierURL: viper.GetString("TEXT_FIN_SENTIMENT_CLASSIFIER_URL"),
	}

	return cfg, nil
}
