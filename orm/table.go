package orm

type TableReference interface {
	table()
}

// Table 普通表
type Table struct {
	alias  string
	entity any
}

func (t Table) Join(right TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  t,
		right: right,
		typ:   "JOIN",
	}
}

func (t Table) LeftJoin(right TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  t,
		right: right,
		typ:   "LEFT JOIN",
	}
}

func (t Table) RightJoin(right TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  t,
		right: right,
		typ:   "RIGHT JOIN",
	}
}

func TableOf(entity any) Table {
	return Table{
		entity: entity,
	}
}

func (t Table) table() {
	panic("implement me")
}

func (t Table) As(alias string) Table {
	return Table{
		entity: t.entity,
		alias:  alias,
	}
}

func (t Table) C(name string) Column {
	return Column{
		name:  name,
		table: t,
	}
}

// JoinBuilder join语句的构建
type JoinBuilder struct {
	left  TableReference
	right TableReference
	typ   string
}

func (j *JoinBuilder) On(ps ...Predicate) Join {
	return Join{
		left:  j.left,
		right: j.right,
		typ:   j.typ,
		on:    ps,
	}
}

func (j *JoinBuilder) Using(cols ...string) Join {
	return Join{
		left:  j.left,
		right: j.right,
		typ:   j.typ,
		using: cols,
	}
}

// Join join实体本身，支持predicate和on字段
type Join struct {
	left  TableReference
	right TableReference
	typ   string
	on    []Predicate
	using []string
}

func (j Join) table() {
	panic("implement me")
}

// Join Join类型本身也支持join
func (j *Join) Join(right TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  j,
		right: right,
		typ:   "JOIN",
	}
}

func (j *Join) LeftJoin(right TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  j,
		right: right,
		typ:   "LEFT JOIN",
	}
}

func (j *Join) RightJoin(right TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  j,
		right: right,
		typ:   "RIGHT JOIN",
	}
}
