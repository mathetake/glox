package main

import (
	"fmt"
	"runtime/debug"
)

type interpreter struct {
	globals, env *environment
	locals       map[expr]int
}

func newInterpreter() *interpreter {
	gs := newEnvironment()
	gs.define("clock", clock{})
	return &interpreter{
		globals: gs,
		env:     gs,
		locals:  map[expr]int{},
	}
}

var (
	_ exprVisitor = &interpreter{}
	_ stmtVisitor = &interpreter{}
)

func (i *interpreter) interpret(ss []stmt) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
			fmt.Println(string(debug.Stack()))
		}
	}()

	for _, s := range ss {
		i.execute(s)
	}
	return
}

func (i *interpreter) visitBinaryExpr(e exprBinary) interface{} {
	left := i.evaluate(e.left)
	right := i.evaluate(e.right)
	switch e.operator.tt {
	case tokenTypeMinus:
		i.checkNumberOperands(e.operator, left, right)
		return left.(float64) - right.(float64)
	case tokenTypeSlash:
		i.checkNumberOperands(e.operator, left, right)
		den := right.(float64)
		if den == 0 {
			reportRuntimeError(e.operator, "Division by zero")
		}
		return left.(float64) / den
	case tokenTypeStar:
		i.checkNumberOperands(e.operator, left, right)
		return left.(float64) * right.(float64)
	case tokenTypePlus:
		ln, lok := left.(float64)
		rn, rok := right.(float64)
		if lok && rok {
			return ln + rn
		}
		sln, slok := left.(string)
		srn, srok := right.(string)
		if slok && srok {
			return sln + srn
		}
		reportRuntimeError(e.operator, "Operands must be two numbers or two strings.")
	case tokenTypeGreater:
		i.checkNumberOperands(e.operator, left, right)
		return left.(float64) > right.(float64)
	case tokenTypeGreaterEqual:
		i.checkNumberOperands(e.operator, left, right)
		return left.(float64) >= right.(float64)
	case tokenTypeLess:
		i.checkNumberOperands(e.operator, left, right)
		return left.(float64) < right.(float64)
	case tokenTypeLessEqual:
		i.checkNumberOperands(e.operator, left, right)
		return left.(float64) <= right.(float64)
	case tokenTypeBangEqual:
		return left != right
	case tokenTypeEqualEqual:
		return left == right
	}
	return nil
}

func (i *interpreter) visitCallExpr(e exprCall) interface{} {
	callee := i.evaluate(e.callee)

	var args []interface{}
	for _, a := range e.args {
		args = append(args, i.evaluate(a))
	}

	f, ok := callee.(callable)
	if !ok {
		reportRuntimeError(e.paren, "Can only call functions and classes.")
	} else if len(args) != f.arity() {
		reportRuntimeError(e.paren, fmt.Sprintf("Expected %d arguments but got %d.",
			f.arity(), len(args)))
	}

	return f.call(i, args)
}

func (i *interpreter) visitGroupingExpr(e exprGrouping) interface{} {
	return i.evaluate(e.exp)
}

func (i *interpreter) visitLiteralExpr(e exprLiteral) interface{} {
	return e.value
}

func (i *interpreter) visitUnaryExpr(e exprUnary) interface{} {
	r := i.evaluate(e.right)
	switch e.operator.tt {
	case tokenTypeMinus:
		i.checkNumberOperand(e.operator, r)
		return -r.(float64)
	case tokenTypeBang:
		return !i.isTruthy(r)
	}
	return nil
}

func (i *interpreter) visitVariableExpr(e exprVariable) interface{} {
	return i.lookUpVariable(e.name, e)
}

func (i *interpreter) lookUpVariable(name token, e expr) interface{} {
	dist, ok := i.locals[e]
	if ok {
		return i.env.getAt(dist, name)
	} else {
		return i.globals.get(name)
	}
}

func (i *interpreter) visitAssignExpr(e exprAssign) interface{} {
	v := i.evaluate(e.value)
	dist, ok := i.locals[e]
	if ok {
		i.env.assignAt(dist, e.name, v)
	} else {
		i.globals.assign(e.name, v)
	}
	return v
}

func (i *interpreter) visitLogicalExpr(e exprLogical) interface{} {
	l := i.evaluate(e.left)
	switch e.operator.tt {
	case tokenTypeAnd:
		if !i.isTruthy(l) {
			return l
		}
	case tokenTypeOr:
		if i.isTruthy(l) {
			return l
		}
	}
	return i.evaluate(e.right)
}

func (i *interpreter) visitGetExpr(e exprGet) interface{} {
	inst, ok := i.evaluate(e.obj).(loxInstance)
	if !ok {
		reportRuntimeError(e.name, "only instances have properties.")
	}
	return inst.get(e.name)
}

