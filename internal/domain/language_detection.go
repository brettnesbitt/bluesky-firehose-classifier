package domain

import "github.com/pemistahl/lingua-go"

func isLikelyEnglish(text string) bool {
	languages := []lingua.Language{
		lingua.English,
		lingua.French,
	}
	detector := lingua.NewLanguageDetectorBuilder().
		FromLanguages(languages...).
		Build()

	detectedLanguage, exists := detector.DetectLanguageOf(text)
	return exists && detectedLanguage == lingua.English
}
