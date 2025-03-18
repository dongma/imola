package sql

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
