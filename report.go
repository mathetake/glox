package main

import "fmt"

var (
	hadParserError     bool
	hadRuntimeError    bool
	hadResolutionError bool
)

func reportParserError(t token, message string) {
	r := func(line int, where, message string) string {
		hadParserError = true
		return fmt.Sprintf("[Parse Error line at %d] Error %s: %s \n", line, where, message)
	}
	if t.tt == tokenTypeEOF {
		panic(r(t.line, " at end", message))
	} else {
		panic(r(t.line, " at '"+t.lexeme+"'", message))
	}
}

func reportRuntimeError(t token, message string) {
	hadRuntimeError = true
	panic(fmt.Sprintf("[Runtime Error at line %d] %s\n", t.line, message))
}

func reportResolutionError(t token, message string) {
	hadResolutionError = true
	panic(fmt.Sprintf("[Resolution Error at line %d] %s\n", t.line, message))
}
