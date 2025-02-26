package valuer

import (
	"database/sql"
	"imola/orm/model"
)

type Value interface {
	SetColumn(rows *sql.Rows) error
}

type Creator func(model *model.Model, entity any) Value
