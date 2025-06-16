package parser

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

// SecretTemplate represents a single secret specification parsed from .env.example
// comments.
type SecretTemplate struct {
	Name   string
	Rule   string
	Params map[string]string
}

// Template holds all parsed secret templates.
type Template struct {
	Secrets []SecretTemplate
}

// ParseEnvExample reads the given file and extracts keptler annotations.
func ParseEnvExample(path string) (*Template, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var tmpl Template
	scanner := bufio.NewScanner(f)
	var pending map[string]string
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "# keptler:") {
			params, err := parseAnnotation(strings.TrimPrefix(trimmed, "# keptler:"))
			if err != nil {
				return nil, err
			}
			pending = params
			continue
		}

		if idx := strings.Index(line, "="); idx != -1 {
			name := strings.TrimSpace(line[:idx])
			rest := line[idx+1:]
			ann := pending
			pending = nil
			if i := strings.Index(rest, "# keptler:"); i != -1 {
				params, err := parseAnnotation(rest[i+len("# keptler:"):])
				if err != nil {
					return nil, err
				}
				ann = params
			}
			if ann != nil {
				tmpl.Secrets = append(tmpl.Secrets, SecretTemplate{Name: name, Rule: ann["rule"], Params: ann})
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return &tmpl, nil
}

func parseAnnotation(text string) (map[string]string, error) {
	parts := strings.Fields(strings.TrimSpace(text))
	if len(parts) == 0 {
		return nil, errors.New("empty annotation")
	}
	params := make(map[string]string)
	params["rule"] = parts[0]
	for _, p := range parts[1:] {
		kv := strings.SplitN(p, "=", 2)
		if len(kv) != 2 {
			return nil, fmt.Errorf("invalid param %q", p)
		}
		params[kv[0]] = kv[1]
	}
	return params, nil
}
