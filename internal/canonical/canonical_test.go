package canonical

import (
	"bytes"
	"math"
	"strings"
	"testing"

	"github.com/pritishertyagi-ops/ContextOS/internal/assets"
)

func TestCanonicalizeSortsObjectKeys(t *testing.T) {
	value := map[string]any{
		"z": 1,
		"a": "hello",
	}

	got, err := Canonicalize(value)
	if err != nil {
		t.Fatal(err)
	}

	want := []byte(`{"a":"hello","z":1}`)

	if !bytes.Equal(got, want) {
		t.Fatalf("Canonicalize() = %s, want %s", got, want)
	}
}

func TestHashIsDeterministic(t *testing.T) {
	value := map[string]any{
		"name":    "research-agent",
		"version": "1.0.0",
	}

	first, err := Hash(value)
	if err != nil {
		t.Fatal(err)
	}

	second, err := Hash(value)
	if err != nil {
		t.Fatal(err)
	}

	if first != second {
		t.Fatalf("hashes differ: %s != %s", first, second)
	}

	if len(first) != 64 {
		t.Fatalf("expected 64-character SHA-256 hash, got %d", len(first))
	}
}

func TestEquivalentYAMLAndJSONHaveSameHash(t *testing.T) {
	yamlData := []byte(`
apiVersion: contextos/v1
kind: Agent
metadata:
  id: "123e4567-e89b-12d3-a456-426614174000"
  name: research-agent
  version: 1.0.0
spec:
  description: Research assistant
  instructions: Find reliable information.
`)

	jsonData := []byte(`{
  "apiVersion": "contextos/v1",
  "kind": "Agent",
  "metadata": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "name": "research-agent",
    "version": "1.0.0"
  },
  "spec": {
    "description": "Research assistant",
    "instructions": "Find reliable information."
  }
}`)

	yamlValue, err := assets.Parse(yamlData, "yaml")
	if err != nil {
		t.Fatalf("parse YAML: %v", err)
	}

	jsonValue, err := assets.Parse(jsonData, "json")
	if err != nil {
		t.Fatalf("parse JSON: %v", err)
	}

	yamlHash, err := Hash(yamlValue)
	if err != nil {
		t.Fatalf("hash YAML: %v", err)
	}

	jsonHash, err := Hash(jsonValue)
	if err != nil {
		t.Fatalf("hash JSON: %v", err)
	}

	if yamlHash != jsonHash {
		t.Fatalf("YAML and JSON hashes differ:\nYAML: %s\nJSON: %s",
			yamlHash, jsonHash)
	}
}

func TestRejectsInexactInteger(t *testing.T) {
	value := map[string]any{
		"number": int64(9007199254740993),
	}

	_, err := Canonicalize(value)
	if err == nil {
		t.Fatal("expected inexact integer to be rejected")
	}
}

func TestRejectsInvalidRoot(t *testing.T) {
	_, err := Canonicalize("not an object")
	if err == nil {
		t.Fatal("expected invalid root to be rejected")
	}
}

func TestRejectsUnsupportedValue(t *testing.T) {
	value := map[string]any{
		"number": complex(1, 2),
	}

	_, err := Canonicalize(value)
	if err == nil || !strings.Contains(err.Error(), "unsupported value type") {
		t.Fatalf("expected unsupported value error, got %v", err)
	}
}

func TestSafeIntegerBoundaries(t *testing.T) {
	tests := []struct {
		name      string
		value     int64
		wantError bool
	}{
		{"maximum safe integer", 9007199254740991, false},
		{"minimum safe integer", -9007199254740991, false},
		{"above maximum safe integer", 9007199254740992, true},
		{"below minimum safe integer", -9007199254740992, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Canonicalize(map[string]any{"number": tt.value})
			if (err != nil) != tt.wantError {
				t.Fatalf("Canonicalize() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestCanonicalizeUnicodeKeyOrdering(t *testing.T) {
	value := map[string]any{
		"\uE000":     "private",
		"\U0001F600": "emoji",
	}

	got, err := Canonicalize(value)
	if err != nil {
		t.Fatal(err)
	}

	want := []byte("{\"😀\":\"emoji\",\"\":\"private\"}")
	if !bytes.Equal(got, want) {
		t.Fatalf("Canonicalize() = %s, want %s", got, want)
	}
}

func TestCanonicalizeNegativeZero(t *testing.T) {
	value := map[string]any{
		"number": math.Copysign(0, -1),
	}

	got, err := Canonicalize(value)
	if err != nil {
		t.Fatal(err)
	}

	want := []byte(`{"number":0}`)
	if !bytes.Equal(got, want) {
		t.Fatalf("Canonicalize() = %s, want %s", got, want)
	}
}

func TestRejectsInvalidUTF8String(t *testing.T) {
	_, err := Canonicalize(map[string]any{
		"value": string([]byte{0xff}),
	})
	if err == nil {
		t.Fatal("expected invalid UTF-8 string to be rejected")
	}
}

func TestRejectsInvalidUTF8ObjectKey(t *testing.T) {
	_, err := Canonicalize(map[string]any{
		string([]byte{0xff}): "value",
	})
	if err == nil {
		t.Fatal("expected invalid UTF-8 object key to be rejected")
	}
}
