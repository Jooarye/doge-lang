package parser

import (
	"doge/ast"
	"doge/lexer"
	"doge/token"
	"fmt"
	"strconv"
)

const (
	_ int = iota
	LOWEST
	AND_OR
	EQUALS
	LESS_GREATER
	SUM
	PRODUCT
	PREFIX
	POWER
	CALL
	INDEX
)

var precedences = map[token.TokenType]int{
	token.EQUAL:    EQUALS,
	token.UNEQUAL:  EQUALS,
	token.LT:       LESS_GREATER,
	token.GT:       LESS_GREATER,
	token.LTEQ:     EQUALS,
	token.GTEQ:     EQUALS,
	token.LAND:     AND_OR,
	token.LOR:      AND_OR,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.AND:      SUM,
	token.PIPE:     SUM,
	token.SHIFTR:   SUM,
	token.SHIFTL:   SUM,
	token.MODULO:   PRODUCT,
	token.SLASH:    PRODUCT,
	token.ASTERISK: PRODUCT,
	token.CARET:    PRODUCT,
	token.POWER:    POWER,
	token.LPAREN:   CALL,
	token.LBRAKET:  INDEX,
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

type Parser struct {
	l      *lexer.Lexer
	errors []string

	curToken  token.Token
	peekToken token.Token

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}

	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.RegisterPrefix(token.LPAREN, p.ParseGroupedExpression)
	p.RegisterPrefix(token.FUNCTION, p.ParseFunctionLiteral)
	p.RegisterPrefix(token.MINUS, p.ParsePrefixExpression)
	p.RegisterPrefix(token.BANG, p.ParsePrefixExpression)
	p.RegisterPrefix(token.STRING, p.ParseStringLiteral)
	p.RegisterPrefix(token.LBRAKET, p.ParseArrayLiteral)
	p.RegisterPrefix(token.INT, p.ParseIntegerLiteral)
	p.RegisterPrefix(token.FLOAT, p.ParseFloatLiteral)
	p.RegisterPrefix(token.LBRACE, p.ParseHashLiteral)
	p.RegisterPrefix(token.IDENT, p.ParseIdentifier)
	p.RegisterPrefix(token.IF, p.ParseIfExpression)
	p.RegisterPrefix(token.WHILE, p.ParseWhileExpression)
	p.RegisterPrefix(token.FOR, p.ParseForExpression)
	p.RegisterPrefix(token.FALSE, p.ParseBoolean)
	p.RegisterPrefix(token.TRUE, p.ParseBoolean)

	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.RegisterInfix(token.ASTERISK, p.ParseInfixExpression)
	p.RegisterInfix(token.UNEQUAL, p.ParseInfixExpression)
	p.RegisterInfix(token.LBRAKET, p.ParseIndexExpression)
	p.RegisterInfix(token.MODULO, p.ParseInfixExpression)
	p.RegisterInfix(token.SHIFTL, p.ParseInfixExpression)
	p.RegisterInfix(token.SHIFTR, p.ParseInfixExpression)
	p.RegisterInfix(token.LPAREN, p.ParseCallExpression)
	p.RegisterInfix(token.MINUS, p.ParseInfixExpression)
	p.RegisterInfix(token.SLASH, p.ParseInfixExpression)
	p.RegisterInfix(token.CARET, p.ParseInfixExpression)
	p.RegisterInfix(token.EQUAL, p.ParseInfixExpression)
	p.RegisterInfix(token.POWER, p.ParseInfixExpression)
	p.RegisterInfix(token.LAND, p.ParseInfixExpression)
	p.RegisterInfix(token.PLUS, p.ParseInfixExpression)
	p.RegisterInfix(token.PIPE, p.ParseInfixExpression)
	p.RegisterInfix(token.LTEQ, p.ParseInfixExpression)
	p.RegisterInfix(token.GTEQ, p.ParseInfixExpression)
	p.RegisterInfix(token.LOR, p.ParseInfixExpression)
	p.RegisterInfix(token.AND, p.ParseInfixExpression)
	p.RegisterInfix(token.LT, p.ParseInfixExpression)
	p.RegisterInfix(token.GT, p.ParseInfixExpression)

	p.NextToken()
	p.NextToken()

	return p
}

