package cli

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os/signal"
	"runtime"
	"runtime/debug"
	"syscall"

	"github.com/catatsuy/bento/internal/openai"
)

const (
	ExitCodeOK   = 0
	ExitCodeFail = 1
)

var (
	Version string
)

type CLI struct {
	outStream, errStream io.Writer
	inputStream          io.Reader

	appVersion string
}

func NewCLI(outStream, errStream io.Writer, inputStream io.Reader) *CLI {
	return &CLI{appVersion: version(), outStream: outStream, errStream: errStream, inputStream: inputStream}
}

func (c *CLI) Run(args []string) int {
	if len(args) <= 1 {
		return ExitCodeFail
	}

	var (
		version bool
	)

	flags := flag.NewFlagSet("bento", flag.ContinueOnError)
	flags.SetOutput(c.errStream)

	flags.BoolVar(&version, "version", false, "Print version information and quit")

	err := flags.Parse(args[1:])
	if err != nil {
		fmt.Fprintf(c.errStream, "Error: %v\n", err)
		return ExitCodeFail
	}

	if version {
		fmt.Fprintf(c.errStream, "bento version %s; %s\n", Version, runtime.Version())
		return ExitCodeOK
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	translatedText, err := translateText(ctx, args[1])
	if err != nil {
		fmt.Fprintf(c.errStream, "Error: %v\n", err)
		return ExitCodeFail
	}

	fmt.Fprintf(c.outStream, "%s\n", translatedText)

	return ExitCodeOK
}

func version() string {
	if Version != "" {
		return Version
	}

	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "(devel)"
	}
	return info.Main.Version
}

func translateText(ctx context.Context, input string) (string, error) {
	if len(input) == 0 {
		return "", fmt.Errorf("no input")
	}

	prompt := fmt.Sprintf("英語を日本語に翻訳してください。返事は翻訳された文章のみにしてください。" + input)

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
		return "", fmt.Errorf("NewClient: %w", err)
	}

	resp, err := client.Chat(ctx, data)
	if err != nil {
		return "", fmt.Errorf("http request: %w", err)
	}

	if len(resp.Choices) > 0 {
		return resp.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("no translation found")
}
