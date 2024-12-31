package formula

import (
	"fmt"
	"unicode"
)

// Token types
type TokenType string

const (
	NUMBER        TokenType = "NUMBER"
	OPERATOR      TokenType = "OPERATOR"
	COMPARISON_OP TokenType = "COMPARISON_OP"
	SEMICOLON     TokenType = "SEMICOLON"
	ASSIGN_OP     TokenType = "ASSIGN_OP"
	IDENTIFIER    TokenType = "IDENTIFIER"
	COMMA         TokenType = "COMMA"
	EOF           TokenType = "EOF"
	LPAREN        TokenType = "LPAREN" // 新增 LPAREN
	RPAREN        TokenType = "RPAREN" // 新增 RPAREN
)

// Token structure
type Token struct {
	Type  TokenType
	Value string
}

// Lexer structure
type Lexer struct {
	input  string
	cursor int
}

func NewLexer(input string) *Lexer {
	return &Lexer{input: input}
}

// Consume a character from the input; return EOF token type when end of input is reached.
func (l *Lexer) consume() (rune, TokenType) {
	if l.cursor >= len(l.input) {
		return 0, EOF // Return EOF token type at the end of input
	}
	char := rune(l.input[l.cursor])
	l.cursor++
	return char, "" // "" means no special token type
}

// Peek at the next character without consuming it; return EOF token type when end of input is reached.
func (l *Lexer) peek() (rune, TokenType) {
	if l.cursor >= len(l.input) {
		return 0, EOF // Return EOF token type at the end of input
	}
	return rune(l.input[l.cursor]), "" // "" means no special token type
}

// Tokenize the input expression
func (l *Lexer) Tokenize() ([]Token, error) {
	tokens := []Token{}
	for {
		char, tokenType := l.consume()
		if tokenType == EOF {
			break // End of input
		}

		if unicode.IsSpace(char) {
			continue // Skip whitespace
		}

		switch {
		case unicode.IsDigit(char) || char == '.':
			// Handle numbers
			numStr := ""
			for {
				numStr += string(char)
				nextChar, tokenType := l.peek() // Peek at the next character, checking for EOF
				if tokenType == EOF || (!unicode.IsDigit(nextChar) && nextChar != '.') {
					break // End of number or EOF
				}
				char, _ = l.consume() // Consume the digit or '.'
			}
			tokens = append(tokens, Token{Type: NUMBER, Value: numStr})

		case unicode.IsLetter(char):
			// Handle identifiers
			identStr := ""
			for {
				identStr += string(char)
				nextChar, tokenType := l.peek() // Peek at the next character, checking for EOF
				if tokenType == EOF || (!unicode.IsLetter(nextChar) && !unicode.IsDigit(nextChar)) {
					break // End of identifier or EOF
				}
				char, _ = l.consume() // Consume the character
			}
			tokens = append(tokens, Token{Type: IDENTIFIER, Value: identStr})

		case string(char) == "+" || string(char) == "-" || string(char) == "*" || string(char) == "/" || char == ',':
			// Handle operators and parentheses
			tokens = append(tokens, Token{Type: OPERATOR, Value: string(char)})
		case string(char) == "(":
			tokens = append(tokens, Token{Type: LPAREN, Value: "("})
		case string(char) == ")":
			tokens = append(tokens, Token{Type: RPAREN, Value: ")"})
		case string(char) == ":":
			nextChar, _ := l.peek()
			if nextChar == '=' {
				tokens = append(tokens, Token{Type: ASSIGN_OP, Value: ":="})
				l.consume()
			} else {
				tokens = append(tokens, Token{Type: ASSIGN_OP, Value: ":"})
			}
		case string(char) == "<" || string(char) == ">" || string(char) == "=" || string(char) == "!":
			opStr := string(char)
			nextChar, _ := l.peek()
			if nextChar == '=' && (char == '<' || char == '>' || char == '=' || char == '!') {
				opStr += string(nextChar)
				l.consume()
			}
			tokens = append(tokens, Token{Type: COMPARISON_OP, Value: opStr})
		case string(char) == ";": // 添加对分号的处理
			tokens = append(tokens, Token{Type: SEMICOLON, Value: ";"})
		default:
			return nil, fmt.Errorf("invalid character: %c", char)
		}
	}
	return tokens, nil
}
