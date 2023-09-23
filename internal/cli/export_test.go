package cli

import "context"

type MockTranslator struct {
	TranslateTextFunc func(ctx context.Context, text string) (string, error)
}

func (m *MockTranslator) translateText(ctx context.Context, text string) (string, error) {
	return m.TranslateTextFunc(ctx, text)
}

func (c *CLI) TranslateFile(ctx context.Context, file string) error {
	return c.translateFile(ctx, file)
}
