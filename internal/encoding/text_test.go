package encoding_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/dedene/memelink-cli/internal/encoding"
)

func TestEncode(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		encoded string
	}{
		// Basic
		{"empty string", "", ""},
		{"simple text", "hello world", "hello_world"},

		// Underscore and dash escaping (must happen before space conversion)
		{"literal underscore", "under_score", "under__score"},
		{"literal dash", "dash-case", "dash--case"},
		{"mixed underscore dash space", "a_b-c d", "a__b--c_d"},

		// All 13 special characters
		{"question mark", "why?", "why~q"},
		{"percent", "100%", "100~p"},
		{"hash", "#1", "~h1"},
		{"double quote", `say "hi"`, "say_''hi''"},
		{"slash", "and/or", "and~sor"},
		{"backslash", `back\slash`, "back~bslash"},
		{"newline", "multi\nline", "multi~nline"},
		{"ampersand", "this & that", "this_~a_that"},
		{"less than", "<3", "~l3"},
		{"greater than", ">9000", "~g9000"},

		// Edge cases
		{"leading space", " leading space", "_leading_space"},
		{"trailing space", "trailing space ", "trailing_space_"},
		{"double special", "??__--", "~q~q____----"},
		{"all specials combined", `a?b%c#d"e/f\g` + "\n" + "h&i<j>k", "a~qb~pc~hd''e~sf~bg~nh~ai~lj~gk"},
		{"only spaces", "   ", "___"},
		{"only underscore", "_", "__"},
		{"only dash", "-", "--"},
		{"consecutive underscores", "__", "____"},
		{"consecutive dashes", "--", "----"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := encoding.Encode(tt.input)
			assert.Equal(t, tt.encoded, got)
		})
	}
}

func TestDecode(t *testing.T) {
	tests := []struct {
		name    string
		encoded string
		decoded string
	}{
		// Basic
		{"empty string", "", ""},
		{"simple text", "hello_world", "hello world"},

		// Underscore and dash restoration
		{"literal underscore", "under__score", "under_score"},
		{"literal dash", "dash--case", "dash-case"},
		{"mixed", "a__b--c_d", "a_b-c d"},

		// All 13 special characters
		{"question mark", "why~q", "why?"},
		{"percent", "100~p", "100%"},
		{"hash", "~h1", "#1"},
		{"double quote", "say_''hi''", `say "hi"`},
		{"slash", "and~sor", "and/or"},
		{"backslash", "back~bslash", `back\slash`},
		{"newline", "multi~nline", "multi\nline"},
		{"ampersand", "this_~a_that", "this & that"},
		{"less than", "~l3", "<3"},
		{"greater than", "~g9000", ">9000"},

		// Edge cases
		{"leading space", "_leading_space", " leading space"},
		{"trailing space", "trailing_space_", "trailing space "},
		{"double special", "~q~q____----", "??__--"},
		// Note: "___" is ambiguous (same as multiple-spaces limitation).
		// Decode sees __ as literal underscore + remaining _ as space -> "_ ".
		{"three underscores ambiguous", "___", "_ "},
		{"only underscore", "__", "_"},
		{"only dash", "--", "-"},
		{"consecutive underscores", "____", "__"},
		{"consecutive dashes", "----", "--"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := encoding.Decode(tt.encoded)
			assert.Equal(t, tt.decoded, got)
		})
	}
}

func TestRoundTrip(t *testing.T) {
	// Decode(Encode(input)) == input for all standard cases.
	inputs := []struct {
		name  string
		input string
	}{
		{"empty", ""},
		{"simple text", "hello world"},
		{"underscore", "under_score"},
		{"dash", "dash-case"},
		{"mixed", "a_b-c d"},
		{"question mark", "why?"},
		{"percent", "100%"},
		{"hash", "#1"},
		{"double quote", `say "hi"`},
		{"slash", "and/or"},
		{"backslash", `back\slash`},
		{"newline", "multi\nline"},
		{"ampersand", "this & that"},
		{"less than", "<3"},
		{"greater than", ">9000"},
		{"leading space", " leading space"},
		{"trailing space", "trailing space "},
		{"double special", "??__--"},
		{"only underscore", "_"},
		{"only dash", "-"},
		{"all specials", `a?b%c#d"e/f\g` + "\n" + "h&i<j>k"},
	}

	for _, tt := range inputs {
		t.Run(tt.name, func(t *testing.T) {
			encoded := encoding.Encode(tt.input)
			decoded := encoding.Decode(encoded)
			assert.Equal(t, tt.input, decoded, "round-trip failed: %q -> %q -> %q", tt.input, encoded, decoded)
		})
	}
}

// TestEncodeMultipleSpacesAmbiguity documents the known limitation where
// multiple consecutive spaces cannot survive a round-trip because "a__b"
// decodes as "a_b" (literal underscore) rather than "a  b" (two spaces).
func TestEncodeMultipleSpacesAmbiguity(t *testing.T) {
	input := "a  b"

	encoded := encoding.Encode(input)
	// Two spaces -> two single underscores, which looks like a double underscore.
	assert.Equal(t, "a__b", encoded, "encoding of multiple spaces")

	decoded := encoding.Decode(encoded)
	// Decode interprets __ as literal underscore, not two spaces.
	assert.Equal(t, "a_b", decoded, "decode interprets __ as literal underscore")
	assert.NotEqual(t, input, decoded, "round-trip is lossy for multiple consecutive spaces")
}

func TestNormalizeQuotes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"no change", "hello world", "hello world"},
		{"right single quote", "it\u2019s", "it's"},
		{"left double quote", "\u201Chello\u201D", `"hello"`},
		{"right double quote only", "say \u201Dhi", `say "hi`},
		{"en dash", "2020\u20132025", "2020-2025"},
		{"em dash", "wait\u2014what", "wait--what"},
		{"mixed smart punctuation", "\u201CHello\u201D \u2014 it\u2019s me", `"Hello" -- it's me`},
		{"empty string", "", ""},
		{"ascii unchanged", `"hello" -- it's me`, `"hello" -- it's me`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := encoding.NormalizeQuotes(tt.input)
			assert.Equal(t, tt.expected, got)
		})
	}
}

// TestNormalizeQuotesThenEncode verifies that normalizing smart quotes before
// encoding produces correct Memegen URLs.
func TestNormalizeQuotesThenEncode(t *testing.T) {
	input := "it\u2019s a \u201Ctest\u201D"
	normalized := encoding.NormalizeQuotes(input)
	encoded := encoding.Encode(normalized)

	assert.Equal(t, "it's_a_''test''", encoded)
}
