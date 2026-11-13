package assets

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"gopkg.in/yaml.v3"
)

// Parse accepts a JSON or YAML document and returns its decoded value.
func Parse(data []byte, format string) (any, error) {
	switch strings.ToLower(format) {
	case "json":
		return parseJSON(data)
	case "yaml", "yml":
		return parseYAML(data)
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

func parseJSON(data []byte) (any, error) {
	dec := json.NewDecoder(strings.NewReader(string(data)))
	dec.UseNumber()

	value, err := parseJSONValue(dec)
	if err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	_, err = dec.Token()
	if err != io.EOF {
		if err == nil {
			return nil, fmt.Errorf("invalid JSON: multiple values")
		}
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	return value, nil
}

func parseJSONValue(dec *json.Decoder) (any, error) {
	token, err := dec.Token()
	if err != nil {
		return nil, err
	}

	switch value := token.(type) {
	case json.Delim:
		switch value {
		case '{':
			object := make(map[string]any)

			for dec.More() {
				keyToken, err := dec.Token()
				if err != nil {
					return nil, err
				}

				key, ok := keyToken.(string)
				if !ok {
					return nil, fmt.Errorf("object key must be a string")
				}

				if _, exists := object[key]; exists {
					return nil, fmt.Errorf("duplicate JSON key: %q", key)
				}

				item, err := parseJSONValue(dec)
				if err != nil {
					return nil, err
				}
				object[key] = item
			}

			end, err := dec.Token()
			if err != nil || end != json.Delim('}') {
				return nil, fmt.Errorf("unterminated JSON object")
			}
			return object, nil

		case '[':
			var array []any

			for dec.More() {
				item, err := parseJSONValue(dec)
				if err != nil {
					return nil, err
				}
				array = append(array, item)
			}

			end, err := dec.Token()
			if err != nil || end != json.Delim(']') {
				return nil, fmt.Errorf("unterminated JSON array")
			}
			return array, nil
		}

	case string, json.Number, bool, nil:
		return value, nil
	}

	return nil, fmt.Errorf("unexpected JSON token")
}

func parseYAML(data []byte) (any, error) {
	dec := yaml.NewDecoder(strings.NewReader(string(data)))

	var document yaml.Node
	if err := dec.Decode(&document); err != nil {
		return nil, fmt.Errorf("invalid YAML: %w", err)
	}

	if len(document.Content) == 0 {
		return nil, fmt.Errorf("empty YAML document")
	}

	if err := validateYAMLNode(&document); err != nil {
		return nil, err
	}

	var extra yaml.Node
	err := dec.Decode(&extra)
	if err != io.EOF {
		if err == nil {
			return nil, fmt.Errorf("multiple YAML documents are not supported")
		}
		return nil, fmt.Errorf("invalid YAML: %w", err)
	}

	var result any
	if err := document.Decode(&result); err != nil {
		return nil, fmt.Errorf("invalid YAML value: %w", err)
	}

	return result, nil
}

func validateYAMLNode(node *yaml.Node) error {
	if node.Kind == yaml.AliasNode {
		return fmt.Errorf("YAML aliases are not supported")
	}

	switch node.Kind {
	case yaml.DocumentNode:
		for _, child := range node.Content {
			if err := validateYAMLNode(child); err != nil {
				return err
			}
		}

	case yaml.MappingNode:
		if node.Tag != "!!map" {
			return fmt.Errorf("unsupported YAML mapping tag: %s", node.Tag)
		}

		seen := make(map[string]bool)

		for i := 0; i < len(node.Content); i += 2 {
			key := node.Content[i]

			if key.Kind != yaml.ScalarNode || key.Tag != "!!str" {
				return fmt.Errorf("YAML mapping keys must be strings")
			}

			if seen[key.Value] {
				return fmt.Errorf("duplicate YAML key: %q", key.Value)
			}
			seen[key.Value] = true
		}

		for _, child := range node.Content {
			if err := validateYAMLNode(child); err != nil {
				return err
			}
		}

	case yaml.SequenceNode:
		if node.Tag != "!!seq" {
			return fmt.Errorf("unsupported YAML sequence tag: %s", node.Tag)
		}

		for _, child := range node.Content {
			if err := validateYAMLNode(child); err != nil {
				return err
			}
		}

	case yaml.ScalarNode:
		switch node.Tag {
		case "!!str", "!!null", "!!bool", "!!int", "!!float":
			return nil
		default:
			return fmt.Errorf("unsupported YAML tag: %s", node.Tag)
		}

	default:
		return fmt.Errorf("unsupported YAML node kind: %d", node.Kind)
	}

	return nil
}