func (i *interpreter) visitSetExpr(e exprSet) interface{} {
	obj, ok := i.evaluate(e.obj).(loxInstance)
	if !ok {
		reportRuntimeError(e.name, "Only instances have fields.")
	}
	v := i.evaluate(e.value)
	obj.set(e.name, v)
	return v
}

func (i *interpreter) visitThisExpr(e exprThis) interface{} {
	return i.lookUpVariable(e.name, e)
}

func (i *interpreter) visitSuperExpr(e exprSuper) interface{} {
	dist := i.locals[e]
	super := i.env.getAt(dist, token{lexeme: "super"}).(loxClass)
	this := i.env.getAt(dist-1, token{lexeme: "this"}).(loxInstance)
	m := super.findMethod(e.method.lexeme)
	if m == nil {
		reportRuntimeError(e.method, fmt.Sprintf("Undefined property '%s'.", e.method.lexeme))
		return nil
	}
	return m.bind(this)
}

func (i *interpreter) resolveLocal(e expr, depth int) {
	i.locals[e] = depth
}

func (i *interpreter) isTruthy(v interface{}) bool {
	if b, ok := v.(bool); ok {
		return b
	} else if v == nil {
		return false
	}
	return true
}

func (i *interpreter) evaluate(e expr) interface{} {
	return e.accept(i)
}

func (i *interpreter) execute(s stmt) {
	s.accept(i)
}

func (i *interpreter) executeBlock(s stmtBlock, env *environment) {
	prev := i.env
	i.env = env
	defer func() {
		i.env = prev
	}()
	for _, s := range s.statements {
		i.execute(s)
	}
	return
}

func (i *interpreter) visitWhileStatement(s stmtWhile) interface{} {
	for i.isTruthy(i.evaluate(s.condition)) {
		i.execute(s.body)
	}
	return nil
}

func (i *interpreter) visitIfStatement(s stmtIf) interface{} {
	if i.isTruthy(i.evaluate(s.condition)) {
		i.execute(s.thenBranch)
	} else if s.elseBranch != nil {
		i.execute(s.elseBranch)
	}
	return nil
}

func (i *interpreter) visitBlockStatement(s stmtBlock) interface{} {
	i.executeBlock(s, newEnvironmentWithParent(i.env))
	return nil
}

func (i *interpreter) visitExpressionStatement(s stmtExpression) interface{} {
	i.evaluate(s.e)
	return nil
}

func (i *interpreter) visitFunctionStatement(s stmtFunction) interface{} {
	i.env.define(s.name.lexeme, loxFunction{declaration: s, closure: i.env})
	return nil
}

func (i *interpreter) visitPrintStatement(s stmtPrint) interface{} {
	e := i.evaluate(s.e)
	fmt.Printf("%v\n", e)
	return nil
}

func (i *interpreter) visitReturnStatement(s stmtReturn) interface{} {
	v := returnValue{value: nil}
	if s.value != nil {
		v.value = i.evaluate(s.value)
	}
	panic(v)
}

func (i *interpreter) visitVarStatement(s stmtVar) interface{} {
	var v interface{}
	if s.initializer != nil {
		v = i.evaluate(s.initializer)
	}

	i.env.define(s.name.lexeme, v)
	return nil
}

func (i *interpreter) visitClassStatement(s stmtClass) interface{} {
	var super loxClass
	if s.superClass != nil {
		var ok bool
		super, ok = i.evaluate(s.superClass).(loxClass)
		if !ok {
			reportRuntimeError(s.superClass.name, "Superclass must be a class.")
		}
	}

	i.env.define(s.name.lexeme, nil)
	if s.superClass != nil {
		i.env = newEnvironmentWithParent(i.env)
		i.env.define("super", super)
	}

	ms := make(map[string]loxFunction, len(s.methods))
	for _, m := range s.methods {
		ms[m.name.lexeme] = loxFunction{
			declaration:   m,
			closure:       i.env,
			isInitializer: m.name.lexeme == "init",
		}
	}

	c := loxClass{name: s.name.lexeme, methods: ms, superClass: &super}
	if s.superClass != nil {
		i.env = i.env.enclosing
	}
	i.env.assign(s.name, c)
	return nil
}

func (i *interpreter) checkNumberOperand(operator token, operand interface{}) {
	if _, ok := operand.(float64); ok {
		return
	}
	reportRuntimeError(operator, "Operand must be a number.")
}

func (i *interpreter) checkNumberOperands(operator token, left, right interface{}) {
	_, lok := left.(float64)
	_, rok := right.(float64)
	if lok && rok {
		return
	}
	reportRuntimeError(operator, "Operands must be numbers.")
}
