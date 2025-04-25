package sql

// Expression 是一个标记接口，代表一个表达式
type Expression interface {
	Expr()
}

// RawExpr 代表原生表达式
type RawExpr struct {
	Raw  string
	Args []any
}

func Raw(expr string, args ...any) RawExpr {
	return RawExpr{
		Raw:  expr,
		Args: args,
	}
}

func (r RawExpr) Selectable() {}

func (r RawExpr) Expr() {}

func (r RawExpr) AsPredicate() Predicate {
	return Predicate{
		Left: r,
	}
}
