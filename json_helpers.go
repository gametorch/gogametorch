package gametorch

import (
	"encoding/json"
	"fmt"
	"strings"
)

// BoolOrInt is a count that the API may encode either as a boolean (per the
// OpenAPI spec) or as an integer (on the wire). `true` decodes to 1 and
// `false` to 0.
type BoolOrInt int64

// Int64 returns the value as an int64.
func (b BoolOrInt) Int64() int64 { return int64(b) }

// UnmarshalJSON accepts a boolean, an integer or null.
func (b *BoolOrInt) UnmarshalJSON(data []byte) error {
	s := strings.TrimSpace(string(data))
	switch s {
	case "true":
		*b = 1
		return nil
	case "false", "null", "":
		*b = 0
		return nil
	}
	var n int64
	if err := json.Unmarshal(data, &n); err != nil {
		return fmt.Errorf("expected boolean or integer, found %s", s)
	}
	*b = BoolOrInt(n)
	return nil
}

// MarshalJSON encodes the value as a JSON number.
func (b BoolOrInt) MarshalJSON() ([]byte, error) {
	return json.Marshal(int64(b))
}

// StringList is a list of strings that the API may encode leniently: an array
// of strings is used as-is, while null, a number or a boolean yields an empty
// list.
type StringList []string

// UnmarshalJSON accepts an array of strings, null, a number or a boolean.
func (l *StringList) UnmarshalJSON(data []byte) error {
	s := strings.TrimSpace(string(data))
	switch {
	case s == "" || s == "null":
		*l = nil
		return nil
	case s[0] == '[':
		var items []string
		if err := json.Unmarshal(data, &items); err != nil {
			return err
		}
		*l = items
		return nil
	case s[0] == '"':
		return fmt.Errorf("expected string array, found %s", s)
	default:
		// A number or boolean counts as "no suggestions".
		*l = nil
		return nil
	}
}

// DecimalMap is a map of decimal values whose entries may be encoded as strings
// or numbers, with null values dropped.
type DecimalMap map[string]Decimal

// UnmarshalJSON accepts an object, or null for an empty map.
func (m *DecimalMap) UnmarshalJSON(data []byte) error {
	s := strings.TrimSpace(string(data))
	if s == "" || s == "null" {
		*m = nil
		return nil
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	out := make(map[string]Decimal, len(raw))
	for key, value := range raw {
		if strings.TrimSpace(string(value)) == "null" {
			continue
		}
		var d Decimal
		if err := d.UnmarshalJSON(value); err != nil {
			return fmt.Errorf("value for %q: %w", key, err)
		}
		out[key] = d
	}
	*m = out
	return nil
}
