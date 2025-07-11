package orm

import (
	"github.com/dongma/imola/orm/sql"
)

type Column struct {
	Table TableReference
	Name  string
	Alias string
}

func (c Column) Assign() {
}

func C(name string) Column {
	return Column{
		Name: name,
	}
}

// As 场景select field_a as a
func (c Column) As(alias string) Column {
	return Column{
		Name:  c.Name,
		Alias: alias,
		Table: c.Table,
	}
}

// Eq C("id").Eq(12)
func (c Column) Eq(arg any) sql.Predicate {
	return sql.Predicate{
		Left:  c,
		Op:    sql.OpEq,
		Right: ValueOf(arg),
	}
}

func ValueOf(arg any) sql.Expression {
	switch val := arg.(type) {
	case sql.Expression:
		return val
	default:
		return sql.Value{Val: val}
	}
}

func (c Column) Lt(arg any) sql.Predicate {
	return sql.Predicate{
		Left: c,
		Op:   sql.OpLt,
		Right: sql.Value{
			Val: arg,
		},
	}
}

// Not 用法: Not(C("Name").Eq("Tom"))
func Not(p sql.Predicate) sql.Predicate {
	return sql.Predicate{
		Op:    sql.OpNot,
		Right: p,
	}
}

func (Column) Expr() {}

func (Column) Selectable() {}
