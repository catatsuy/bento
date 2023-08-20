package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/catatsuy/bento/openai"
)

func TranslateText() (string, error) {
	prompt := fmt.Sprintf("英語を日本語に翻訳してください。返事は翻訳された文章のみにしてください。Hello. I am a student.")

	data := &openai.Payload{
		Model: "gpt-3.5-turbo",
		Messages: []openai.Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	client, err := openai.NewClient(openai.OpenAIAPIURL)
	if err != nil {
		return "", err
	}

	resp, err := client.Chat(context.Background(), data)
	if err != nil {
		return "", err
	}

	if len(resp.Choices) > 0 {
		return strings.TrimSpace(resp.Choices[0].Message.Content), nil
	}

	return "", fmt.Errorf("no translation found")
}

func main() {
	translatedText, err := TranslateText()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println(translatedText)
}
