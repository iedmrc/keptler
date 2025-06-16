package parser

import (
	"os"
	"testing"
)

func TestParseEnvExample(t *testing.T) {
	content := `# keptler: random length=10 charset=hex
SECRET_ONE=
SECRET_TWO= # keptler: rsa-private-key bits=1024
`
	f, err := os.CreateTemp("", "env-example")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
	f.Close()

	tmpl, err := ParseEnvExample(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	if len(tmpl.Secrets) != 2 {
		t.Fatalf("expected 2 secrets, got %d", len(tmpl.Secrets))
	}
	if tmpl.Secrets[0].Name != "SECRET_ONE" || tmpl.Secrets[0].Rule != "random" {
		t.Fatalf("unexpected first secret: %+v", tmpl.Secrets[0])
	}
	if tmpl.Secrets[1].Rule != "rsa-private-key" {
		t.Fatalf("unexpected second secret rule: %s", tmpl.Secrets[1].Rule)
	}
}
