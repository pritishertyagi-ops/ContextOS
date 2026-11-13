package assets

import "testing"

func TestParseValidJSON(t *testing.T) {
	_, err := Parse([]byte(`{"name":"ContextOS","version":1}`), "json")
	if err != nil {
		t.Fatalf("valid JSON rejected: %v", err)
	}
}

func TestParseDuplicateJSONKey(t *testing.T) {
	_, err := Parse([]byte(`{"name":"one","name":"two"}`), "json")
	if err == nil {
		t.Fatal("duplicate JSON key was accepted")
	}
}

func TestParseValidYAML(t *testing.T) {
	_, err := Parse([]byte("name: ContextOS\nversion: 1\n"), "yaml")
	if err != nil {
		t.Fatalf("valid YAML rejected: %v", err)
	}
}

func TestParseDuplicateYAMLKey(t *testing.T) {
	_, err := Parse([]byte("name: one\nname: two\n"), "yaml")
	if err == nil {
		t.Fatal("duplicate YAML key was accepted")
	}
}

func TestParseUnsupportedYAMLTag(t *testing.T) {
	_, err := Parse([]byte("value: !custom hello\n"), "yaml")
	if err == nil {
		t.Fatal("unsupported YAML tag was accepted")
	}
}

func TestParseMultipleYAMLDocuments(t *testing.T) {
	_, err := Parse([]byte("---\nname: one\n---\nname: two\n"), "yaml")
	if err == nil {
		t.Fatal("multiple YAML documents were accepted")
	}
}

func TestParseMalformedJSON(t *testing.T) {
	_, err := Parse([]byte(`{"name":`), "json")
	if err == nil {
		t.Fatal("malformed JSON was accepted")
	}
}

func TestParseUnsupportedFormat(t *testing.T) {
	_, err := Parse([]byte("test"), "xml")
	if err == nil {
		t.Fatal("unsupported format was accepted")
	}
}
