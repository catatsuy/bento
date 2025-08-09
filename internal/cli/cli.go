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

	"github.com/catatsuy/bento/internal/gemini"
	"github.com/catatsuy/bento/internal/openai"
)

const (
	ExitCodeOK   = 0
	ExitCodeFail = 1

	DefaultExceedThreshold = 4000

	DefaultOpenAIModel = "gpt-5-nano"
	DefaultGeminiModel = "gemini-2.0-flash-lite"
)

var (
	Version string
)

// CLI holds the input/output streams and the Translator.
type CLI struct {
	outStream, errStream io.Writer
	inputStream          io.Reader

	isStdinTerminal bool

	appVersion string

	translator Translator
}

// Translator is the interface used to request a response.
type Translator interface {
	request(ctx context.Context, systemPrompt, prompt, input, model string) (string, error)
}

// NewCLI returns a new CLI instance.
func NewCLI(outStream, errStream io.Writer, inputStream io.Reader, tr Translator, isStdinTerminal bool) *CLI {
	return &CLI{
		appVersion:      version(),
		outStream:       outStream,
		errStream:       errStream,
		inputStream:     inputStream,
		translator:      tr,
		isStdinTerminal: isStdinTerminal,
	}
}

// Run parses CLI arguments and executes the appropriate functionality.
func (c *CLI) Run(args []string) int {
	if len(args) <= 1 {
		fmt.Fprintf(c.errStream, "Error: Insufficient arguments provided\n")
		return ExitCodeFail
	}

	var (
		version bool
		help    bool

		branchSuggestion bool
		commitMessage    bool
		translate        bool
		review           bool

		language     string
		prompt       string
		systemPrompt string
		useModel     string
		targetFile   string
		repoPath     string

		dump        bool
		description string

		isMultiMode  bool
		isSingleMode bool

		limit int

		backend string
	)

	flags := flag.NewFlagSet("bento", flag.ContinueOnError)
	flags.SetOutput(c.errStream)

	flags.BoolVar(&version, "version", false, "Print version information and quit")
	flags.BoolVar(&help, "help", false, "Print help information and quit")
	flags.BoolVar(&help, "h", false, "Print help information and quit")

	flags.StringVar(&targetFile, "file", "", "Specify a target file")

	flags.BoolVar(&branchSuggestion, "branch", false, "Suggest branch name")
	flags.BoolVar(&commitMessage, "commit", false, "Suggest commit message")
	flags.BoolVar(&translate, "translate", false, "Translate text")
	flags.BoolVar(&review, "review", false, "Review source code")

	flags.BoolVar(&dump, "dump", false, "Dump repository contents")
	flags.StringVar(&description, "description", "", "Description of the repository (dump mode)")

	flags.IntVar(&limit, "limit", DefaultExceedThreshold, "Limit the number of characters to translate")

	flags.BoolVar(&isMultiMode, "multi", false, "Multi mode")
	flags.BoolVar(&isSingleMode, "single", false, "Single mode (default)")

	flags.StringVar(&language, "language", "", "Specify the output language")
	flags.StringVar(&prompt, "prompt", "", "Prompt text")
	flags.StringVar(&systemPrompt, "system", "", "System prompt text")
	flags.StringVar(&useModel, "model", DefaultOpenAIModel, "Use models such as gpt-4o-mini, gpt-4-turbo, and gpt-4o. (When using the gemini backend, the default model becomes "+DefaultGeminiModel+")")
	flags.StringVar(&backend, "backend", "openai", "Backend to use: openai or gemini")

	err := flags.Parse(args[1:])
	if err != nil {
		fmt.Fprintf(c.errStream, "Error: %v\n", err)
		return ExitCodeFail
	}

	if version {
		fmt.Fprintf(c.errStream, "bento version %s; %s\n", c.appVersion, runtime.Version())
		return ExitCodeOK
	}

	if help {
		fmt.Fprintf(c.errStream, "bento version %s; %s\n", c.appVersion, runtime.Version())
		fmt.Fprintf(c.errStream, "Usage of bento:\n")
		flags.PrintDefaults()
		return ExitCodeOK
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	if isSingleMode && isMultiMode {
		fmt.Fprintf(c.errStream, "Error: Both 'multi' and 'single' modes cannot be specified simultaneously.\n")
		return ExitCodeFail
	}

	if !isMultiMode && !isSingleMode {
		// Default to single mode if no mode is specified.
		isSingleMode = true
	}

	if (!translate && !review) && language != "" {
		fmt.Fprintf(c.errStream, "Error: The '-language' option can only be used with '-translate' or '-review'.\n")
		return ExitCodeFail
	}

	// If not in dump mode, ensure a translator is set.
	if !dump {
		if c.translator == nil {
			// Choose translator based on the backend flag.
			switch strings.ToLower(backend) {
			case "gemini":
				apiKey := os.Getenv("GEMINI_API_KEY")
				if apiKey == "" {
					fmt.Fprintln(c.errStream, "Error: You need to set GEMINI_API_KEY")
					return ExitCodeFail
				}
				if useModel == DefaultOpenAIModel {
					useModel = DefaultGeminiModel
				}
				gt, err := NewGeminiTranslator(apiKey)
				if err != nil {
					fmt.Fprintf(c.errStream, "Error creating Gemini translator: %v\n", err)
					return ExitCodeFail
				}
				c.translator = gt
			default:
				apiKey := os.Getenv("OPENAI_API_KEY")
				if apiKey == "" {
					fmt.Fprintln(c.errStream, "Error: You need to set OPENAI_API_KEY")
					return ExitCodeFail
				}
				ot, err := NewOpenAITranslator(apiKey)
				if err != nil {
					fmt.Fprintf(c.errStream, "Error creating OpenAI translator: %v\n", err)
					return ExitCodeFail
				}
				c.translator = ot
			}
		}
	}

	if dump {
		if flags.NArg() < 1 {
			repoPath, err = os.Getwd()
			if err != nil {
				fmt.Fprintf(c.errStream, "Error: A repository path must be specified for dump mode.\n")
				return ExitCodeFail
			}
		} else {
			repoPath = flags.Arg(0)
		}

		if err := c.RunDump(repoPath, description); err != nil {
			fmt.Fprintf(c.errStream, "Error: %v\n", err)
			return ExitCodeFail
		}

		return ExitCodeOK
	}

	if c.isStdinTerminal && targetFile == "" {
		fmt.Fprintf(c.errStream, "Error: The '-file' option is required when reading from standard input.\n")
		return ExitCodeFail
	}

	if !c.isStdinTerminal && targetFile != "" {
		fmt.Fprintf(c.errStream, "Error: The '-file' option cannot be used when reading from a file.\n")
		return ExitCodeFail
	}

	if branchSuggestion {
		isSingleMode = true
		isMultiMode = false
		prompt = "Generate a branch name directly from the provided source code differences without any additional text or formatting:\n\n"
	} else if commitMessage {
		isSingleMode = true
		isMultiMode = false
		prompt = "Generate a commit message directly from the provided source code differences without any additional text or formatting within 72 characters:\n\n"
	} else if translate {
		if language == "" {
			language = "en"
		}
		isMultiMode = true
		isSingleMode = false
		prompt = "Translate the following text to " + language + " without any additional text or formatting:\n\n"
	} else if review {
		isSingleMode = true
		isMultiMode = false
		prompt = `Please review the following code as an experienced engineer, focusing only on areas where there are issues. The code is provided as a Git diff, where lines prefixed with + represent additions and lines prefixed with - represent deletions. Analyze the changes accordingly.
Provide feedback only if there is a problem in any of the following aspects: Completeness, Bugs, Security, Code Style, Performance, Readability, Documentation, Testing, Scalability, Dependencies, or Error Handling.
If you find a problem, briefly explain the issue and provide a specific suggestion for improvement. When possible, include a code example that demonstrates how to fix the issue. If there are no issues in a particular area, you do not need to mention it. Avoid numbering the feedback items.`

		if language != "" {
			prompt += " Please provide the feedback in " + language + "."
		}

		prompt += "\n\n"
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
		content, err := io.ReadAll(c.inputStream)
		if err != nil {
			fmt.Fprintf(c.errStream, "Error: %v\n", err)
			return ExitCodeFail
		}

		suggestion, err := c.translator.request(ctx, systemPrompt, prompt, string(content), useModel)
		if err != nil {
			fmt.Fprintf(c.errStream, "Error: %v\n", err)
			return ExitCodeFail
		}
		fmt.Fprintf(c.outStream, "%s\n", suggestion)
		return ExitCodeOK
	}

	if isMultiMode {
		err = c.multiRequest(ctx, systemPrompt, prompt, useModel, limit)
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

// multiRequest processes multi-line input in chunks.
func (c *CLI) multiRequest(ctx context.Context, systemPrompt, prompt, useModel string, limit int) error {
	var b strings.Builder
	reader := bufio.NewReader(c.inputStream)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("error reading input: %w", err)
		}
		b.Write(line)
		if b.Len() > limit {
			translatedText, err := c.translator.request(ctx, systemPrompt, prompt, b.String(), useModel)
			if err != nil {
				return fmt.Errorf("failed to translate text: %w", err)
			}
			fmt.Fprintf(c.outStream, "%s\n", translatedText)
			b.Reset()
		}
	}
	if b.Len() > 0 {
		translatedText, err := c.translator.request(ctx, systemPrompt, prompt, b.String(), useModel)
		if err != nil {
			return fmt.Errorf("failed to translate text: %w", err)
		}
		fmt.Fprintf(c.outStream, "%s\n", translatedText)
		b.Reset()
	}
	return nil
}

// GeminiTranslator implements the Translator interface using the Gemini API client.
type GeminiTranslator struct {
	client *gemini.Client
}

// NewGeminiTranslator creates a new GeminiTranslator with the given API key.
func NewGeminiTranslator(apiKey string) (*GeminiTranslator, error) {
	client, err := gemini.NewClient(gemini.GeminiAPIURL, apiKey)
	if err != nil {
		return nil, fmt.Errorf("NewClient: %w", err)
	}
	return &GeminiTranslator{client: client}, nil
}

// request sends a request to the Gemini API and returns the response text.
// It constructs a Payload using the provided system prompt (if any) and user prompt.
func (gt *GeminiTranslator) request(ctx context.Context, systemPrompt, prompt, input, useModel string) (string, error) {
	if len(input) == 0 {
		return "", fmt.Errorf("no input")
	}
	var data *gemini.Payload
	if systemPrompt != "" {
		data = &gemini.Payload{
			Model: useModel, // e.g., "gemini-2.0-flash"
			Messages: []gemini.Message{
				{Role: "user", Content: prompt + input},
			},
		}
	} else {
		data = &gemini.Payload{
			Model: useModel,
			Messages: []gemini.Message{
				{Role: "user", Content: prompt + input},
			},
		}
	}
	resp, err := gt.client.Chat(ctx, data)
	if err != nil {
		return "", fmt.Errorf("http request: %w", err)
	}
	if len(resp.Choices) > 0 {
		return resp.Choices[0].Message.Content, nil
	}
	return "", fmt.Errorf("no translation found")
}

// NewOpenAITranslator creates a new translator using the OpenAI API.
func NewOpenAITranslator(apiKey string) (Translator, error) {
	client, err := openai.NewClient(openai.OpenAIAPIURL, apiKey)
	if err != nil {
		return nil, fmt.Errorf("NewClient: %w", err)
	}
	return &openaiTranslator{client: client}, nil
}

type openaiTranslator struct {
	client *openai.Client
}

func (ot *openaiTranslator) request(ctx context.Context, systemPrompt, prompt, input, useModel string) (string, error) {
	if len(input) == 0 {
		return "", fmt.Errorf("no input")
	}
	var data *openai.Payload
	if systemPrompt != "" {
		data = &openai.Payload{
			Model: useModel,
			Messages: []openai.Message{
				{Role: "system", Content: systemPrompt},
				{Role: "user", Content: prompt + input},
			},
		}
	} else {
		data = &openai.Payload{
			Model: useModel,
			Messages: []openai.Message{
				{Role: "user", Content: prompt + input},
			},
		}
	}
	resp, err := ot.client.Chat(ctx, data)
	if err != nil {
		return "", fmt.Errorf("http request: %w", err)
	}
	if len(resp.Choices) > 0 {
		return resp.Choices[0].Message.Content, nil
	}
	return "", fmt.Errorf("no translation found")
}
