package main

import "fmt"

type loxFunction struct {
	declaration   stmtFunction
	closure       *environment
	isInitializer bool
}

var _ callable = loxFunction{}

func (l loxFunction) call(i *interpreter, args []interface{}) (v interface{}) {
	env := newEnvironmentWithParent(l.closure)
	for i, arg := range args {
		env.define(l.declaration.params[i].lexeme, arg)
	}
	defer func() {
		if raw := recover(); raw != nil {
			rawValue, ok := raw.(returnValue)
			if !ok {
				panic(raw)
			}
			v = rawValue.value
		}

		if l.isInitializer {
			v = l.closure.getAt(0, token{lexeme: "this"})
		}
	}()
	i.executeBlock(l.declaration.body, env)
	return
}

func (l loxFunction) arity() int {
	return len(l.declaration.params)
}

func (l loxFunction) String() string {
	return fmt.Sprintf("<fn %s>", l.declaration.name.lexeme)
}

func (l loxFunction) bind(inst loxInstance) loxFunction {
	env := newEnvironmentWithParent(l.closure)
	env.define("this", inst)
	return loxFunction{
		declaration:   l.declaration,
		closure:       env,
		isInitializer: l.isInitializer,
	}
}
