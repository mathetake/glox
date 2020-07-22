package main

import (
	"fmt"
	"runtime/debug"
)

type parser struct {
	tokens  []token
	current int
}

func (p *parser) parse() []stmt {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
			fmt.Println(string(debug.Stack()))
		}
	}()

	var ret []stmt

	for !p.isAtEnd() {
		ret = append(ret, p.declaration())
	}
	return ret
}

func (p *parser) expression() expr {
	return p.assignment()
}

func (p *parser) assignment() expr {
	expr := p.or()
	if p.match(tokenTypeEqual) {
		equal := p.previous()
		v := p.assignment()
		if ev, ok := expr.(exprVariable); ok {
			return exprAssign{name: ev.name, value: v}
		} else if get, ok := expr.(exprGet); ok {
			return exprSet{name: get.name, obj: get.obj, value: v}
		}
		reportRuntimeError(equal, "Invalid assignment target")
	}
	return expr
}

func (p *parser) or() expr {
	e := p.and()
	for p.match(tokenTypeOr) {
		op := p.previous()
		right := p.and()
		e = exprLogical{left: e, right: right, operator: op}
	}
	return e
}

func (p *parser) and() expr {
	e := p.equality()
	for p.match(tokenTypeAnd) {
		op := p.previous()
		right := p.equality()
		e = exprLogical{left: e, right: right, operator: op}
	}
	return e
}

func (p *parser) declaration() stmt {
	if p.match(tokenTypeVar) {
		return p.varDeclaration()
	} else if p.match(tokenTypeFun) {
		return p.fun("function")
	} else if p.match(tokenTypeClass) {
		return p.classDeclaration()
	}
	return p.statement()
}

func (p *parser) classDeclaration() stmt {
	name := p.consume(tokenTypeIdentifier, "Expect class name.")

	var super *exprVariable
	if p.match(tokenTypeLess) {
		p.consume(tokenTypeIdentifier, "Expect superclass name.")
		super = &exprVariable{name: p.previous()}
	}
	p.consume(tokenTypeLeftBrace, "Expect '{' before class body.")

	var methods []stmtFunction
	for !p.check(tokenTypeRightBrace) && !p.isAtEnd() {
		methods = append(methods, p.fun("method"))
	}

	p.consume(tokenTypeRightBrace, "Expect '}' after class body")
	return stmtClass{
		methods:    methods,
		name:       name,
		superClass: super,
	}
}

func (p *parser) fun(kind string) stmtFunction {
	name := p.consume(tokenTypeIdentifier, fmt.Sprintf("Expect %s name.", kind))
	p.consume(tokenTypeLeftParen, fmt.Sprintf("Expect '(' after %s name of %v", kind, name))

	var ps []token
	for !p.check(tokenTypeRightParen) {
		ps = append(ps, p.consume(tokenTypeIdentifier, "Expect parameter name."))
		if !p.match(tokenTypeComma) {
			break
		}
	}

	p.consume(tokenTypeRightParen, "Expect ')' after parameters")
	p.consume(tokenTypeLeftBrace, fmt.Sprintf("Expect '{' before %s body", kind))
	body := p.blockStatement().(stmtBlock)
	return stmtFunction{
		params: ps,
		body:   body,
		name:   name,
	}
}

func (p *parser) statement() stmt {
	if p.match(tokenTypePrint) {
		return p.printStatement()
	} else if p.match(tokenTypeLeftBrace) {
		return p.blockStatement()
	} else if p.match(tokenTypeIf) {
		return p.ifStatement()
	} else if p.match(tokenTypeWhile) {
		v := p.whileStatement()
		return v
	} else if p.match(tokenTypeFor) {
		return p.forStatement()
	} else if p.match(tokenTypeReturn) {
		return p.returnStatement()
	}
	return p.expressionStatement()
}

func (p *parser) returnStatement() stmt {
	k := p.previous()
	var exp expr
	if !p.check(tokenTypeSemicolon) {
		exp = p.expression()
	}
	p.consume(tokenTypeSemicolon, "Expect ';' at the end of return statement")
	return &stmtReturn{
		keyword: k,
		value:   exp,
	}
}

func (p *parser) forStatement() stmt {
	p.consume(tokenTypeLeftParen, "Expect '(' after 'while'.")

	var init stmt
	if p.match(tokenTypeSemicolon) {
	} else if p.match(tokenTypeVar) {
		init = p.varDeclaration()
	} else {
		init = p.expressionStatement()
	}

	var cond expr
	if !p.check(tokenTypeSemicolon) {
		cond = p.expression()
	} else {
		cond = exprLiteral{value: true}
	}

	p.consume(tokenTypeSemicolon, "Expect ';' after loop condition.")

	var incr expr
	if !p.check(tokenTypeRightParen) {
		incr = p.expression()
	}

	p.consume(tokenTypeRightParen, "Expect ')' after condition")

	body := p.statement()
	if incr != nil {
		body = stmtBlock{statements: []stmt{body, stmtExpression{e: incr}}}
	}

	body = stmtWhile{
		condition: cond,
		body:      body,
	}
	if init != nil {
		body = stmtBlock{statements: []stmt{init, body}}
	}
	return body
}

func (p *parser) whileStatement() stmt {
	p.consume(tokenTypeLeftParen, "Expect '(' after 'while'.")
	cond := p.expression()
	p.consume(tokenTypeRightParen, "Expect ')' after condition")
	body := p.statement()
	return stmtWhile{
		condition: cond,
		body:      body,
	}
}