func (p *Parser) RegisterPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) RegisterInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) NextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) PeekError(t token.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead", t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

func (p *Parser) PeekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}

	return LOWEST
}

func (p *Parser) CurPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}

	return LOWEST
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for p.curToken.Type != token.EOF {
		stmt := p.ParseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.NextToken()
	}

	return program
}

func (p *Parser) ParseArrayLiteral() ast.Expression {
	array := &ast.ArrayLiteral{Token: p.curToken}

	array.Elements = p.ParseExpressionList(token.RBRAKET)

	return array
}

func (p *Parser) ParseHashLiteral() ast.Expression {
	hash := &ast.HashLiteral{Token: p.curToken}
	hash.Pairs = make(map[ast.Expression]ast.Expression)

	for !p.PeekTokenIs(token.RBRACE) {
		p.NextToken()
		key := p.ParseExpression(LOWEST)

		if !p.ExpectPeek(token.COLON) {
			return nil
		}

		p.NextToken()
		value := p.ParseExpression(LOWEST)

		hash.Pairs[key] = value

		if !p.PeekTokenIs(token.RBRACE) && !p.ExpectPeek(token.COMMA) {
			return nil
		}
	}

	if !p.ExpectPeek(token.RBRACE) {
		return nil
	}

	return hash
}

func (p *Parser) ParseExpressionList(end token.TokenType) []ast.Expression {
	list := []ast.Expression{}

	if p.PeekTokenIs(end) {
		p.NextToken()
		return list
	}

	p.NextToken()
	list = append(list, p.ParseExpression(LOWEST))

	for p.PeekTokenIs(token.COMMA) {
		p.NextToken()
		p.NextToken()
		list = append(list, p.ParseExpression(LOWEST))
	}

	if !p.ExpectPeek(end) {
		return nil
	}

	return list
}

func (p *Parser) ParseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.RETURN:
		return p.ParseReturnStatement()
	case token.BREAK:
		if p.PeekTokenIs(token.SEMICOLON) {
			p.NextToken()
		}

		return &ast.BreakStatement{}
	default:
		return p.ParseExpressionStatement()
	}
}

func (p *Parser) ParseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) ParseIdentifier() ast.Expression {
	ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if p.PeekTokenIs(token.ASSIGN) {
		p.NextToken()
		expression := &ast.AssignExpression{
			Token: p.curToken,
			Left:  ident,
		}

		precedence := p.CurPrecedence()
		p.NextToken()
		expression.Right = p.ParseExpression(precedence)

		return expression
	}

	return ident
}

func (p *Parser) ParseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.NoPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()

	for !p.PeekTokenIs(token.SEMICOLON) && precedence < p.PeekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		p.NextToken()
		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *Parser) ParseIndexExpression(left ast.Expression) ast.Expression {
	exp := &ast.IndexExpression{Token: p.curToken, Left: left}

	p.NextToken()
	exp.Index = p.ParseExpression(LOWEST)

	if !p.ExpectPeek(token.RBRAKET) {
		return nil
	}

	return exp
}

func (p *Parser) ParsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}

	p.NextToken()

	expression.Right = p.ParseExpression(PREFIX)

	return expression
}

func (p *Parser) ParseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}

	precedence := p.CurPrecedence()
	p.NextToken()
	expression.Right = p.ParseExpression(precedence)

	return expression
}

func (p *Parser) ParseAssignExpression() ast.Expression {
	expression := &ast.AssignExpression{
		Token: p.curToken,
	}

	precedence := p.CurPrecedence()
	p.NextToken()
	expression.Right = p.ParseExpression(precedence)

	return expression
}

func (p *Parser) ParseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curToken}

	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value

	return lit
}

