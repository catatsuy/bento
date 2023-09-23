package main

import (
	"os"

	"github.com/catatsuy/bento/internal/cli"
)

func main() {
	tr, _ := cli.NewTranslator()
	cl := cli.NewCLI(os.Stdout, os.Stderr, os.Stdin, tr)
	os.Exit(cl.Run(os.Args))
}
