package gametorch

import (
	"encoding/json"
	"testing"
)

func TestBoolOrInt(t *testing.T) {
	cases := map[string]int64{
		"true":  1,
		"false": 0,
		"null":  0,
		"4":     4,
		"0":     0,
	}
	for input, want := range cases {
		var b BoolOrInt
		if err := json.Unmarshal([]byte(input), &b); err != nil {
			t.Fatalf("unmarshal %q: %v", input, err)
		}
		if b.Int64() != want {
			t.Errorf("%q = %d, want %d", input, b.Int64(), want)
		}
	}

	encoded, err := json.Marshal(BoolOrInt(4))
	if err != nil {
		t.Fatal(err)
	}
	if string(encoded) != "4" {
		t.Errorf("marshal = %s", encoded)
	}
}

func TestStringListLenient(t *testing.T) {
	var list StringList
	if err := json.Unmarshal([]byte(`["enemy","fauna"]`), &list); err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 || list[0] != "enemy" {
		t.Errorf("list = %v", list)
	}

	for _, input := range []string{"null", "3", "true"} {
		list = StringList{"stale"}
		if err := json.Unmarshal([]byte(input), &list); err != nil {
			t.Fatalf("unmarshal %q: %v", input, err)
		}
		if len(list) != 0 {
			t.Errorf("%q should yield an empty list, got %v", input, list)
		}
	}
}

func TestDecimalMap(t *testing.T) {
	var m DecimalMap
	if err := json.Unmarshal([]byte(`{"api_key:abc":"0","other":5,"skip":null}`), &m); err != nil {
		t.Fatal(err)
	}
	if m["api_key:abc"].String() != "0" {
		t.Errorf("api_key:abc = %s", m["api_key:abc"].String())
	}
	if m["other"].String() != "5" {
		t.Errorf("other = %s", m["other"].String())
	}
	if _, ok := m["skip"]; ok {
		t.Error("null values should be dropped")
	}

	m = DecimalMap{"x": MustDecimal("1")}
	if err := json.Unmarshal([]byte(`null`), &m); err != nil {
		t.Fatal(err)
	}
	if len(m) != 0 {
		t.Errorf("null should clear the map, got %v", m)
	}
}
