package main

type stmt interface {
	accept(v stmtVisitor) interface{}
}

type stmtVisitor interface {
	visitExpressionStatement(s stmtExpression) interface{}
	visitPrintStatement(s stmtPrint) interface{}
	visitVarStatement(s stmtVar) interface{}
	visitBlockStatement(s stmtBlock) interface{}
	visitIfStatement(s stmtIf) interface{}
	visitWhileStatement(s stmtWhile) interface{}
	visitFunctionStatement(s stmtFunction) interface{}
	visitReturnStatement(s stmtReturn) interface{}
	visitClassStatement(s stmtClass) interface{}
}

type stmtExpression struct {
	e expr
}

func (s stmtExpression) accept(v stmtVisitor) interface{} {
	return v.visitExpressionStatement(s)
}

type stmtPrint struct {
	e expr
}

var _ stmt = stmtPrint{}

func (s stmtPrint) accept(v stmtVisitor) interface{} {
	return v.visitPrintStatement(s)
}

type stmtVar struct {
	name        token
	initializer expr
}

func (s stmtVar) accept(v stmtVisitor) interface{} {
	return v.visitVarStatement(s)
}

type stmtBlock struct {
	statements []stmt
}

func (s stmtBlock) accept(v stmtVisitor) interface{} {
	return v.visitBlockStatement(s)
}

type stmtIf struct {
	condition              expr
	thenBranch, elseBranch stmt
}

func (s stmtIf) accept(v stmtVisitor) interface{} {
	return v.visitIfStatement(s)
}

type stmtWhile struct {
	condition expr
	body      stmt
}

func (s stmtWhile) accept(v stmtVisitor) interface{} {
	return v.visitWhileStatement(s)
}

type stmtFunction struct {
	params []token
	body   stmtBlock
	name   token
}

func (s stmtFunction) accept(v stmtVisitor) interface{} {
	return v.visitFunctionStatement(s)
}

type stmtReturn struct {
	keyword token
	value   expr
}

func (s stmtReturn) accept(v stmtVisitor) interface{} {
	return v.visitReturnStatement(s)
}

type stmtClass struct {
	methods    []stmtFunction
	name       token
	superClass *exprVariable
}

func (s stmtClass) accept(v stmtVisitor) interface{} {
	return v.visitClassStatement(s)
}
