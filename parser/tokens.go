package parse

import (
	"unicode"
)

var eof = rune(0)

//go:generate stringer -type=token
type token uint8

const (
	// Tokens which end the stream (<= tokenEOF)
	tokenIllegal token = iota
	tokenEOF

	// Ignored
	tokenIgnored
	tokenComma
	tokenWhitespace
	tokenLineTerminator

	// Tokens
	tokenName // /[_A-Za-z][_0-9A-Za-z]*/

	// Values
	tokenVariableValue
	tokenListValue
	tokenObjectValue
	tokenBooleanValue
	tokenEnumValue
	tokenIntValue
	tokenFloatValue
	tokenStringValue

	// Punctuators
	tokenAt
	tokenColon
	tokenDollar
	tokenEqual
	tokenExclam
	tokenLeftBracket
	tokenRightBracket
	tokenLeftCurly
	tokenRightCurly
	tokenLeftParen
	tokenRightParen
	tokenSpread
)

func isTerminal(tok token) bool {
	return tok <= tokenEOF
}

func isIgnored(tok token) bool {
	return tok >= tokenIgnored && tok <= tokenLineTerminator
}

func isValue(tok token) bool {
	return tok >= tokenVariableValue && tok <= tokenStringValue
}

func isWhitespace(ch rune) bool {
	return (ch == '\u0009' || // tab
		ch == '\u000b' || // vert tab
		ch == '\u000c' || // form feed
		ch == '\u0020' || // space
		ch == '\u00a0') // nbsp
}

func isLineTerminator(ch rune) bool {
	return (ch == '\u000a' || // new line
		ch == '\u000d' || // CR
		ch == '\u2028' || // line sep
		ch == '\u2029') // paragraph sep
}

func isIgnoredRune(ch rune) bool {
	return ch == ',' || isWhitespace(ch) || isLineTerminator(ch)
}

func isLetter(ch rune) bool {
	return unicode.IsLetter(ch)
}

func isIdentifier(ch rune) bool {
	return unicode.IsDigit(ch) || unicode.IsLetter(ch) || ch == '_'
}

func isDigit(ch rune) bool {
	return unicode.IsDigit(ch)
}
