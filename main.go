package main

import (
	"os"

	"github.com/catatsuy/bento/internal/cli"
	"golang.org/x/term"
)

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		cl := cli.NewCLI(os.Stdout, os.Stderr, os.Stdin, nil, term.IsTerminal(int(os.Stdin.Fd())))
		os.Exit(cl.Run(os.Args))
	}
	tr, err := cli.NewTranslator(apiKey)
	if err != nil {
		os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}
	cl := cli.NewCLI(os.Stdout, os.Stderr, os.Stdin, tr, term.IsTerminal(int(os.Stdin.Fd())))
	os.Exit(cl.Run(os.Args))
}
