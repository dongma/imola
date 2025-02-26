package valuer

import (
	"database/sql"
	"imola/orm/internal/errs"
	"imola/orm/model"
	"reflect"
	"unsafe"
)

type unsafeValue struct {
	model *model.Model
	val   any
}

var _ Creator = NewUnsafeValue

func NewUnsafeValue(model *model.Model, val any) Value {
	return &unsafeValue{model: model, val: val}
}

func (u unsafeValue) SetColumn(rows *sql.Rows) error {
	// 拿到select出来的列
	cols, err := rows.Columns()
	if err != nil {
		return err
	}

	var vals []any
	// 实体对象的起始地址 address
	address := reflect.ValueOf(u.val).UnsafePointer()
	for _, col := range cols {
		field, ok := u.model.ColumnMap[col]
		if !ok {
			return errs.NewErrUnknownColumn(col)
		}
		// 计算字段的地址：起始地址 + 偏移量
		fdAddress := unsafe.Pointer(uintptr(address) + field.Offset)
		val := reflect.NewAt(field.Typ, fdAddress)
		vals = append(vals, val.Interface())
	}
	err = rows.Scan(vals...)
	return err
}
