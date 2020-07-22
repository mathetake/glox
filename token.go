package main

import "fmt"

type tokenType int

const (
	// single char
	tokenTypeLeftParen tokenType = iota
	tokenTypeRightParen
	tokenTypeLeftBrace
	tokenTypeRightBrace
	tokenTypeComma
	tokenTypeDot
	tokenTypeMinus
	tokenTypePlus
	tokenTypeSemicolon
	tokenTypeSlash
	tokenTypeStar

	// one or two chars
	tokenTypeBang
	tokenTypeBangEqual
	tokenTypeEqual
	tokenTypeEqualEqual
	tokenTypeGreater
	tokenTypeGreaterEqual
	tokenTypeLess
	tokenTypeLessEqual

	// literals
	tokenTypeIdentifier
	tokenTypeString
	tokenTypeNumber

	// keywords
	tokenTypeAnd
	tokenTypeClass
	tokenTypeElse
	tokenTypeFalse
	tokenTypeFun
	tokenTypeFor
	tokenTypeIf
	tokenTypeNil
	tokenTypeOr
	tokenTypePrint
	tokenTypeReturn
	tokenTypeSuper
	tokenTypeThis
	tokenTypeTrue
	tokenTypeVar
	tokenTypeWhile

	tokenTypeEOF
)

type token struct {
	tt      tokenType
	lexeme  string
	literal interface{}
	line    int
}

var literalToKeywordTokenType = map[string]tokenType{
	"and":    tokenTypeAnd,
	"class":  tokenTypeClass,
	"else":   tokenTypeElse,
	"false":  tokenTypeFalse,
	"for":    tokenTypeFor,
	"fun":    tokenTypeFun,
	"if":     tokenTypeIf,
	"nil":    tokenTypeNil,
	"or":     tokenTypeOr,
	"print":  tokenTypePrint,
	"return": tokenTypeReturn,
	"super":  tokenTypeSuper,
	"this":   tokenTypeThis,
	"true":   tokenTypeTrue,
	"var":    tokenTypeVar,
	"while":  tokenTypeWhile,
}

func (t *token) toString() string {
	return fmt.Sprintf("%d %s %v", t.tt, t.lexeme, t.literal)
}
