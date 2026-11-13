package validation

import "testing"

func validAgent() map[string]any {
	return map[string]any{
		"apiVersion": "contextos/v1",
		"kind":       "Agent",
		"metadata": map[string]any{
			"id":      "00000000-0000-4000-8000-000000000001",
			"name":    "research-agent",
			"version": "1.0.0",
		},
		"spec": map[string]any{
			"description":  "Research assistant",
			"instructions": "Research the requested topic.",
		},
		"extensions": map[string]any{},
	}
}

func TestValidateAgent(t *testing.T) {
	if err := ValidateAgent(validAgent()); err != nil {
		t.Fatalf("valid agent rejected: %v", err)
	}
}

func TestRejectInvalidUUID(t *testing.T) {
	agent := validAgent()
	agent["metadata"].(map[string]any)["id"] = "invalid"

	if err := ValidateAgent(agent); err == nil {
		t.Fatal("invalid UUID accepted")
	}
}

func TestRejectUnknownField(t *testing.T) {
	agent := validAgent()
	agent["unexpected"] = true

	if err := ValidateAgent(agent); err == nil {
		t.Fatal("unknown field accepted")
	}
}

func TestRejectUnregisteredExtension(t *testing.T) {
	agent := validAgent()
	agent["extensions"] = map[string]any{
		"example.com/provider": map[string]any{},
	}

	if err := ValidateAgent(agent); err == nil {
		t.Fatal("unregistered extension accepted")
	}
}

func TestRejectMissingInstructions(t *testing.T) {
	agent := validAgent()
	delete(agent["spec"].(map[string]any), "instructions")

	if err := ValidateAgent(agent); err == nil {
		t.Fatal("missing instructions accepted")
	}
}