func (p *Parser) ParseFloatLiteral() ast.Expression {
	lit := &ast.FloatLiteral{Token: p.curToken}

	value, err := strconv.ParseFloat(p.curToken.Literal, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as float", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value

	return lit
}

func (p *Parser) ParseWhileExpression() ast.Expression {
	expression := &ast.WhileExpression{Token: p.curToken}

	if !p.ExpectPeek(token.LPAREN) {
		return nil
	}

	p.NextToken()
	expression.Condition = p.ParseExpression(LOWEST)

	if !p.ExpectPeek(token.RPAREN) {
		return nil
	}

	if !p.ExpectPeek(token.LBRACE) {
		return nil
	}

	expression.Consequence = p.ParseBlockStatement()

	return expression
}

func (p *Parser) ParseForExpression() ast.Expression {
	expression := &ast.ForExpression{Token: p.curToken}

	if !p.ExpectPeek(token.LPAREN) {
		return nil
	}

	p.NextToken()
	expression.Initial = p.ParseExpression(LOWEST)

	if !p.ExpectPeek(token.SEMICOLON) {
		return nil
	}

	p.NextToken()
	expression.Condition = p.ParseExpression(LOWEST)

	if !p.ExpectPeek(token.SEMICOLON) {
		return nil
	}

	p.NextToken()
	expression.Increment = p.ParseExpression(LOWEST)

	if !p.ExpectPeek(token.RPAREN) {
		return nil
	}

	if !p.ExpectPeek(token.LBRACE) {
		return nil
	}

	expression.Consequence = p.ParseBlockStatement()

	return expression
}

func (p *Parser) ParseIfExpression() ast.Expression {
	expression := &ast.IfExpression{Token: p.curToken}

	if !p.ExpectPeek(token.LPAREN) {
		return nil
	}

	p.NextToken()
	expression.Condition = p.ParseExpression(LOWEST)

	if !p.ExpectPeek(token.RPAREN) {
		return nil
	}

	if !p.ExpectPeek(token.LBRACE) {
		return nil
	}

	expression.Consequence = p.ParseBlockStatement()

	if p.PeekTokenIs(token.ELSE) {
		p.NextToken()

		if !p.ExpectPeek(token.LBRACE) {
			return nil
		}

		expression.Alternative = p.ParseBlockStatement()
	}

	return expression
}

func (p *Parser) ParseFunctionLiteral() ast.Expression {
	lit := &ast.FunctionLiteral{Token: p.curToken}

	if !p.ExpectPeek(token.LPAREN) {
		return nil
	}

	lit.Parameters = p.ParseFunctionParameters()

	if !p.ExpectPeek(token.LBRACE) {
		return nil
	}

	lit.Body = p.ParseBlockStatement()

	return lit
}

func (p *Parser) ParseFunctionParameters() []*ast.Identifier {
	identifiers := []*ast.Identifier{}

	if p.PeekTokenIs(token.RPAREN) {
		p.NextToken()
		return identifiers
	}

	p.NextToken()

	ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	identifiers = append(identifiers, ident)

	for p.PeekTokenIs(token.COMMA) {
		p.NextToken()
		p.NextToken()
		tempIdent := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		identifiers = append(identifiers, tempIdent)
	}

	if !p.ExpectPeek(token.RPAREN) {
		return nil
	}

	return identifiers
}

func (p *Parser) ParseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.curToken, Function: function}
	exp.Arguments = p.ParseExpressionList(token.RPAREN)
	return exp
}

func (p *Parser) ParseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken}
	block.Statements = []ast.Statement{}

	p.NextToken()

	for !p.CurTokenIs(token.RBRACE) && !p.CurTokenIs(token.EOF) {
		stmt := p.ParseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.NextToken()
	}

	return block
}

func (p *Parser) ParseGroupedExpression() ast.Expression {
	p.NextToken()

	exp := p.ParseExpression(LOWEST)

	if !p.ExpectPeek(token.RPAREN) {
		return nil
	}

	return exp
}

func (p *Parser) ParseBoolean() ast.Expression {
	return &ast.Boolean{
		Token: p.curToken,
		Value: p.CurTokenIs(token.TRUE),
	}
}

func (p *Parser) ParseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}

	stmt.Expression = p.ParseExpression(LOWEST)

	if p.PeekTokenIs(token.SEMICOLON) {
		p.NextToken()
	}

	return stmt
}

func (p *Parser) ParseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}

	p.NextToken()

	stmt.ReturnValue = p.ParseExpression(LOWEST)

	if p.PeekTokenIs(token.SEMICOLON) {
		p.NextToken()
	}

	return stmt
}

func (p *Parser) NoPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.errors = append(p.errors, msg)
}

func (p *Parser) CurTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) PeekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) ExpectPeek(t token.TokenType) bool {
	if p.PeekTokenIs(t) {
		p.NextToken()
		return true
	} else {
		p.PeekError(t)
		return false
	}
}
