package main

import (
	"flag"
	"fmt"
	"path/filepath"

	"keptler/internal/generate"
	"keptler/internal/parser"
)

func runGenerate(args []string, quiet, jsonOut, noColor bool) error {
	fs := flag.NewFlagSet("generate", flag.ContinueOnError)
	template := fs.String("f", "env.example", "Template file")
	outPath := fs.String("o", "secret.env", "Output file")
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}

	tmpl, err := parser.ParseEnvExample(*template)
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}

	statePath := filepath.Join(filepath.Dir(*outPath), ".keptler.state.age")
	values, err := generate.Materialise(tmpl, *outPath, statePath)
	if err != nil {
		return err
	}

	if !quiet {
		for k := range values {
			fmt.Println("generated", k)
		}
	}

	return nil
}
