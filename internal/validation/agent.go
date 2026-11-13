package validation

import (
	"fmt"
	"regexp"
)

var (
	uuidPattern = regexp.MustCompile(
		`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`,
	)
	namePattern = regexp.MustCompile(
		`^[a-z0-9][a-z0-9-]*$`,
	)
	versionPattern = regexp.MustCompile(
		`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$`,
	)
)

func object(value any, path string) (map[string]any, error) {
	result, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%s must be an object", path)
	}
	return result, nil
}

func checkFields(value map[string]any, path string, allowed ...string) error {
	for key := range value {
		found := false
		for _, field := range allowed {
			if key == field {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("%s contains unknown field %q", path, key)
		}
	}
	return nil
}

// ValidateAgent checks the provider-neutral ContextOS Agent structure.
func ValidateAgent(value any) error {
	root, err := object(value, "agent")
	if err != nil {
		return err
	}

	if err := checkFields(root, "agent",
		"apiVersion", "kind", "metadata", "spec", "extensions"); err != nil {
		return err
	}

	if root["apiVersion"] != "contextos/v1" {
		return fmt.Errorf("apiVersion must be contextos/v1")
	}

	if root["kind"] != "Agent" {
		return fmt.Errorf("kind must be Agent")
	}

	metadata, err := object(root["metadata"], "metadata")
	if err != nil {
		return err
	}

	if err := checkFields(metadata, "metadata", "id", "name", "version"); err != nil {
		return err
	}

	id, ok := metadata["id"].(string)
	if !ok || !uuidPattern.MatchString(id) {
		return fmt.Errorf("metadata.id must be a valid UUID")
	}

	name, ok := metadata["name"].(string)
	if !ok || !namePattern.MatchString(name) {
		return fmt.Errorf("metadata.name must be a valid lowercase name")
	}

	version, ok := metadata["version"].(string)
	if !ok || !versionPattern.MatchString(version) {
		return fmt.Errorf("metadata.version must use MAJOR.MINOR.PATCH format")
	}

	spec, err := object(root["spec"], "spec")
	if err != nil {
		return err
	}

	if err := checkFields(spec, "spec", "description", "instructions"); err != nil {
		return err
	}

	instructions, ok := spec["instructions"].(string)
	if !ok || len(instructions) == 0 {
		return fmt.Errorf("spec.instructions must be a non-empty string")
	}

	if description, exists := spec["description"]; exists {
		if _, ok := description.(string); !ok {
			return fmt.Errorf("spec.description must be a string")
		}
	}

	if extensions, exists := root["extensions"]; exists {
		extensionMap, err := object(extensions, "extensions")
		if err != nil {
			return err
		}
		if len(extensionMap) != 0 {
			return fmt.Errorf("extensions are not registered yet")
		}
	}

	return nil
}
