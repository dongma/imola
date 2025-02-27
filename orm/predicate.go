package orm

// op op为string的衍生类型，用于定义操作类型
type op string

func (o op) String() string {
	return string(o)
}

const (
	opEq  op = "="
	opNot op = "NOT"
	opAnd op = "AND"
	opOr  op = "OR"
	opLt  op = "<"
)

type Predicate struct {
	left  Expression
	op    op
	right Expression
}

func (Predicate) expr() {}

// Expression 是一个标记接口，代表一个表达式
type Expression interface {
	expr()
}

type value struct {
	val any
}

func (value) expr() {}
