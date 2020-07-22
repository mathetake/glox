package main

import (
	"fmt"
	"runtime/debug"
)

type resolver struct {
	inter               *interpreter
	scopes              []map[string]bool
	currentFunctionType functionType
	currentClass        classType
}

type functionType int

const (
	functionTypeNone functionType = iota
	functionTypeFunction
	functionTypeInitializer
	functionTypeMethod
)

type classType int

const (
	classTypeNone classType = iota
	classTypeClass
	classTypeSubclass
)

func (r *resolver) resolve(ss []stmt) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
			fmt.Println(string(debug.Stack()))
		}
	}()
	r.resolveStatements(ss)
}

func (r *resolver) pushScope(scope map[string]bool) {
	r.scopes = append(r.scopes, scope)
}

func (r *resolver) popScope() {
	r.scopes = r.scopes[:len(r.scopes)-1]
}

func (r *resolver) peekScope() map[string]bool {
	return r.scopes[len(r.scopes)-1]
}

func (r *resolver) isScopeEmpty() bool {
	return len(r.scopes) == 0
}

var (
	_ exprVisitor = &resolver{}
	_ stmtVisitor = &resolver{}
)

func (r *resolver) visitBinaryExpr(e exprBinary) interface{} {
	r.resolveExpression(e.left)
	r.resolveExpression(e.right)
	return nil
}

func (r *resolver) visitGroupingExpr(e exprGrouping) interface{} {
	r.resolveExpression(e.exp)
	return nil
}

func (r *resolver) visitLiteralExpr(exprLiteral) interface{} { return nil }

func (r *resolver) visitUnaryExpr(e exprUnary) interface{} {
	r.resolveExpression(e.right)
	return nil
}

func (r *resolver) visitLogicalExpr(e exprLogical) interface{} {
	r.resolveExpression(e.left)
	r.resolveExpression(e.right)
	return nil
}

func (r *resolver) visitCallExpr(e exprCall) interface{} {
	r.resolveExpression(e.callee)
	for _, a := range e.args {
		r.resolveExpression(a)
	}
	return nil
}

func (r *resolver) visitExpressionStatement(s stmtExpression) interface{} {
	r.resolveExpression(s.e)
	return nil
}

func (r *resolver) visitPrintStatement(s stmtPrint) interface{} {
	r.resolveExpression(s.e)
	return nil
}

func (r *resolver) visitIfStatement(s stmtIf) interface{} {
	r.resolveExpression(s.condition)
	r.resolveStatement(s.thenBranch)
	if s.elseBranch != nil {
		r.resolveStatement(s.elseBranch)
	}
	return nil
}

func (r *resolver) visitWhileStatement(s stmtWhile) interface{} {
	r.resolveStatement(s.body)
	r.resolveExpression(s.condition)
	return nil
}

func (r *resolver) visitReturnStatement(s stmtReturn) interface{} {
	if r.currentFunctionType == functionTypeNone {
		reportResolutionError(s.keyword, "Cannot return from top-level code.")
	}
	if s.value != nil {
		if r.currentFunctionType == functionTypeInitializer {
			reportResolutionError(s.keyword, "Cannot return from an initializer.")
		}
		r.resolveExpression(s.value)
	}
	return nil
}

func (r *resolver) visitClassStatement(s stmtClass) interface{} {
	ec := r.currentClass
	r.currentClass = classTypeClass
	r.declare(s.name)
	r.define(s.name)
	if s.superClass != nil && s.superClass.name.lexeme == s.name.lexeme {
		reportResolutionError(s.superClass.name, "A class cannot inherit from itself.")
	}
	if s.superClass != nil {
		r.currentClass = classTypeSubclass
		r.resolveExpression(s.superClass)
		r.beginScope()
		r.peekScope()["super"] = true
	}

	r.beginScope()
	r.peekScope()["this"] = true
	for _, m := range s.methods {
		var d = functionTypeMethod
		if m.name.lexeme == "init" {
			d = functionTypeInitializer
		}
		r.resolveFunctionStmt(m, d)
	}
	r.endScope()
	if s.superClass != nil {
		r.endScope()
	}

	r.currentClass = ec
	return nil
}

