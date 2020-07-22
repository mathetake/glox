package main

import "fmt"

type loxClass struct {
	name       string
	methods    map[string]loxFunction
	superClass *loxClass
}

func (l loxClass) String() string {
	return l.name
}

var _ callable = loxClass{}

func (l loxClass) call(i *interpreter, args []interface{}) interface{} {
	inst := loxInstance{klass: l, fields: map[string]interface{}{}}
	if init := l.findMethod("init"); init != nil {
		init.bind(inst).call(i, args)
	}
	return inst
}

func (l loxClass) arity() int {
	if init := l.findMethod("init"); init != nil {
		return init.arity()
	}
	return 0
}

func (l loxClass) findMethod(name string) *loxFunction {
	m, ok := l.methods[name]
	if ok {
		return &m
	}

	if l.superClass != nil {
		return l.superClass.findMethod(name)
	}

	return nil
}

type loxInstance struct {
	klass  loxClass
	fields map[string]interface{}
}

func (l loxInstance) String() string {
	return l.klass.String() + fmt.Sprintf(" instance: fields: %v", l.fields)
}

func (l loxInstance) get(name token) interface{} {
	v, ok := l.fields[name.lexeme]
	if ok {
		return v
	}

	m := l.klass.findMethod(name.lexeme)
	if m != nil {
		return m.bind(l)
	}

	reportRuntimeError(name, fmt.Sprintf("Undefined property '%s'.", name.lexeme))
	return nil
}

func (l loxInstance) set(name token, v interface{}) {
	l.fields[name.lexeme] = v
}
