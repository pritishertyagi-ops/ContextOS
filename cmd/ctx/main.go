package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pritishertyagi-ops/ContextOS/internal/assets"
	"github.com/pritishertyagi-ops/ContextOS/internal/canonical"
	"github.com/pritishertyagi-ops/ContextOS/internal/validation"
)

const version = "0.1.0-dev"

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		printUsage()
		return nil
	}

	switch args[0] {
	case "version":
		fmt.Println("ContextOS CLI", version)
		return nil

	case "help", "--help", "-h":
		printUsage()
		return nil

	case "asset":
		return runAsset(args[1:])

	default:
		return fmt.Errorf("unknown command: %s", args[0])
	}
}

func runAsset(args []string) error {
	if len(args) != 2 || args[0] != "validate" {
		return fmt.Errorf("usage: ctx asset validate <file>")
	}

	filePath := args[1]

	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("cannot read file: %w", err)
	}

	format := strings.TrimPrefix(
		strings.ToLower(filepath.Ext(filePath)), ".",
	)

	if format != "yaml" && format != "yml" && format != "json" {
		return fmt.Errorf("unsupported file format: %s", format)
	}

	if format == "yml" {
		format = "yaml"
	}

	value, err := assets.Parse(data, format)
	if err != nil {
		return fmt.Errorf("asset validation failed: %w", err)
	}

	if err := validation.ValidateAgent(value); err != nil {
		return fmt.Errorf("agent validation failed: %w", err)
	}

	hash, err := canonical.Hash(value)
	if err != nil {
		return fmt.Errorf("asset fingerprint failed: %w", err)
	}

	fmt.Println("Asset validation successful")
	fmt.Println("File:", filePath)
	fmt.Println("Format:", format)
	fmt.Println("SHA-256:", hash)

	return nil
}

func printUsage() {
	fmt.Println("ContextOS CLI")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  ctx <command>")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  version           Print version")
	fmt.Println("  help              Show help")
	fmt.Println("  asset validate    Validate and fingerprint an asset")
}
