package textutils

import (
	"strings"
	"unicode/utf8"
)

// SplitText splits input text into sections with a maximum word count.
func SplitText(input string, maxWords int) []string {
	sections := make([]string, 0, 100)
	var currentSection []string
	var currentWordCount int

	paragraphs := strings.Split(input, "\n\n")

	for _, paragraph := range paragraphs {
		words := strings.Fields(paragraph)
		wordCount := len(words)

		if wordCount <= maxWords {
			// If the paragraph can fit into the current section
			if currentWordCount+wordCount <= maxWords {
				if currentWordCount > 0 {
					currentSection = append(currentSection, "\n\n")
				}
				currentSection = append(currentSection, paragraph)
				currentWordCount += wordCount
			} else {
				sections = append(sections, strings.Join(currentSection, " "))
				currentSection = []string{paragraph}
				currentWordCount = wordCount
			}
		} else {
			// Paragraph is too long, split it by sentences using rune-based iteration.
			sentenceStart := 0
			for i := 0; i < len(paragraph); {
				r, size := utf8.DecodeRuneInString(paragraph[i:])
				i += size
				if r == '.' {
					sentenceEnd := i
					sentence := strings.TrimSpace(paragraph[sentenceStart:sentenceEnd])
					sentenceStart = sentenceEnd

					sentenceWords := strings.Fields(sentence)
					sentenceWordCount := len(sentenceWords)

					if currentWordCount+sentenceWordCount <= maxWords || currentWordCount == 0 {
						currentSection = append(currentSection, sentence)
						currentWordCount += sentenceWordCount
					} else {
						sections = append(sections, strings.Join(currentSection, " "))
						currentSection = []string{sentence}
						currentWordCount = sentenceWordCount
					}
				}
			}

			// Add any remaining text as the last section.
			remainingText := strings.TrimSpace(paragraph[sentenceStart:])
			if remainingText != "" {
				remainingWords := strings.Fields(remainingText)
				remainingWordCount := len(remainingWords)
				if currentWordCount+remainingWordCount > maxWords && currentWordCount > 0 {
					sections = append(sections, strings.Join(currentSection, " "))
					currentSection = []string{remainingText}
					currentWordCount = remainingWordCount
				} else {
					currentSection = append(currentSection, remainingText)
					currentWordCount += remainingWordCount
				}
			}
		}
	}

	// Add the last section if it has content.
	if len(currentSection) > 0 {
		sections = append(sections, strings.Join(currentSection, " "))
	}

	return sections
}
