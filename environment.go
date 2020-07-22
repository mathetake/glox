package main

import "fmt"

type environment struct {
	enclosing *environment
	values    map[string]interface{}
}

func newEnvironment() *environment {
	return &environment{values: map[string]interface{}{}}
}

func newEnvironmentWithParent(parent *environment) *environment {
	return &environment{values: map[string]interface{}{}, enclosing: parent}
}

func (e *environment) define(name string, v interface{}) {
	e.values[name] = v
}

func (e *environment) get(name token) interface{} {
	v, ok := e.values[name.lexeme]
	if !ok && e.enclosing != nil {
		return e.enclosing.get(name)
	} else if !ok {
		reportRuntimeError(name, fmt.Sprintf("Undefined variable: '%s'", name.lexeme))
	}
	return v
}

func (e *environment) getAt(dist int, name token) interface{} {
	return e.ancestor(dist).get(name)
}

func (e *environment) assign(name token, v interface{}) {
	_, ok := e.values[name.lexeme]
	if !ok && e.enclosing != nil {
		e.enclosing.assign(name, v)
		return
	} else if !ok {
		reportRuntimeError(name, fmt.Sprintf("Undefined variable: '%s'", name.lexeme))
	}
	e.values[name.lexeme] = v
}

func (e *environment) assignAt(dist int, name token, v interface{}) {
	e.ancestor(dist).assign(name, v)
}

func (e *environment) ancestor(dist int) *environment {
	ret := e
	for i := 0; i < dist; i++ {
		ret = ret.enclosing
	}
	return ret
}
