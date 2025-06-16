package generate

import (
	"bufio"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"math/big"
	"os"
	"sort"
	"strconv"
	"strings"

	"filippo.io/age"
	"filippo.io/age/armor"

	"keptler/internal/parser"
)

type stateFile struct {
	Values map[string]string `json:"values"`
}

// Materialise ensures all secrets exist and writes them to outPath and statePath.
func Materialise(tmpl *parser.Template, outPath, statePath string) (map[string]string, error) {
	state := loadState(statePath)
	outVals := loadEnv(outPath)

	changed := false
	for _, s := range tmpl.Secrets {
		if v, ok := outVals[s.Name]; ok {
			state.Values[s.Name] = v
			continue
		}
		if v, ok := state.Values[s.Name]; ok {
			outVals[s.Name] = v
			continue
		}
		v, err := generateValue(s)
		if err != nil {
			return nil, fmt.Errorf("generate %s: %w", s.Name, err)
		}
		state.Values[s.Name] = v
		outVals[s.Name] = v
		changed = true
	}

	if changed {
		if err := writeEnv(outPath, outVals); err != nil {
			return nil, err
		}
		if err := saveState(statePath, state); err != nil {
			return nil, err
		}
	}

	return outVals, nil
}

func generateValue(s parser.SecretTemplate) (string, error) {
	switch s.Rule {
	case "random":
		length := atoiDefault(s.Params["length"], 32)
		charset := s.Params["charset"]
		if charset == "" {
			charset = "alnum"
		}
		return randomString(length, charset)
	case "rsa-private-key":
		bits := atoiDefault(s.Params["bits"], 2048)
		format := s.Params["format"]
		if format == "" {
			format = "pkcs1"
		}
		return generateRSA(bits, format)
	case "derive":
		src := s.Params["source"]
		if src == "" {
			return "", errors.New("derive requires source")
		}
		return "${" + src + "}", nil
	default:
		return "", fmt.Errorf("unsupported rule %q", s.Rule)
	}
}

func randomString(length int, charset string) (string, error) {
	switch charset {
	case "alnum":
		const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
		return randChars(length, letters)
	case "hex":
		b := make([]byte, length)
		if _, err := rand.Read(b); err != nil {
			return "", err
		}
		return hex.EncodeToString(b)[:length], nil
	case "base64":
		b := make([]byte, length)
		if _, err := rand.Read(b); err != nil {
			return "", err
		}
		return base64.StdEncoding.EncodeToString(b)[:length], nil
	case "urlsafe":
		b := make([]byte, length)
		if _, err := rand.Read(b); err != nil {
			return "", err
		}
		return base64.RawURLEncoding.EncodeToString(b)[:length], nil
	default:
		return "", fmt.Errorf("unknown charset %q", charset)
	}
}

func randChars(length int, letters string) (string, error) {
	var sb strings.Builder
	max := big.NewInt(int64(len(letters)))
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		sb.WriteByte(letters[n.Int64()])
	}
	return sb.String(), nil
}

func generateRSA(bits int, format string) (string, error) {
	key, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return "", err
	}
	if format == "pkcs8" {
		b, err := x509.MarshalPKCS8PrivateKey(key)
		if err != nil {
			return "", err
		}
		return string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: b})), nil
	}
	b := x509.MarshalPKCS1PrivateKey(key)
	return string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: b})), nil
}

func atoiDefault(s string, def int) int {
	if v, err := strconv.Atoi(s); err == nil {
		return v
	}
	return def
}

func statePassphrase() string {
	if p := os.Getenv("KEPTLER_STATE_PASSPHRASE"); p != "" {
		return p
	}
	return "localdev"
}

func loadState(path string) stateFile {
	var s stateFile
	s.Values = make(map[string]string)
	f, err := os.Open(path)
	if err != nil {
		return s
	}
	defer f.Close()
	fi, err := f.Stat()
	if err == nil && fi.Mode().Perm()&0o077 != 0 {
		return s
	}
	ar := armor.NewReader(f)
	id, err := age.NewScryptIdentity(statePassphrase())
	if err != nil {
		return s
	}
	r, err := age.Decrypt(ar, id)
	if err != nil {
		return s
	}
	json.NewDecoder(r).Decode(&s)
	return s
}

func saveState(path string, s stateFile) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()
	recipient, err := age.NewScryptRecipient(statePassphrase())
	if err != nil {
		return err
	}
	aw := armor.NewWriter(f)
	w, err := age.Encrypt(aw, recipient)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(&s); err != nil {
		w.Close()
		aw.Close()
		return err
	}
	if err := w.Close(); err != nil {
		aw.Close()
		return err
	}
	return aw.Close()
}

func loadEnv(path string) map[string]string {
	m := make(map[string]string)
	f, err := os.Open(path)
	if err != nil {
		return m
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if idx := strings.Index(line, "="); idx != -1 {
			k := strings.TrimSpace(line[:idx])
			v := strings.TrimSpace(line[idx+1:])
			m[k] = v
		}
	}
	return m
}

func writeEnv(path string, values map[string]string) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if _, err := io.WriteString(f, fmt.Sprintf("%s=%s\n", k, values[k])); err != nil {
			return err
		}
	}
	return nil
}
