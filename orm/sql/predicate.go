package sql

// Op op为string的衍生类型，用于定义操作类型
type Op string

func (o Op) String() string {
	return string(o)
}

const (
	OpEq  Op = "="
	OpNot Op = "NOT"
	OpAnd Op = "AND"
	OpOr  Op = "OR"
	OpLt  Op = "<"
	OpRt  Op = ">"
)

type Predicate struct {
	Left  Expression
	Op    Op
	Right Expression
}

func (Predicate) Expr() {}

type Value struct {
	Val any
}

func (Value) Expr() {}

// And 用法: C("id").Eq(12).And(C("Name").Eq("Tom"))
func (left Predicate) And(right Predicate) Predicate {
	return Predicate{
		Left:  left,
		Op:    OpAnd,
		Right: right,
	}
}

// Or 用法: C("id").Eq(12).Or(C("Name").Eq("Tom"))
func (left Predicate) Or(right Predicate) Predicate {
	return Predicate{
		Left:  left,
		Op:    OpOr,
		Right: right,
	}
}
