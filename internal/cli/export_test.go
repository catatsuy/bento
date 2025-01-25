package cli

import "context"

type MockTranslator struct {
	TranslateTextFunc func(ctx context.Context, systemPrompt, prompt, text, model string) (string, error)
}

func (m *MockTranslator) request(ctx context.Context, systemPrompt, prompt, text, model string) (string, error) {
	return m.TranslateTextFunc(ctx, systemPrompt, prompt, text, model)
}

func (c *CLI) MultiRequest(ctx context.Context, systemPrompt, prompt, useModel string, limit int) error {
	return c.multiRequest(ctx, systemPrompt, prompt, useModel, limit)
}
