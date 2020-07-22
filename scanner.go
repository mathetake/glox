package main

import (
	"strconv"
)

type scanner struct {
	source string
	tokens []token

	start, current, line int
}

func (s *scanner) scanTokens() []token {
	for !s.isAtEnd() {
		s.start = s.current
		s.scanToken()
	}

	s.tokens = append(s.tokens, token{
		tt:      tokenTypeEOF,
		lexeme:  "",
		literal: nil,
		line:    s.line,
	})
	return s.tokens
}

func (s *scanner) isAtEnd() bool {
	return s.current >= len(s.source)
}

func (s *scanner) advance() byte {
	s.current++
	return s.source[s.current-1]
}

func (s *scanner) scanToken() {
	switch c := s.advance(); c {
	case '(':
		s.addToken(tokenTypeLeftParen, nil)
	case ')':
		s.addToken(tokenTypeRightParen, nil)
	case '{':
		s.addToken(tokenTypeLeftBrace, nil)
	case '}':
		s.addToken(tokenTypeRightBrace, nil)
	case ',':
		s.addToken(tokenTypeComma, nil)
	case '.':
		s.addToken(tokenTypeDot, nil)
	case '-':
		s.addToken(tokenTypeMinus, nil)
	case '+':
		s.addToken(tokenTypePlus, nil)
	case ';':
		s.addToken(tokenTypeSemicolon, nil)
	case '*':
		s.addToken(tokenTypeStar, nil)
	case '!':
		if s.match('=') {
			s.addToken(tokenTypeBangEqual, nil)
		} else {
			s.addToken(tokenTypeBang, nil)
		}
	case '=':
		if s.match('=') {
			s.addToken(tokenTypeEqualEqual, nil)
		} else {
			s.addToken(tokenTypeEqual, nil)
		}
	case '>':
		if s.match('=') {
			s.addToken(tokenTypeGreaterEqual, nil)
		} else {
			s.addToken(tokenTypeGreater, nil)
		}
	case '<':
		if s.match('=') {
			s.addToken(tokenTypeLessEqual, nil)
		} else {
			s.addToken(tokenTypeLess, nil)
		}
	case '/':
		if s.match('/') {
			for s.peek() != '\n' && !s.isAtEnd() {
				s.advance()
			}
		} else if s.match('*') {
			for s.peek() != '*' && !s.isAtEnd() {
				s.advance()
			}
			if s.isAtEnd() || s.peekNext() != '/' {
				reportParserError(token{line: s.line}, "unterminated comment")
				return
			}
			s.advance()
			s.advance()
		} else {
			s.addToken(tokenTypeSlash, nil)
		}
	case ' ', '\r', '\t':
	case '\n':
		s.line++
	case '"':
		s.parseString()
	default:
		if s.isDigit(c) {
			s.parseNumber()
		} else if s.isAlpha(c) {
			s.parseIdentifier()
		} else {
			reportParserError(token{line: s.line}, "unexpected character")
		}
	}
}

func (s *scanner) isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func (s *scanner) isAlpha(c byte) bool {
	return (c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z') ||
		c == '_'
}

func (s *scanner) parseString() {
	for s.peek() != '"' && !s.isAtEnd() {
		s.advance()
	}

	if s.isAtEnd() {
		reportParserError(token{line: s.line}, "unterminated string")
		return
	}

	s.advance()
	s.addToken(tokenTypeString, s.source[s.start+1:s.current-1])
}

func (s *scanner) parseNumber() {
	for s.isDigit(s.peek()) {
		s.advance()
	}

	if s.peek() == '.' && s.isDigit(s.peekNext()) {
		s.advance()
		for s.isDigit(s.peek()) {
			s.advance()
		}
	}

	v, err := strconv.ParseFloat(s.source[s.start:s.current], 64)
	if err != nil {
		reportParserError(token{line: s.line}, err.Error())
		return
	}
	s.addToken(tokenTypeNumber, v)
}

func (s *scanner) parseIdentifier() {
	for s.isAlpha(s.peek()) || s.isDigit(s.peek()) {
		s.advance()
	}
	tt, ok := literalToKeywordTokenType[s.source[s.start:s.current]]
	if !ok {
		tt = tokenTypeIdentifier
	}
	s.addToken(tt, nil)
}

func (s *scanner) match(char byte) bool {
	if s.isAtEnd() {
		return false
	} else if s.source[s.current] != char {
		return false
	}
	s.current++
	return true
}

func (s *scanner) peek() byte {
	if s.isAtEnd() {
		return '\000'
	}
	return s.source[s.current]
}

func (s *scanner) peekNext() byte {
	s.current++
	ret := s.peek()
	s.current--
	return ret
}

func (s *scanner) addToken(tt tokenType, literal interface{}) {
	s.tokens = append(s.tokens, token{
		tt:      tt,
		lexeme:  s.source[s.start:s.current],
		literal: literal,
		line:    s.line,
	})
}
