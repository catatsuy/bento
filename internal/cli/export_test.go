package cli

import "context"

type MockTranslator struct {
	TranslateTextFunc func(ctx context.Context, prompt, text, model string) (string, error)
}

func (m *MockTranslator) request(ctx context.Context, prompt, text, model string) (string, error) {
	return m.TranslateTextFunc(ctx, prompt, text, model)
}

func (c *CLI) TranslateFile(ctx context.Context, targetFile, prompt, useModel string, limit int) error {
	return c.translateFile(ctx, targetFile, prompt, useModel, limit)
}
