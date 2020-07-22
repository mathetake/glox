package main

type returnValue struct {
	value interface{}
}

type callable interface {
	call(i *interpreter, args []interface{}) interface{}
	arity() int
}
