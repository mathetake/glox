package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) > 1 {
		fmt.Println("usage: glox [script]")
		os.Exit(1)
	} else if len(args) == 1 {
		runFile(args[0])
	} else {
		runPrompt()
	}
}

func runFile(name string) {
	bs, err := ioutil.ReadFile(name)
	if err != nil {
		log.Fatal(err)
	}
	run(newInterpreter(), string(bs))
	if hadParserError || hadRuntimeError || hadResolutionError {
		os.Exit(1)
	}
}

func runPrompt() {
	scanner := bufio.NewScanner(os.Stdin)
	it := newInterpreter()
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}
		run(it, scanner.Text())
		hadParserError = false
		hadRuntimeError = false
		hadResolutionError = false
	}
}

func run(it *interpreter, s string) {
	sc := &scanner{source: s}
	ts := sc.scanTokens()
	if hadParserError {
		return
	}
	p := &parser{tokens: ts}
	ss := p.parse()
	if hadParserError {
		return
	}

	r := &resolver{inter: it, scopes: nil}
	r.resolve(ss)
	if hadResolutionError {
		return
	}

	it.interpret(ss)
	if hadRuntimeError {
		return
	}
}
