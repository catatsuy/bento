package main

import (
	"os"

	"github.com/catatsuy/bento/internal/cli"
	"golang.org/x/term"
)

func main() {
	tr, _ := cli.NewTranslator()
	cl := cli.NewCLI(os.Stdout, os.Stderr, os.Stdin, tr, term.IsTerminal(int(os.Stdin.Fd())))
	os.Exit(cl.Run(os.Args))
}
