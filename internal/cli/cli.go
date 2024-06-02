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

	DefaultExceedThreshold = 4000
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
	request(ctx context.Context, prompt, input, model string) (string, error)
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
		commitMessage    bool
		translate        bool

		language   string
		prompt     string
		useModel   string
		targetFile string

		isMultiMode  bool
		isSingleMode bool

		limit int
	)

	flags := flag.NewFlagSet("bento", flag.ContinueOnError)
	flags.SetOutput(c.errStream)

	flags.BoolVar(&version, "version", false, "Print version information and quit")
	flags.BoolVar(&help, "help", false, "Print help information and quit")
	flags.BoolVar(&help, "h", false, "Print help information and quit")

	flags.StringVar(&targetFile, "file", "", "specify a target file")

	flags.BoolVar(&branchSuggestion, "branch", false, "Suggest branch name")
	flags.BoolVar(&commitMessage, "commit", false, "Suggest commit message")
	flags.BoolVar(&translate, "translate", false, "Translate text")

	flags.IntVar(&limit, "limit", DefaultExceedThreshold, "Limit the number of characters to translate")

	flags.BoolVar(&isMultiMode, "multi", false, "Multi mode")
	flags.BoolVar(&isSingleMode, "single", false, "Single mode")

	flags.StringVar(&language, "language", "en", "Translate to language (default: en)")
	flags.StringVar(&prompt, "prompt", "", "Prompt text")
	flags.StringVar(&useModel, "model", "gpt-3.5-turbo", "Use model (gpt-3.5-turbo, gpt-4-turbo and gpt-4o etc (default: gpt-3.5-turbo))")

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
		isSingleMode = true
		isMultiMode = false
		prompt = "Generate a branch name directly from the provided source code differences without any additional text or formatting:\n\n"
	} else if commitMessage {
		isSingleMode = true
		isMultiMode = false
		prompt = "Generate a commit message directly from the provided source code differences without any additional text or formatting within 72 characters:\n\n"
	} else if translate {
		isMultiMode = true
		isSingleMode = false
		prompt = "Translate the following text to " + language + " without any additional text or formatting:\n\n"
	}

	if !isSingleMode && !isMultiMode {
		fmt.Fprintf(c.errStream, "Error: no mode specified\n")
		return ExitCodeFail
	}

	if targetFile != "" {
		f, err := os.Open(targetFile)
		if err != nil {
			fmt.Fprintf(c.errStream, "Error: %v\n", err)
			return ExitCodeFail
		}
		defer f.Close()
		c.inputStream = f
	}

	if isSingleMode {
		by, err := io.ReadAll(c.inputStream)
		if err != nil {
			fmt.Fprintf(c.errStream, "Error: %v\n", err)
			return ExitCodeFail
		}

		suggestion, err := c.translator.request(ctx, prompt, string(by), useModel)
		if err != nil {
			fmt.Fprintf(c.errStream, "Error: %v\n", err)
			return ExitCodeFail
		}

		fmt.Fprintf(c.outStream, "%s\n", suggestion)

		return ExitCodeOK
	}

	if isMultiMode {
		err = c.multiRequest(ctx, prompt, useModel, limit)
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

func (c *CLI) multiRequest(ctx context.Context, prompt, useModel string, limit int) error {
	var b strings.Builder

	scanner := bufio.NewScanner(c.inputStream)
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if len(text) == 0 {
			continue
		}
		b.WriteString(text + "\n")

		if b.Len() > limit {
			translatedText, err := c.translator.request(ctx, prompt, b.String(), useModel)
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
		translatedText, err := c.translator.request(ctx, prompt, b.String(), useModel)
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

func (tr *translator) request(ctx context.Context, prompt, input, useModel string) (string, error) {
	if len(input) == 0 {
		return "", fmt.Errorf("no input")
	}

	data := &openai.Payload{
		Model: useModel,
		Messages: []openai.Message{
			{
				Role:    "user",
				Content: prompt + input,
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
