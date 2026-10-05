package gametorch

import (
	"encoding/json"
	"testing"
)

func TestDecimalParsePreservesScale(t *testing.T) {
	cases := map[string]string{
		"300.000000000000": "300.000000000000",
		"5.003952358800":   "5.003952358800",
		"0":                "0",
		"0.000000000000":   "0.000000000000",
		"-1.5":             "-1.5",
		".5":               "0.5",
		"5.":               "5",
		"1.5e-3":           "0.0015",
		"1e3":              "1000",
		"+42":              "42",
		" 7.25 ":           "7.25",
	}
	for input, want := range cases {
		got, err := NewDecimalFromString(input)
		if err != nil {
			t.Fatalf("parse %q: %v", input, err)
		}
		if got.String() != want {
			t.Errorf("parse %q = %q, want %q", input, got.String(), want)
		}
	}
}

func TestDecimalParseRejectsInvalid(t *testing.T) {
	for _, input := range []string{"", "abc", "1.2.3", "--1", "1e", "1e2.5"} {
		if _, err := NewDecimalFromString(input); err == nil {
			t.Errorf("expected error for %q", input)
		}
	}
}

func TestDecimalArithmetic(t *testing.T) {
	a := MustDecimal("1.5")
	b := MustDecimal("2.25")

	if got := a.Add(b).String(); got != "3.75" {
		t.Errorf("1.5 + 2.25 = %s, want 3.75", got)
	}
	if got := b.Sub(a).String(); got != "0.75" {
		t.Errorf("2.25 - 1.5 = %s, want 0.75", got)
	}
	if got := a.Mul(MustDecimal("2")).String(); got != "3.0" {
		t.Errorf("1.5 * 2 = %s, want 3.0", got)
	}
	if got := a.Neg().String(); got != "-1.5" {
		t.Errorf("-1.5 = %s", got)
	}

	quotient, err := MustDecimal("100").Div(MustDecimal("100"))
	if err != nil {
		t.Fatal(err)
	}
	if got := quotient.String(); got != "1" {
		t.Errorf("100 / 100 = %s, want 1", got)
	}

	third, err := MustDecimal("1").Div(MustDecimal("3"))
	if err != nil {
		t.Fatal(err)
	}
	if got := third.String(); got != "0.3333333333333333333333333333" {
		t.Errorf("1 / 3 = %s", got)
	}

	if _, err := MustDecimal("1").Div(MustDecimal("0")); err != ErrDecimalDivideByZero {
		t.Errorf("expected divide-by-zero error, got %v", err)
	}
}

func TestDecimalCmp(t *testing.T) {
	if MustDecimal("1.5").Cmp(MustDecimal("1.50")) != 0 {
		t.Error("1.5 should equal 1.50")
	}
	if MustDecimal("1.5").Cmp(MustDecimal("1.6")) != -1 {
		t.Error("1.5 < 1.6")
	}
	if MustDecimal("2").Cmp(MustDecimal("1.9")) != 1 {
		t.Error("2 > 1.9")
	}
}

func TestCreditsToUSD(t *testing.T) {
	if got := CreditsToUSD(MustDecimal("100")).String(); got != "1" {
		t.Errorf("100 credits = $%s, want 1", got)
	}
	if got := CreditsToUSD(MustDecimal("5.003952358800")).String(); got != "0.050039523588" {
		t.Errorf("5.003952358800 credits = $%s", got)
	}
}

func TestDecimalJSON(t *testing.T) {
	// Decimal strings, numbers and null all decode.
	var d Decimal
	if err := json.Unmarshal([]byte(`"300.000000000000"`), &d); err != nil {
		t.Fatal(err)
	}
	if d.String() != "300.000000000000" {
		t.Errorf("string form = %s", d.String())
	}

	if err := json.Unmarshal([]byte(`300`), &d); err != nil {
		t.Fatal(err)
	}
	if d.String() != "300" {
		t.Errorf("number form = %s", d.String())
	}

	if err := json.Unmarshal([]byte(`null`), &d); err != nil {
		t.Fatal(err)
	}
	if !d.IsZero() {
		t.Errorf("null should decode to zero, got %s", d.String())
	}

	encoded, err := json.Marshal(MustDecimal("5.003952358800"))
	if err != nil {
		t.Fatal(err)
	}
	if string(encoded) != `"5.003952358800"` {
		t.Errorf("marshal = %s", encoded)
	}
}
