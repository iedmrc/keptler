package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	quiet := flag.Bool("quiet", false, "Suppress non-error output")
	jsonOut := flag.Bool("json", false, "Machine-readable output")
	noColor := flag.Bool("no-color", false, "Disable ANSI colours")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "expected command")
		os.Exit(1)
	}

	cmd := flag.Arg(0)
	args := flag.Args()[1:]

	switch cmd {
	case "generate":
		if err := runGenerate(args, *quiet, *jsonOut, *noColor); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	default:
		fmt.Fprintln(os.Stderr, "unknown command:", cmd)
		os.Exit(1)
	}
}
