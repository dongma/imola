package orm

type Column struct {
	name  string
	alias string
}

func (c Column) assign() {
}

func C(name string) Column {
	return Column{
		name: name,
	}
}

func (c Column) As(alias string) Column {
	return Column{
		name:  c.name,
		alias: alias,
	}
}

// Eq C("id").Eq(12)
func (c Column) Eq(arg any) Predicate {
	return Predicate{
		left:  c,
		op:    opEq,
		right: valueOf(arg),
	}
}

func valueOf(arg any) Expression {
	switch val := arg.(type) {
	case Expression:
		return val
	default:
		return value{val: val}
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

func (Column) expr() {}

func (Column) selectable() {}
