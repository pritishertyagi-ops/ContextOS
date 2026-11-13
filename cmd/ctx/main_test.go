package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestVersion(t *testing.T) {
	if version != "0.1.0-dev" {
		t.Fatalf("unexpected version: %s", version)
	}
}

func TestUsage(t *testing.T) {
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	printUsage()

	_ = w.Close()
	os.Stdout = old

	var output bytes.Buffer
	_, _ = output.ReadFrom(r)

	if !strings.Contains(output.String(), "ContextOS CLI") {
		t.Fatal("usage output does not contain CLI name")
	}
}
