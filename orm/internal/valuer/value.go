package valuer

import (
	"database/sql"
	"github.com/dongma/imola/orm/model"
)

type Value interface {
	Field(name string) (any, error)
	SetColumn(rows *sql.Rows) error
}

type Creator func(model *model.Model, entity any) Value
