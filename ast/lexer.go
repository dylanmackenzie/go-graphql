package ast

import (
	"bufio"
	"bytes"
	"io"
)

type pos uint64
type lexer struct {
	r *bufio.Reader

	name       string // name of input for debugging
	leftDelim  token
	rightDelim token
	pos        pos
	start      pos
	width      pos

	lastSuccess bool
	lastToken   token
	lastLiteral string
}

// newLexer returns a new lexer by buffering the given io.Reader
func newLexer(r io.Reader) *lexer {
	return &lexer{
		r:           bufio.NewReader(r),
		lastSuccess: true,
	}
}

// read reads the next rune from the bufferred reader.
// Returns the rune(0) if an error occurs (or io.EOF is returned).
func (l *lexer) read() rune {
	ch, width, err := l.r.ReadRune()
	if err != nil {
		return eof
	}

	l.width = pos(width)
	l.pos += l.width
	return ch
}

// unread places the previously read rune back on the reader.
func (l *lexer) unread() {
	_ = l.r.UnreadRune()
	l.pos -= l.width
}

// peek returns but does not consume the next rune in the input
func (l *lexer) peek() rune {
	r := l.read()
	l.unread()
	return r
}

// scanIgnored consumes all commas, whitespace and line terminators
func (l *lexer) scanIgnored() (tok token, lit string) {
	buf := new(bytes.Buffer)

	for ch := l.read(); isIgnoredRune(ch); {
		buf.WriteRune(ch)
		ch = l.read()
	}

	l.unread()
	return tokenIgnored, buf.String()
}

func (l *lexer) scanName() (tok token, lit string) {
	buf := new(bytes.Buffer)

	// Consume first character /[_a-zA-Z]/
	ch := l.read()
	buf.WriteRune(ch)

	for ch = l.read(); isIdentifier(ch); {
		buf.WriteRune(ch)
		ch = l.read()
	}

	l.unread()
	return tokenName, buf.String()
}

func (l *lexer) scanNumber() (tok token, lit string) {
	tokenType := tokenIntValue
	buf := new(bytes.Buffer)
	ch := l.read()

	// Consume the negative sign
	if ch == '-' {
		buf.WriteRune(ch)
		ch = l.read()

		if !isDigit(ch) {
			return tokenIllegal, "Negative sign without corresponding literal"
		}
	}

	// Ensure that nonzero non numeric literals do not begin with 0
	if ch == '0' {
		if peek := l.peek(); isDigit(peek) || peek == 'e' || peek == 'E' {
			return tokenIllegal, "Non zero numbers must not start with a zero"
		}
	}

	// Integral part of number
	for isDigit(ch) {
		buf.WriteRune(ch)
		ch = l.read()
	}

	// Floating point fraction
	if ch == '.' {
		tokenType = tokenFloatValue
		buf.WriteRune(ch)
		ch = l.read()
		for isDigit(ch) {
			buf.WriteRune(ch)
			ch = l.read()
		}
	}

	// Floating point exponent
	if ch == 'e' || ch == 'E' {
		tokenType = tokenFloatValue
		buf.WriteRune(ch)
		ch = l.read()

		// Consume exponent sign
		if ch == '+' || ch == '-' {
			buf.WriteRune(ch)
			ch = l.read()
		}

		for isDigit(ch) {
			buf.WriteRune(ch)
			ch = l.read()
		}
	}

	l.unread()
	return tokenType, buf.String()
}

func (l *lexer) scanString() (tok token, lit string) {
	buf := new(bytes.Buffer)

	// Consume opening quote character
	ch := l.read()

	for ch = l.read(); ch != '"'; {
		if isLineTerminator(ch) {
			return tokenIllegal, "Line terminator in string literal"
		}

		// Consume escaped characters
		if ch == '\\' {
			buf.WriteRune(ch)
			ch = l.read()
		}

		buf.WriteRune(ch)
		ch = l.read()
	}

	return tokenStringValue, buf.String()
}

func (l *lexer) last() (tok token, lit string) {
	return l.lastToken, l.lastLiteral
}

// Scan returns the next token and literal value.
func (l *lexer) scan() (tok token, lit string) {
	ch := l.read()

	switch {
	case isIgnoredRune(ch):
		l.unread()
		tok, lit = l.scanIgnored()
	case isLetter(ch) || ch == '_':
		l.unread()
		tok, lit = l.scanName()
	case isDigit(ch) || ch == '-':
		l.unread()
		tok, lit = l.scanNumber()
	case ch == '"':
		l.unread()
		tok, lit = l.scanString()
	case ch == '$':
		_, lit = l.scanName()
		tok = tokenVariableValue
	case ch == eof:
		tok, lit = tokenEOF, ""
	case ch == '!':
		tok, lit = tokenExclam, "!"
	case ch == '(':
		tok, lit = tokenLeftParen, "("
	case ch == ')':
		tok, lit = tokenRightParen, ")"
	case ch == ':':
		tok, lit = tokenColon, ":"
	case ch == '=':
		tok, lit = tokenEqual, "="
	case ch == '@':
		tok, lit = tokenAt, "@"
	case ch == '[':
		tok, lit = tokenLeftBracket, "["
	case ch == ']':
		tok, lit = tokenRightBracket, "]"
	case ch == '{':
		tok, lit = tokenLeftCurly, "{"
	case ch == '}':
		tok, lit = tokenRightCurly, "}"
	case ch == '.':
		if l.read() != '.' || l.read() != '.' {
			tok, lit = tokenIllegal, "Periods must be part of a spread operator"
		} else {
			tok, lit = tokenSpread, "..."
		}
	default:
		tok, lit = tokenIllegal, "Unrecognized token: "+string(ch)
	}

	l.lastToken = tok
	l.lastLiteral = lit
	return tok, lit
}

func (l *lexer) consumeIgnored() (token, string) {
	tok, lit := l.scan()
	for isIgnored(tok) {
		tok, lit = l.scan()
	}

	return tok, lit
}

func (l *lexer) Assert(asserted token) bool {
	lastTok, _ := l.last()
	return lastTok == asserted
}

func (l *lexer) Optional(optional token) bool {
	tok, _ := l.last()
	if l.lastSuccess {
		tok, _ = l.consumeIgnored()
	}

	isSuccess := (tok == optional)
	l.lastSuccess = isSuccess
	return isSuccess
}

func (l *lexer) Advance() (token, string) {
	if l.lastSuccess {
		return l.consumeIgnored()
	} else {
		l.lastSuccess = true
		return l.last()
	}
}

func (l *lexer) Discard() {
	if l.lastSuccess {
		l.consumeIgnored()
	}
	l.lastSuccess = false
}

func (l *lexer) Expect(expected token) bool {
	tok, _ := l.last()
	if l.lastSuccess {
		tok, _ = l.consumeIgnored()
	}

	l.lastSuccess = true
	return tok == expected
}

func (l *lexer) ExpectFunc(f func(token) bool) bool {
	tok, _ := l.last()
	if l.lastSuccess {
		tok, _ = l.consumeIgnored()
	}

	l.lastSuccess = true
	return f(tok)
}
