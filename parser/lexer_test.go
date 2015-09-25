package parse

import (
	"strings"
	"testing"
)

type lexTest struct {
	input  string
	tokens []token
}

var lexTests = map[string]lexTest{
	"empty":  {"", []token{tokenEOF}},
	"spaces": {" \t\n", []token{tokenIgnored, tokenEOF}},
	"string": {`"\"what's up\", he said."`, []token{tokenStringValue, tokenEOF}},
	"simple": {
		`{
		  user(id: 4) {
			id
			name
			profilePic(size: 12.3e2)
		  }
		}`,
		[]token{
			tokenLeftCurly,
			tokenIgnored,
			tokenName,
			tokenLeftParen,
			tokenName,
			tokenColon,
			tokenIgnored,
			tokenIntValue,
			tokenRightParen,
			tokenIgnored,
			tokenLeftCurly,
			tokenIgnored,
			tokenName,
			tokenIgnored,
			tokenName,
			tokenIgnored,
			tokenName,
			tokenLeftParen,
			tokenName,
			tokenColon,
			tokenIgnored,
			tokenFloatValue,
			tokenRightParen,
			tokenIgnored,
			tokenRightCurly,
			tokenIgnored,
			tokenRightCurly,
			tokenEOF,
		},
	},
}

func compare(a, b []token) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func runLexer(input string) []token {
	lex := newLexer(strings.NewReader(input))
	tokens := make([]token, 0)

	tok, _ := lex.scan()
	tokens = append(tokens, tok)
	for !isTerminal(tok) {
		tok, _ = lex.scan()
		tokens = append(tokens, tok)
	}

	return tokens
}

func TestLexer(t *testing.T) {
	for name, test := range lexTests {
		actual := runLexer(test.input)
		if !compare(actual, test.tokens) {
			t.Errorf("Test %s:\n%v\n%v", name, actual, test.tokens)
		}
	}
}
