package gametorch

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

// Decimal is an arbitrary-precision signed decimal number.
//
// GameTorch encodes credit and USD amounts as decimal strings (for example
// "5.003952358800") and occasionally as JSON numbers (for example 300). Decimal
// preserves the exact scale of a parsed string so values round-trip unchanged,
// and it accepts either representation when unmarshalling JSON. It is
// implemented on top of math/big so the SDK needs no third-party dependency.
//
// The zero value is the number 0.
type Decimal struct {
	coef  big.Int
	scale int32
}

// ErrDecimalDivideByZero is returned by Decimal.Div when the divisor is zero.
var ErrDecimalDivideByZero = errors.New("division by zero")

// NewDecimalFromInt64 returns a Decimal equal to n.
func NewDecimalFromInt64(n int64) Decimal {
	var d Decimal
	d.coef.SetInt64(n)
	return d
}

// NewDecimalFromString parses a decimal literal. It accepts an optional sign,
// an optional fractional part and an optional base-10 exponent, for example
// "300", "-1.5", "5.003952358800" or "1.5e-3".
func NewDecimalFromString(s string) (Decimal, error) {
	orig := s
	s = strings.TrimSpace(s)
	if s == "" {
		return Decimal{}, fmt.Errorf("invalid decimal %q: empty", orig)
	}

	neg := false
	switch s[0] {
	case '+':
		s = s[1:]
	case '-':
		neg = true
		s = s[1:]
	}

	var exp int
	if i := strings.IndexAny(s, "eE"); i >= 0 {
		parsed, err := strconv.Atoi(s[i+1:])
		if err != nil {
			return Decimal{}, fmt.Errorf("invalid decimal %q: %w", orig, err)
		}
		exp = parsed
		s = s[:i]
	}

	intPart := s
	fracPart := ""
	if i := strings.IndexByte(s, '.'); i >= 0 {
		intPart = s[:i]
		fracPart = s[i+1:]
	}
	digits := intPart + fracPart
	if digits == "" {
		return Decimal{}, fmt.Errorf("invalid decimal %q", orig)
	}
	for _, r := range digits {
		if r < '0' || r > '9' {
			return Decimal{}, fmt.Errorf("invalid decimal %q", orig)
		}
	}

	coef := new(big.Int)
	if _, ok := coef.SetString(digits, 10); !ok {
		return Decimal{}, fmt.Errorf("invalid decimal %q", orig)
	}

	scale := int32(len(fracPart)) - int32(exp)
	if scale < 0 {
		coef.Mul(coef, pow10(-scale))
		scale = 0
	}
	if neg {
		coef.Neg(coef)
	}
	return Decimal{coef: *coef, scale: scale}, nil
}

// MustDecimal parses s and panics if it is not a valid decimal. It is intended
// for package-level constants and tests.
func MustDecimal(s string) Decimal {
	d, err := NewDecimalFromString(s)
	if err != nil {
		panic(err)
	}
	return d
}

// String returns the decimal in plain notation, preserving the parsed scale.
func (d Decimal) String() string {
	if d.coef.Sign() == 0 {
		if d.scale <= 0 {
			return "0"
		}
		return "0." + strings.Repeat("0", int(d.scale))
	}

	digits := new(big.Int).Abs(&d.coef).String()
	var b strings.Builder
	if d.coef.Sign() < 0 {
		b.WriteByte('-')
	}
	if d.scale <= 0 {
		b.WriteString(digits)
		return b.String()
	}

	sc := int(d.scale)
	if len(digits) <= sc {
		b.WriteString("0.")
		b.WriteString(strings.Repeat("0", sc-len(digits)))
		b.WriteString(digits)
	} else {
		b.WriteString(digits[:len(digits)-sc])
		b.WriteByte('.')
		b.WriteString(digits[len(digits)-sc:])
	}
	return b.String()
}

// IsZero reports whether d is zero.
func (d Decimal) IsZero() bool { return d.coef.Sign() == 0 }

// Sign returns -1, 0 or 1 depending on the sign of d.
func (d Decimal) Sign() int { return d.coef.Sign() }

