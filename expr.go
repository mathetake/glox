package main

type expr interface {
	accept(v exprVisitor) interface{}
}

type exprVisitor interface {
	visitBinaryExpr(e exprBinary) interface{}
	visitGroupingExpr(e exprGrouping) interface{}
	visitLiteralExpr(e exprLiteral) interface{}
	visitUnaryExpr(e exprUnary) interface{}
	visitVariableExpr(e exprVariable) interface{}
	visitAssignExpr(e exprAssign) interface{}
	visitLogicalExpr(e exprLogical) interface{}
	visitCallExpr(e exprCall) interface{}
	visitGetExpr(e exprGet) interface{}
	visitSetExpr(e exprSet) interface{}
	visitThisExpr(e exprThis) interface{}
	visitSuperExpr(e exprSuper) interface{}
}

type exprBinary struct {
	left, right expr
	operator    token
}

func (e exprBinary) accept(v exprVisitor) interface{} {
	return v.visitBinaryExpr(e)
}

type exprGrouping struct {
	exp expr
}

func (e exprGrouping) accept(v exprVisitor) interface{} {
	return v.visitGroupingExpr(e)
}

type exprLiteral struct {
	value interface{}
}

func (e exprLiteral) accept(v exprVisitor) interface{} {
	return v.visitLiteralExpr(e)
}

type exprUnary struct {
	operator token
	right    expr
}

func (e exprUnary) accept(v exprVisitor) interface{} {
	return v.visitUnaryExpr(e)
}

type exprVariable struct {
	name token
}

func (e exprVariable) accept(v exprVisitor) interface{} {
	return v.visitVariableExpr(e)
}

type exprAssign struct {
	name  token
	value expr
}

func (e exprAssign) accept(v exprVisitor) interface{} {
	return v.visitAssignExpr(e)
}

type exprLogical struct {
	left, right expr
	operator    token
}

func (e exprLogical) accept(v exprVisitor) interface{} {
	return v.visitLogicalExpr(e)
}

type exprCall struct {
	paren  token
	args   []expr
	callee expr
}

func (e exprCall) accept(v exprVisitor) interface{} {
	return v.visitCallExpr(e)
}

type exprGet struct {
	name token
	obj  expr
}

func (e exprGet) accept(v exprVisitor) interface{} {
	return v.visitGetExpr(e)
}

type exprSet struct {
	name       token
	obj, value expr
}

func (e exprSet) accept(v exprVisitor) interface{} {
	return v.visitSetExpr(e)
}

type exprThis struct {
	name token
}

func (e exprThis) accept(v exprVisitor) interface{} {
	return v.visitThisExpr(e)
}

type exprSuper struct {
	keyword, method token
}

func (e exprSuper) accept(v exprVisitor) interface{} {
	return v.visitSuperExpr(e)
}
