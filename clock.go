package main

import "time"

type clock struct{}

var _ callable = clock{}

func (c clock) arity() (ret int) { return }

func (c clock) call(*interpreter, []interface{}) interface{} {
	return time.Now().Unix()
}
