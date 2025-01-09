package domain

import (
	"testing"

	"github.com/pemistahl/lingua-go"
)

func BenchmarkIsLikelyEnglishShort(b *testing.B) {
	text := "This is a short English sentence."
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		isLikelyEnglish(text)
	}
}

func BenchmarkIsLikelyEnglishLong(b *testing.B) {
	text := "This is a very long English sentence that is very descriptive and contains many words and more words and even more words to make it even longer for the benchmark test."
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		isLikelyEnglish(text)
	}
}

func BenchmarkIsLikelyEnglishGibberish(b *testing.B) {
	text := "asdfghjklqwertyuiopzxcvbnm"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		isLikelyEnglish(text)
	}
}

func BenchmarkDetectorCreation(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		languages := []lingua.Language{
			lingua.English,
			lingua.French,
		}
		lingua.NewLanguageDetectorBuilder().
			FromLanguages(languages...).
			Build()
	}
}
