package lexer

import (
	"doge/token"
)

type Lexer struct {
	input        string
	position     int
	readPosition int
	ch           byte
}

func New(input string) *Lexer {
	l := &Lexer{input: input}
	l.ReadChar()
	return l
}

func (l *Lexer) ReadChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}

	l.position = l.readPosition
	l.readPosition += 1
}

func (l *Lexer) PeekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	} else {
		return l.input[l.readPosition]
	}
}

func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	l.SkipWhitespace()

	switch l.ch {
	case '=':
		if l.PeekChar() == '=' {
			ch := l.ch
			l.ReadChar()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.EQUAL, Literal: literal}
		} else {
			tok = NewToken(token.ASSIGN, l.ch)
		}
	case '[':
		tok = NewToken(token.LBRAKET, l.ch)
	case ']':
		tok = NewToken(token.RBRAKET, l.ch)
	case '"':
		tok.Type = token.STRING
		tok.Literal = l.ReadString()
	case '+':
		tok = NewToken(token.PLUS, l.ch)
	case '-':
		tok = NewToken(token.MINUS, l.ch)
	case '!':
		if l.PeekChar() == '=' {
			ch := l.ch
			l.ReadChar()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.UNEQUAL, Literal: literal}
		} else {
			tok = NewToken(token.BANG, l.ch)
		}
	case '*':
		tok = NewToken(token.ASTERISK, l.ch)
	case '/':
		tok = NewToken(token.SLASH, l.ch)
	case '<':
		tok = NewToken(token.LT, l.ch)
	case '>':
		tok = NewToken(token.GT, l.ch)
	case ';':
		tok = NewToken(token.SEMICOLON, l.ch)
	case ',':
		tok = NewToken(token.COMMA, l.ch)
	case '(':
		tok = NewToken(token.LPAREN, l.ch)
	case ')':
		tok = NewToken(token.RPAREN, l.ch)
	case '{':
		tok = NewToken(token.LBRACE, l.ch)
	case '}':
		tok = NewToken(token.RBRACE, l.ch)
	case 0:
		tok.Literal = ""
		tok.Type = token.EOF
	default:
		if IsLetter(l.ch) {
			tok.Literal = l.ReadIdentifier()
			tok.Type = token.LookupIdent(tok.Literal)
			return tok
		} else if IsDigit(l.ch) {
			tok.Type = token.INT
			tok.Literal = l.ReadNumber()
			return tok
		} else {
			tok = NewToken(token.ILLEGAL, l.ch)
		}
	}

	l.ReadChar()
	return tok
}

func (l *Lexer) ReadIdentifier() string {
	position := l.position
	for IsLetter(l.ch) {
		l.ReadChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) ReadNumber() string {
	position := l.position
	for IsDigit(l.ch) {
		l.ReadChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) ReadString() string {
	position := l.position + 1
	for {
		l.ReadChar()
		if l.ch == '"' || l.ch == 0 {
			break
		}
	}
	return l.input[position:l.position]
}

func (l *Lexer) SkipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.ReadChar()
	}
}

func NewToken(tokenType token.TokenType, ch byte) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch)}
}

func IsLetter(ch byte) bool {
	return ('a' <= ch && ch <= 'z') || ('A' <= ch && ch <= 'Z') || ch == '_'
}

func IsDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}