func (r *resolver) visitFunctionStatement(s stmtFunction) interface{} {
	r.declare(s.name)
	r.define(s.name)
	r.resolveFunctionStmt(s, functionTypeFunction)
	return nil
}

func (r *resolver) visitBlockStatement(s stmtBlock) interface{} {
	r.beginScope()
	r.resolveStatements(s.statements)
	r.endScope()
	return nil
}

func (r *resolver) visitVarStatement(s stmtVar) interface{} {
	r.declare(s.name)
	if s.initializer != nil {
		r.resolveExpression(s.initializer)
	}
	r.define(s.name)
	return nil
}

func (r *resolver) visitAssignExpr(e exprAssign) interface{} {
	r.resolveExpression(e.value)
	r.resolveLocal(e, e.name)
	return nil
}

func (r *resolver) visitGetExpr(e exprGet) interface{} {
	r.resolveExpression(e.obj)
	return nil
}
func (r *resolver) visitSetExpr(e exprSet) interface{} {
	r.resolveExpression(e.obj)
	r.resolveExpression(e.value)
	return nil
}

func (r *resolver) visitThisExpr(e exprThis) interface{} {
	if r.currentClass == classTypeNone {
		reportResolutionError(e.name, "Cannot use 'this' outside of a class.")
	}
	r.resolveLocal(e, e.name)
	return nil
}

func (r *resolver) visitSuperExpr(e exprSuper) interface{} {
	switch r.currentClass {
	case classTypeNone:
		reportResolutionError(e.keyword, "Cannot use 'super' outside of a class.")
	case classTypeClass:
		reportResolutionError(e.keyword, "Cannot use 'super' in a class with no superclass.")
	}
	r.resolveLocal(e, e.keyword)
	return nil
}

func (r *resolver) visitVariableExpr(e exprVariable) interface{} {
	if !r.isScopeEmpty() {
		defined, ok := r.peekScope()[e.name.lexeme]
		if ok && !defined {
			reportResolutionError(e.name, "Cannot read local variable in its own initializer.")
		}
	}

	r.resolveLocal(e, e.name)
	return nil
}

func (r *resolver) resolveFunctionStmt(s stmtFunction, t functionType) {
	enclosing := r.currentFunctionType
	r.currentFunctionType = t
	r.beginScope()
	for _, p := range s.params {
		r.declare(p)
		r.define(p)
	}

	r.resolveStatements(s.body.statements)
	r.endScope()
	r.currentFunctionType = enclosing
}

func (r *resolver) resolveStatements(statements []stmt) {
	for _, s := range statements {
		r.resolveStatement(s)
	}
}
func (r *resolver) resolveStatement(s stmt) {
	s.accept(r)
}

func (r *resolver) resolveExpression(e expr) {
	e.accept(r)
}

func (r *resolver) resolveLocal(e expr, name token) {
	for i := len(r.scopes) - 1; i >= 0; i-- {
		if defined := r.scopes[i][name.lexeme]; defined {
			r.inter.resolveLocal(e, len(r.scopes)-1-i)
			return
		}
	}
}

func (r *resolver) beginScope() {
	r.pushScope(map[string]bool{})
}

func (r *resolver) endScope() {
	r.popScope()
}

func (r *resolver) declare(name token) {
	if r.isScopeEmpty() {
		return
	}
	p := r.peekScope()
	if _, ok := p[name.lexeme]; ok {
		reportResolutionError(name, "Variable with this name already declared in this scope.")
	} else {
		p[name.lexeme] = false
	}
}

func (r *resolver) define(name token) {
	if r.isScopeEmpty() {
		return
	}
	r.peekScope()[name.lexeme] = true
}
