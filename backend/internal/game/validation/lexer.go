package validation

import (
	"regexp"
	"strings"
	"unicode"
)

type Lexer struct {
	input    string
	position int
	current  rune
}

func NewLexer(input string) *Lexer {
	lexer := &Lexer{
		input:    normalizeInput(input),
		position: 0,
	}
	if len(lexer.input) > 0 {
		lexer.current = rune(lexer.input[0])
	}
	return lexer
}

func normalizeInput(input string) string {
	input = strings.ToLower(input)
	input = strings.TrimSpace(input)

	dashRegex := regexp.MustCompile(`[‒–—―−]`)
	input = dashRegex.ReplaceAllString(input, "-")

	input = regexp.MustCompile(`\s*/\s*`).ReplaceAllString(input, "/")
	input = regexp.MustCompile(`\s+`).ReplaceAllString(input, " ")
	input = regexp.MustCompile(`\s*-\s*`).ReplaceAllString(input, " - ")
	input = regexp.MustCompile(`\s+`).ReplaceAllString(input, " ")
	input = strings.TrimSpace(input)

	return input
}

func (l *Lexer) advance() {
	l.position++
	if l.position >= len(l.input) {
		l.current = 0
	} else {
		l.current = rune(l.input[l.position])
	}
}

func (l *Lexer) skipWhitespace() {
	for l.current != 0 && unicode.IsSpace(l.current) {
		l.advance()
	}
}

func (l *Lexer) readWord() string {
	start := l.position
	for l.current != 0 && (unicode.IsLetter(l.current) || unicode.IsDigit(l.current) || l.current == '_') {
		l.advance()
	}
	return l.input[start:l.position]
}

func (l *Lexer) readProvinceWithCoast() string {
	start := l.position

	for l.current != 0 && (unicode.IsLetter(l.current) || unicode.IsDigit(l.current) || l.current == '_') {
		l.advance()
	}

	if l.current == '/' {
		l.advance()
		for l.current != 0 && (unicode.IsLetter(l.current) || unicode.IsDigit(l.current) || l.current == '_') {
			l.advance()
		}
	}

	return l.input[start:l.position]
}

func (l *Lexer) NextToken() Token {
	for l.current != 0 {
		l.skipWhitespace()

		if l.current == 0 {
			break
		}

		position := l.position

		if l.current == '-' {
			l.advance()
			return Token{Type: DASH, Value: "-", Position: position}
		}

		if unicode.IsLetter(l.current) || unicode.IsDigit(l.current) || l.current == '_' {
			word := l.readProvinceWithCoast()
			tokenType := classifyWord(word)
			return Token{Type: tokenType, Value: word, Position: position}
		}

		char := string(l.current)
		l.advance()
		return Token{Type: INVALID, Value: char, Position: position}
	}

	return Token{Type: EOF, Value: "", Position: l.position}
}

func classifyWord(word string) TokenType {
	switch word {
	case "hold", "h":
		return HOLD
	case "support", "supports", "s":
		return SUPPORT
	case "convoy", "convoys", "c":
		return CONVOY
	case "build", "b":
		return BUILD
	case "disband", "remove", "d", "r":
		return DISBAND
	case "army", "a":
		return UNIT_TYPE
	case "fleet", "f":
		return UNIT_TYPE
	default:
		return PROVINCE
	}
}

func Tokenize(input string) []Token {
	lexer := NewLexer(input)
	var tokens []Token

	for {
		token := lexer.NextToken()
		tokens = append(tokens, token)
		if token.Type == EOF {
			break
		}
	}

	return tokens
}

func TokenizeWithErrors(input string) ([]Token, []SyntaxError) {
	tokens := Tokenize(input)
	var errors []SyntaxError

	for _, token := range tokens {
		if token.Type == INVALID {
			errors = append(errors, SyntaxError{
				Position: token.Position,
				Token:    token.Value,
				Expected: []string{"province", "unit type", "order keyword", "-"},
				Message:  "unexpected character",
			})
		}
	}

	return tokens, errors
}
