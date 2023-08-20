package main

import (
	"os"

	"github.com/catatsuy/bento/cli"
)

func main() {
	cl := cli.NewCLI(os.Stdout, os.Stderr, os.Stdin)
	os.Exit(cl.Run(os.Args))
}
