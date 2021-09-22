package token

type TokenType string

type Token struct {
	Type    TokenType
	Literal string
}

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	IDENT  = "IDENT"
	INT    = "INT"
	FLOAT  = "FLOAT"
	STRING = "STRING"

	// Operators
	ASSIGN   = "="
	PLUS     = "+"
	MINUS    = "-"
	BANG     = "!"
	ASTERISK = "*"
	MODULO   = "%"
	AND      = "&"
	PIPE     = "|"
	CARET    = "^"
	SLASH    = "/"
	LT       = "<"
	GT       = ">"
	POWER    = "**"
	SHIFTR   = ">>"
	SHIFTL   = "<<"

	// Comparasion
	EQUAL   = "=="
	UNEQUAL = "!="
	LTEQ    = "<="
	GTEQ    = ">="
	LAND    = "&&"
	LOR     = "||"

	// Syntax Characters
	COMMA     = ","
	SEMICOLON = ";"
	COLON     = ":"
	LPAREN    = "("
	RPAREN    = ")"
	LBRACE    = "{"
	RBRACE    = "}"
	LBRAKET   = "["
	RBRAKET   = "]"

	// Keywords
	FUNCTION = "FUNCTION"
	RETURN   = "RETURN"
	BREAK    = "BREAK"
	TRUE     = "TRUE"
	FALSE    = "FALSE"
	IF       = "IF"
	ELSE     = "ELSE"
	WHILE    = "WHILE"
	FOR      = "FOR"
)

var keywords = map[string]TokenType{
	"func":   FUNCTION,
	"return": RETURN,
	"true":   TRUE,
	"false":  FALSE,
	"if":     IF,
	"else":   ELSE,
	"while":  WHILE,
	"for":    FOR,
	"break":  BREAK,
}

func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