// Scale returns the number of digits after the decimal point.
func (d Decimal) Scale() int32 { return d.scale }

// Add returns d + o.
func (d Decimal) Add(o Decimal) Decimal {
	a, b, scale := align(d, o)
	var sum big.Int
	sum.Add(a, b)
	return Decimal{coef: sum, scale: scale}
}

// Sub returns d - o.
func (d Decimal) Sub(o Decimal) Decimal {
	a, b, scale := align(d, o)
	var diff big.Int
	diff.Sub(a, b)
	return Decimal{coef: diff, scale: scale}
}

// Mul returns d * o.
func (d Decimal) Mul(o Decimal) Decimal {
	var prod big.Int
	prod.Mul(&d.coef, &o.coef)
	return Decimal{coef: prod, scale: d.scale + o.scale}
}

// Div returns d / o, rounded to at most 28 fractional digits. It returns
// ErrDecimalDivideByZero when o is zero.
func (d Decimal) Div(o Decimal) (Decimal, error) {
	if o.coef.Sign() == 0 {
		return Decimal{}, ErrDecimalDivideByZero
	}
	ratio := new(big.Rat).Quo(d.ratNum(), o.ratNum())
	out, err := NewDecimalFromString(ratio.FloatString(28))
	if err != nil {
		return Decimal{}, err
	}
	return out.normalize(), nil
}

// Neg returns -d.
func (d Decimal) Neg() Decimal {
	var neg big.Int
	neg.Neg(&d.coef)
	return Decimal{coef: neg, scale: d.scale}
}

// Cmp compares d and o, returning -1, 0 or 1.
func (d Decimal) Cmp(o Decimal) int {
	return d.ratNum().Cmp(o.ratNum())
}

// Float64 returns the nearest float64 representation of d.
func (d Decimal) Float64() (float64, error) {
	return strconv.ParseFloat(d.String(), 64)
}

// MarshalJSON encodes d as a JSON string, matching the API's wire format.
func (d Decimal) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

// UnmarshalJSON decodes a decimal from a JSON string or number. A JSON null is
// treated as zero.
func (d *Decimal) UnmarshalJSON(data []byte) error {
	s := strings.TrimSpace(string(data))
	if s == "" || s == "null" {
		*d = Decimal{}
		return nil
	}
	if s[0] == '"' {
		var str string
		if err := json.Unmarshal(data, &str); err != nil {
			return err
		}
		s = str
	}
	parsed, err := NewDecimalFromString(s)
	if err != nil {
		return err
	}
	*d = parsed
	return nil
}

func (d Decimal) ratNum() *big.Rat {
	den := pow10(d.scale)
	return new(big.Rat).SetFrac(&d.coef, den)
}

func (d Decimal) normalize() Decimal {
	if d.scale <= 0 || d.coef.Sign() == 0 {
		if d.coef.Sign() == 0 {
			return Decimal{}
		}
		return d
	}
	ten := big.NewInt(10)
	for d.scale > 0 {
		q, rem := new(big.Int).QuoRem(&d.coef, ten, new(big.Int))
		if rem.Sign() != 0 {
			break
		}
		d.coef = *q
		d.scale--
	}
	return d
}

func align(a, b Decimal) (*big.Int, *big.Int, int32) {
	if a.scale == b.scale {
		return &a.coef, &b.coef, a.scale
	}
	if a.scale > b.scale {
		scaled := new(big.Int).Mul(&b.coef, pow10(a.scale-b.scale))
		return &a.coef, scaled, a.scale
	}
	scaled := new(big.Int).Mul(&a.coef, pow10(b.scale-a.scale))
	return scaled, &b.coef, b.scale
}

func pow10(n int32) *big.Int {
	return new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(n)), nil)
}

// CreditsToUSD converts GameTorch credits to US dollars (100 credits = $1).
func CreditsToUSD(credits Decimal) Decimal {
	usd, err := credits.Div(NewDecimalFromInt64(100))
	if err != nil {
		return Decimal{}
	}
	return usd
}
