package generate

import (
	"os"
	"testing"

	"keptler/internal/parser"
)

func TestMaterialiseRandom(t *testing.T) {
	tmpl := &parser.Template{Secrets: []parser.SecretTemplate{
		{Name: "A", Rule: "random", Params: map[string]string{"length": "8", "charset": "alnum"}},
	}}
	out, err := os.CreateTemp("", "out-env")
	if err != nil {
		t.Fatal(err)
	}
	os.Remove(out.Name())
	state := out.Name() + ".state.age"

	values, err := Materialise(tmpl, out.Name(), state)
	if err != nil {
		t.Fatal(err)
	}
	if v, ok := values["A"]; !ok || len(v) != 8 {
		t.Fatalf("unexpected value: %v", values)
	}
}
