package canonical

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"unicode/utf8"

	"github.com/cyberphone/json-canonicalization/go/src/webpki.org/jsoncanonicalizer"
)

// Canonicalize converts a JSON-compatible value into RFC 8785 canonical JSON.
func Canonicalize(value any) ([]byte, error) {
	switch value.(type) {
	case map[string]any, []any:
	default:
		return nil, fmt.Errorf("canonical JSON requires an object or array at the root")
	}

	normalized, err := normalize(value)
	if err != nil {
		return nil, fmt.Errorf("normalize value: %w", err)
	}

	data, err := json.Marshal(normalized)
	if err != nil {
		return nil, fmt.Errorf("marshal JSON: %w", err)
	}

	result, err := jsoncanonicalizer.Transform(data)
	if err != nil {
		return nil, fmt.Errorf("canonicalize JSON: %w", err)
	}

	return result, nil
}

// Hash returns the SHA-256 hexadecimal fingerprint of canonical JSON.
func Hash(value any) (string, error) {
	data, err := Canonicalize(value)
	if err != nil {
		return "", err
	}

	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

func normalize(value any) (any, error) {
	switch v := value.(type) {
	case nil, bool:
		return v, nil

	case string:
		if !utf8.ValidString(v) {
			return nil, fmt.Errorf("string contains invalid UTF-8")
		}
		return v, nil

	case json.Number:
		return normalizeNumber(v.String())

	case int:
		return normalizeNumber(strconv.FormatInt(int64(v), 10))
	case int8:
		return normalizeNumber(strconv.FormatInt(int64(v), 10))
	case int16:
		return normalizeNumber(strconv.FormatInt(int64(v), 10))
	case int32:
		return normalizeNumber(strconv.FormatInt(int64(v), 10))
	case int64:
		return normalizeNumber(strconv.FormatInt(v, 10))

	case uint:
		return normalizeNumber(strconv.FormatUint(uint64(v), 10))
	case uint8:
		return normalizeNumber(strconv.FormatUint(uint64(v), 10))
	case uint16:
		return normalizeNumber(strconv.FormatUint(uint64(v), 10))
	case uint32:
		return normalizeNumber(strconv.FormatUint(uint64(v), 10))
	case uint64:
		return normalizeNumber(strconv.FormatUint(v, 10))

	case float32:
		return normalizeFloat(float64(v))
	case float64:
		return normalizeFloat(v)

	case map[string]any:
		result := make(map[string]any, len(v))
		for key, item := range v {
			if !utf8.ValidString(key) {
				return nil, fmt.Errorf("object key contains invalid UTF-8")
			}
			normalized, err := normalize(item)
			if err != nil {
				return nil, fmt.Errorf("field %q: %w", key, err)
			}
			result[key] = normalized
		}
		return result, nil

	case []any:
		result := make([]any, len(v))
		for i, item := range v {
			normalized, err := normalize(item)
			if err != nil {
				return nil, fmt.Errorf("array index %d: %w", i, err)
			}
			result[i] = normalized
		}
		return result, nil

	default:
		return nil, fmt.Errorf("unsupported value type %T", value)
	}
}

func normalizeNumber(value string) (float64, error) {
	number, err := strconv.ParseFloat(value, 64)
	if err != nil || math.IsInf(number, 0) || math.IsNaN(number) {
		return 0, fmt.Errorf("number %q is outside the supported IEEE-754 range", value)
	}

	const maxSafeInteger = float64(1<<53 - 1)

	if math.Trunc(number) == number && math.Abs(number) > maxSafeInteger {
		return 0, fmt.Errorf(
			"integer %q exceeds the safe IEEE-754 integer range",
			value,
		)
	}

	return number, nil
}

func normalizeFloat(value float64) (float64, error) {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return 0, fmt.Errorf("non-finite floating-point numbers are not supported")
	}
	return value, nil
}
