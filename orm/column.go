package orm

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

func (Column) expr() {}

func (Column) selectable() {}
