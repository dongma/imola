package valuer

import (
	"database/sql"
	"github.com/dongma/imola/orm/internal/errs"
	"github.com/dongma/imola/orm/model"
	"reflect"
	"unsafe"
)

type unsafeValue struct {
	model *model.Model
	// 起始地址
	address unsafe.Pointer
}

var _ Creator = NewUnsafeValue

func NewUnsafeValue(model *model.Model, val any) Value {
	// 实体对象的起始地址 address
	address := reflect.ValueOf(val).UnsafePointer()
	return &unsafeValue{
		model:   model,
		address: address,
	}
}

func (u unsafeValue) SetColumn(rows *sql.Rows) error {
	// 拿到select出来的列
	cols, err := rows.Columns()
	if err != nil {
		return err
	}

	var vals []any
	for _, col := range cols {
		field, ok := u.model.ColumnMap[col]
		if !ok {
			return errs.NewErrUnknownColumn(col)
		}
		// 计算字段的地址：起始地址 + 偏移量
		fdAddress := unsafe.Pointer(uintptr(u.address) + field.Offset)
		val := reflect.NewAt(field.Typ, fdAddress)
		vals = append(vals, val.Interface())
	}
	err = rows.Scan(vals...)
	return err
}

func (u unsafeValue) Field(name string) (any, error) {
	fd, ok := u.model.FieldMap[name]
	if !ok {
		return nil, errs.NewErrUnknownField(name)
	}
	fdAddress := unsafe.Pointer(uintptr(u.address) + fd.Offset)
	val := reflect.NewAt(fd.Typ, fdAddress)
	return val.Elem().Interface(), nil
}
