package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	flag.Usage = usage
	quiet := flag.Bool("quiet", false, "Suppress non-error output")
	jsonOut := flag.Bool("json", false, "Machine-readable output")
	noColor := flag.Bool("no-color", false, "Disable ANSI colours")
	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
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
		flag.Usage()
		os.Exit(1)
	}
}

func usage() {
	out := flag.CommandLine.Output()
	fmt.Fprintf(out, "Usage: %s [flags] <command>\n\n", os.Args[0])
	fmt.Fprintln(out, "Commands:")
	fmt.Fprintln(out, "  generate\tGenerate secrets from a template")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Global flags:")
	prev := flag.CommandLine.Output()
	flag.CommandLine.SetOutput(out)
	flag.CommandLine.PrintDefaults()
	flag.CommandLine.SetOutput(prev)
	fmt.Fprintln(out)
	fmt.Fprintf(out, "Usage of generate:\n  %s generate [flags]\n", os.Args[0])
	fs := flag.NewFlagSet("generate", flag.ContinueOnError)
	fs.SetOutput(out)
	fs.String("f", "env.example", "Template file")
	fs.String("o", "secret.env", "Output file")
	fs.PrintDefaults()
}
