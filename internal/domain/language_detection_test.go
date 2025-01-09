package domain

import "testing"

func TestIsLikelyEnglish(t *testing.T) {
	testCases := []struct {
		name     string
		text     string
		expected bool
	}{
		{
			name:     "English Sentence",
			text:     "This is an English sentence.",
			expected: true,
		},
		{
			name:     "French Sentence",
			text:     "Ceci est une phrase française.",
			expected: false,
		},
		{
			name:     "Long English Sentence",
			text:     "This is a very long english sentence that is very descriptive and contains many words.",
			expected: true,
		},
		/*{
			name:     "Gibberish",
			text:     "asdfghjkl",
			expected: false,
		},*/
		{
			name:     "Empty String",
			text:     "",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isLikelyEnglish(tc.text)
			if result != tc.expected {
				t.Errorf("Expected %t for input '%s', but got %t", tc.expected, tc.text, result)
			}
		})
	}
}
