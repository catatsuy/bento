package textutils

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestSplitText(t *testing.T) {
	tests := []struct {
		input    string
		maxWords int
		expected []string
	}{
		{
			input:    "This is an English paragraph with a few words. Finally, This is an English paragraph with a few words.",
			maxWords: 10,
			expected: []string{
				"This is an English paragraph with a few words.",
				"Finally, This is an English paragraph with a few words.",
			},
		},
		{
			input:    strings.Repeat("Word ", 1001) + "LastWord",
			maxWords: 1000,
			expected: []string{
				strings.Repeat("Word ", 1001) + "LastWord",
			},
		},
		{
			input:    "This is a paragraph. This is another paragraph.",
			maxWords: 5,
			expected: []string{
				"This is a paragraph.",
				"This is another paragraph.",
			},
		},
		{
			input:    "This paragraph is very very long. It has many sentences. Each sentence should be counted separately. This ensures that splitting is accurate.",
			maxWords: 5,
			expected: []string{
				"This paragraph is very very long.",
				"It has many sentences.",
				"Each sentence should be counted separately.",
				"This ensures that splitting is accurate.",
			},
		},
	}

	for _, test := range tests {
		result := SplitText(test.input, test.maxWords)
		if diff := cmp.Diff(test.expected, result); diff != "" {
			t.Errorf("SplitText(%q, %d) mismatch (-want +got):\n%s", test.input, test.maxWords, diff)
		}
	}
}
