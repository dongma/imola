package orm

import (
	"imola/orm/sql"
)

type TableReference interface {
	Table()
}

// Table 普通表
type Table struct {
	Alias  string
	Entity any
}

func (t Table) Join(right TableReference) *JoinBuilder {
	return &JoinBuilder{
		Left:  t,
		Right: right,
		Typ:   "JOIN",
	}
}

func (t Table) LeftJoin(right TableReference) *JoinBuilder {
	return &JoinBuilder{
		Left:  t,
		Right: right,
		Typ:   "LEFT JOIN",
	}
}

func (t Table) RightJoin(right TableReference) *JoinBuilder {
	return &JoinBuilder{
		Left:  t,
		Right: right,
		Typ:   "RIGHT JOIN",
	}
}

func TableOf(entity any) Table {
	return Table{
		Entity: entity,
	}
}

func (t Table) Table() {
	panic("implement me")
}

func (t Table) As(alias string) Table {
	return Table{
		Entity: t.Entity,
		Alias:  alias,
	}
}

func (t Table) C(name string) Column {
	return Column{
		Name:  name,
		Table: t,
	}
}

// JoinBuilder join语句的构建
type JoinBuilder struct {
	Left  TableReference
	Right TableReference
	Typ   string
}

func (j *JoinBuilder) On(ps ...sql.Predicate) Join {
	return Join{
		Left:  j.Left,
		Right: j.Right,
		Typ:   j.Typ,
		On:    ps,
	}
}

func (j *JoinBuilder) Using(cols ...string) Join {
	return Join{
		Left:  j.Left,
		Right: j.Right,
		Typ:   j.Typ,
		Using: cols,
	}
}

// Join join实体本身，支持predicate和on字段
type Join struct {
	Left  TableReference
	Right TableReference
	Typ   string
	On    []sql.Predicate
	Using []string
}

func (j Join) Table() {
	panic("implement me")
}

// Join Join类型本身也支持join
func (j *Join) Join(right TableReference) *JoinBuilder {
	return &JoinBuilder{
		Left:  j,
		Right: right,
		Typ:   "JOIN",
	}
}

func (j *Join) LeftJoin(right TableReference) *JoinBuilder {
	return &JoinBuilder{
		Left:  j,
		Right: right,
		Typ:   "LEFT JOIN",
	}
}

func (j *Join) RightJoin(right TableReference) *JoinBuilder {
	return &JoinBuilder{
		Left:  j,
		Right: right,
		Typ:   "RIGHT JOIN",
	}
}
