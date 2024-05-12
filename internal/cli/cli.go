package cli

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"strings"
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

	translator Translator
}

type Translator interface {
	translateText(ctx context.Context, text string) (string, error)
	suggestBranch(ctx context.Context, input string) (string, error)
}

func NewCLI(outStream, errStream io.Writer, inputStream io.Reader, tr Translator) *CLI {
	return &CLI{appVersion: version(), outStream: outStream, errStream: errStream, inputStream: inputStream, translator: tr}
}

func (c *CLI) Run(args []string) int {
	if len(args) <= 1 {
		return ExitCodeFail
	}

	var (
		version bool
		help    bool

		branchSuggestion bool

		translateFile string
	)

	flags := flag.NewFlagSet("bento", flag.ContinueOnError)
	flags.SetOutput(c.errStream)

	flags.BoolVar(&version, "version", false, "Print version information and quit")
	flags.BoolVar(&help, "help", false, "Print help information and quit")
	flags.BoolVar(&help, "h", false, "Print help information and quit")

	flags.StringVar(&translateFile, "translate", "", "Translate file")

	flags.BoolVar(&branchSuggestion, "branch", false, "Suggest branch name")

	err := flags.Parse(args[1:])
	if err != nil {
		fmt.Fprintf(c.errStream, "Error: %v\n", err)
		return ExitCodeFail
	}

	if version {
		fmt.Fprintf(c.errStream, "bento version %s; %s\n", Version, runtime.Version())
		return ExitCodeOK
	}

	if help {
		fmt.Fprintf(c.errStream, "bento version %s; %s\n", Version, runtime.Version())
		fmt.Fprintf(c.errStream, "Usage of bento:\n")
		flags.PrintDefaults()
		return ExitCodeOK
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	if branchSuggestion {
		b := strings.Builder{}
		scanner := bufio.NewScanner(c.inputStream)
		for scanner.Scan() {
			text := strings.TrimSpace(scanner.Text())
			if len(text) == 0 {
				continue
			}
			b.WriteString(text + "\n")
		}

		suggestion, err := c.translator.suggestBranch(ctx, b.String())
		if err != nil {
			fmt.Fprintf(c.errStream, "Error: %v\n", err)
			return ExitCodeFail
		}

		fmt.Fprintf(c.outStream, "%s\n", suggestion)

		return ExitCodeOK
	}

	if translateFile != "" {
		err := c.translateFile(ctx, translateFile)
		if err != nil {
			fmt.Fprintf(c.errStream, "Error: %v\n", err)
			return ExitCodeFail
		}
		return ExitCodeOK
	}

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

func (c *CLI) translateFile(ctx context.Context, file string) error {
	f, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	var b strings.Builder

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if len(text) == 0 {
			continue
		}
		lastChar := text[len(text)-1]
		if lastChar == '.' {
			text = text + "\n"
		} else if lastChar <= 0x7f {
			text = text + " "
		}
		b.WriteString(text)

		if b.Len() > 1000 {
			translatedText, err := c.translator.translateText(ctx, b.String())
			if err != nil {
				return fmt.Errorf("failed to translate text: %w", err)
			}

			fmt.Fprintf(c.outStream, "%s\n", translatedText)

			b.Reset()
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to scan file: %w", err)
	}

	if b.Len() > 0 {
		translatedText, err := c.translator.translateText(ctx, b.String())
		if err != nil {
			return fmt.Errorf("failed to translate text: %w", err)
		}

		fmt.Fprintf(c.outStream, "%s\n", translatedText)

		b.Reset()
	}

	return nil
}

type translator struct {
	Translator

	client *openai.Client
}

func NewTranslator() (*translator, error) {
	client, err := openai.NewClient(openai.OpenAIAPIURL)
	if err != nil {
		return nil, fmt.Errorf("NewClient: %w", err)
	}
	return &translator{
		client: client,
	}, nil
}

func (tr *translator) translateText(ctx context.Context, input string) (string, error) {
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

	resp, err := tr.client.Chat(ctx, data)
	if err != nil {
		return "", fmt.Errorf("http request: %w", err)
	}

	if len(resp.Choices) > 0 {
		return resp.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("no translation found")
}

func (tr *translator) suggestBranch(ctx context.Context, input string) (string, error) {
	if len(input) == 0 {
		return "", fmt.Errorf("no input")
	}

	prompt := fmt.Sprintf("Generate a branch name directly from the provided source code differences without any additional text or formatting:\n\n" + input)

	data := &openai.Payload{
		Model: "gpt-3.5-turbo",
		Messages: []openai.Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	resp, err := tr.client.Chat(ctx, data)
	if err != nil {
		return "", fmt.Errorf("http request: %w", err)
	}

	if len(resp.Choices) > 0 {
		return resp.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("no translation found")
}
