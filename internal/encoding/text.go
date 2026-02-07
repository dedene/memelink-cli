// Package encoding provides Memegen text encoding and decoding.
//
// Memegen URLs encode special characters in meme text using a set of
// replacement rules. The encoding order is critical: underscores and dashes
// must be escaped BEFORE spaces are converted to underscores.
package encoding

import "strings"

// Encode converts user text to Memegen URL-safe format.
//
// The replacement order matters:
//  1. Escape literal underscores (_) and dashes (-) by doubling them.
//  2. Convert spaces to underscores.
//  3. Replace remaining special characters with tilde or quote sequences.
func Encode(text string) string {
	// Step 1: Escape literal underscores and dashes (must be before spaces).
	text = strings.ReplaceAll(text, "_", "__")
	text = strings.ReplaceAll(text, "-", "--")

	// Step 2: Spaces become single underscores.
	text = strings.ReplaceAll(text, " ", "_")

	// Step 3: Special characters.
	text = strings.ReplaceAll(text, "?", "~q")
	text = strings.ReplaceAll(text, "%", "~p")
	text = strings.ReplaceAll(text, "#", "~h")
	text = strings.ReplaceAll(text, `"`, "''")
	text = strings.ReplaceAll(text, "/", "~s")
	text = strings.ReplaceAll(text, `\`, "~b")
	text = strings.ReplaceAll(text, "\n", "~n")
	text = strings.ReplaceAll(text, "&", "~a")
	text = strings.ReplaceAll(text, "<", "~l")
	text = strings.ReplaceAll(text, ">", "~g")

	return text
}

// Decode converts Memegen URL format back to user text.
//
// This reverses Encode by processing replacements in the opposite order:
//  1. Restore tilde-encoded specials and double-quote sequences.
//  2. Restore underscores (double underscore = literal, single = space).
//  3. Restore dashes (double dash = literal dash).
func Decode(text string) string {
	// Step 1: Tilde-encoded specials and quotes.
	text = strings.ReplaceAll(text, "~q", "?")
	text = strings.ReplaceAll(text, "~p", "%")
	text = strings.ReplaceAll(text, "~h", "#")
	text = strings.ReplaceAll(text, "~a", "&")
	text = strings.ReplaceAll(text, "~l", "<")
	text = strings.ReplaceAll(text, "~g", ">")
	text = strings.ReplaceAll(text, "~b", `\`)
	text = strings.ReplaceAll(text, "~s", "/")
	text = strings.ReplaceAll(text, "~n", "\n")
	text = strings.ReplaceAll(text, "''", `"`)

	// Step 2: Underscores via placeholder technique.
	// __ -> \x00 (placeholder), _ -> space, \x00 -> literal underscore.
	text = strings.ReplaceAll(text, "__", "\x00")
	text = strings.ReplaceAll(text, "_", " ")
	text = strings.ReplaceAll(text, "\x00", "_")

	// Step 3: Dashes via placeholder technique.
	// -- -> \x01 (placeholder), \x01 -> literal dash.
	text = strings.ReplaceAll(text, "--", "\x01")
	text = strings.ReplaceAll(text, "\x01", "-")

	return text
}

// NormalizeQuotes replaces common Unicode smart quotes and dashes with their
// ASCII equivalents so that user input from rich-text editors encodes correctly.
func NormalizeQuotes(text string) string {
	// Right single quote / apostrophe (U+2019).
	text = strings.ReplaceAll(text, "\u2019", "'")
	// Left double quote (U+201C).
	text = strings.ReplaceAll(text, "\u201C", `"`)
	// Right double quote (U+201D).
	text = strings.ReplaceAll(text, "\u201D", `"`)
	// En dash (U+2013) -> single ASCII dash.
	text = strings.ReplaceAll(text, "\u2013", "-")
	// Em dash (U+2014) -> double ASCII dash.
	text = strings.ReplaceAll(text, "\u2014", "--")

	return text
}
