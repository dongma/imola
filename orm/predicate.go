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

type Column struct {
	name string
}

func C(name string) Column {
	return Column{
		name: name,
	}
}

// Eq C("id").Eq(12)
func (c Column) Eq(arg any) Predicate {
	return Predicate{
		left: c,
		op:   opEq,
		right: value{
			val: arg,
		},
	}
}

func (c Column) Lt(arg any) Predicate {
	return Predicate{
		left: c,
		op:   opLt,
		right: value{
			val: arg,
		},
	}
}

// Not 用法: Not(C("name").Eq("Tom"))
func Not(p Predicate) Predicate {
	return Predicate{
		op:    opNot,
		right: p,
	}
}

// And 用法: C("id").Eq(12).And(C("name").Eq("Tom"))
func (left Predicate) And(right Predicate) Predicate {
	return Predicate{
		left:  left,
		op:    opAnd,
		right: right,
	}
}

// Or 用法: C("id").Eq(12).Or(C("name").Eq("Tom"))
func (left Predicate) Or(right Predicate) Predicate {
	return Predicate{
		left:  left,
		op:    opOr,
		right: right,
	}
}

// Expression 是一个标记接口，代表一个表达式
type Expression interface {
	expr()
}

func (Predicate) expr() {}

func (Column) expr() {}

type value struct {
	val any
}

func (value) expr() {}