func (p *parser) ifStatement() stmt {
	p.consume(tokenTypeLeftParen, "Expect '(' after 'if'.")
	cond := p.expression()
	p.consume(tokenTypeRightParen, "Expect ')' after if condition.")

	thenBr := p.statement()
	var elseBr stmt
	if p.match(tokenTypeElse) {
		elseBr = p.statement()
	}

	return stmtIf{
		condition:  cond,
		thenBranch: thenBr,
		elseBranch: elseBr,
	}
}

func (p *parser) blockStatement() stmt {
	var ss []stmt
	for !p.check(tokenTypeRightBrace) && !p.isAtEnd() {
		ss = append(ss, p.declaration())
	}

	p.consume(tokenTypeRightBrace, "Expect '}' after block.")
	return stmtBlock{statements: ss}
}

func (p *parser) printStatement() stmt {
	e := p.expression()
	p.consume(tokenTypeSemicolon, "Expect ';' after expression.")
	return stmtPrint{e: e}
}

func (p *parser) varDeclaration() stmt {
	n := p.consume(tokenTypeIdentifier, "Expect variable name.")
	var init expr
	if p.match(tokenTypeEqual) {
		init = p.expression()
	}

	p.consume(tokenTypeSemicolon, "Expect ';' after variable declaration.")
	return stmtVar{
		name:        n,
		initializer: init,
	}
}

func (p *parser) expressionStatement() stmt {
	e := p.expression()
	p.consume(tokenTypeSemicolon, "Expect ';' after expression.")
	return stmtExpression{e: e}
}

func (p *parser) equality() expr {
	e := p.comparison()
	for p.match(tokenTypeBangEqual, tokenTypeEqualEqual) {
		op := p.previous()
		right := p.comparison()
		e = exprBinary{
			left:     e,
			right:    right,
			operator: op,
		}
	}
	return e
}

func (p *parser) comparison() expr {
	e := p.addition()
	for p.match(tokenTypeGreater, tokenTypeGreaterEqual, tokenTypeLess, tokenTypeLessEqual) {
		o := p.previous()
		r := p.addition()
		e = exprBinary{left: e, right: r, operator: o}
	}
	return e
}

func (p *parser) addition() expr {
	e := p.multiplication()
	for p.match(tokenTypeMinus, tokenTypePlus) {
		o := p.previous()
		r := p.multiplication()
		e = exprBinary{left: e, right: r, operator: o}
	}
	return e
}

func (p *parser) multiplication() expr {
	e := p.unary()
	for p.match(tokenTypeSlash, tokenTypeStar) {
		o := p.previous()
		r := p.unary()
		e = exprBinary{left: e, right: r, operator: o}
	}
	return e
}

func (p *parser) unary() expr {
	if p.match(tokenTypeBang, tokenTypeMinus) {
		o := p.previous()
		return exprUnary{
			operator: o,
			right:    p.unary(),
		}
	}
	return p.call()
}

func (p *parser) call() expr {
	pr := p.primary()
	for {
		if p.match(tokenTypeLeftParen) {
			pr = p.finishCall(pr)
		} else if p.match(tokenTypeDot) {
			name := p.consume(tokenTypeIdentifier, "Expect property name after '.'.")
			pr = exprGet{name: name, obj: pr}
		} else {
			break
		}
	}
	return pr
}

func (p *parser) finishCall(callee expr) expr {
	var args []expr
	if !p.check(tokenTypeRightParen) {
		for {
			if len(args) >= 255 {
				reportParserError(p.peek(), "Cannot have more than 255 arguments")
			}
			args = append(args, p.expression())
			if !p.match(tokenTypeComma) {
				break
			}
		}
	}
	paren := p.consume(tokenTypeRightParen, "Expect ')' after arguments.")
	return exprCall{
		paren:  paren,
		args:   args,
		callee: callee,
	}
}

func (p *parser) primary() expr {
	switch {
	case p.match(tokenTypeFalse):
		return exprLiteral{value: false}
	case p.match(tokenTypeTrue):
		return exprLiteral{value: true}
	case p.match(tokenTypeNil):
		return exprLiteral{value: nil}
	case p.match(tokenTypeNumber, tokenTypeString):
		return exprLiteral{value: p.previous().literal}
	case p.match(tokenTypeLeftParen):
		e := p.expression()
		p.consume(tokenTypeRightParen, "Expect ')' after expression.")
		return exprGrouping{exp: e}
	case p.match(tokenTypeSuper):
		k := p.previous()
		p.consume(tokenTypeDot, "Expect '.' after 'super'.")
		m := p.consume(tokenTypeIdentifier, "Expect superclass method")
		return exprSuper{keyword: k, method: m}
	case p.match(tokenTypeThis):
		return exprThis{name: p.previous()}
	case p.match(tokenTypeIdentifier):
		return exprVariable{name: p.previous()}
	}

	reportParserError(p.peek(), "Expect expression.")
	return nil
}

func (p *parser) consume(t tokenType, msg string) token {
	if p.check(t) {
		return p.advance()
	}

	reportParserError(p.peek(), msg)
	return token{}
}

func (p *parser) match(ts ...tokenType) bool {
	for _, t := range ts {
		if p.check(t) {
			p.advance()
			return true
		}
	}
	return false
}

func (p *parser) check(t tokenType) bool {
	if p.isAtEnd() {
		return false
	}

	return p.peek().tt == t
}

func (p *parser) advance() token {
	if !p.isAtEnd() {
		p.current++
	}
	return p.previous()
}

func (p *parser) isAtEnd() bool {
	return p.peek().tt == tokenTypeEOF
}

func (p *parser) peek() token {
	return p.tokens[p.current]
}

func (p *parser) previous() token {
	return p.tokens[p.current-1]
}
