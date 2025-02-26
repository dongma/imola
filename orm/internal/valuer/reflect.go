package valuer

import (
	"database/sql"
	"imola/orm/internal/errs"
	"imola/orm/model"
	"reflect"
)

type reflectValue struct {
	model *model.Model
	// 对应于T的指针
	val any
}

// 确保Value签名发生变化后，这边可以得到通知，会发生compile error
var _ Creator = NewReflectValue

func NewReflectValue(model *model.Model, val any) Value {
	return &reflectValue{model: model, val: val}
}

func (r reflectValue) SetColumn(rows *sql.Rows) error {
	// 拿到select出来的列
	cols, err := rows.Columns()
	if err != nil {
		return err
	}
	// 如何利用cols来解决顺序和类型问题
	vals := make([]any, 0, len(cols))
	valElems := make([]reflect.Value, 0, len(cols))
	for _, col := range cols {
		field, ok := r.model.ColumnMap[col]
		if !ok {
			return errs.NewErrUnknownColumn(col)
		}
		// 用反射创建一个实例 (原本类型的指针类型), 例如: fd.type = int, 那么val是*int
		val := reflect.New(field.Typ)
		vals = append(vals, val.Interface())
		// 此处要调用val.Elem()，因为fd.type = int, 那么val是*int
		valElems = append(valElems, val.Elem())
	}

	// select id, first_name, age, last_name
	err = rows.Scan(vals...)
	if err != nil {
		return err
	}

	// 将vals值塞到tp里面
	tpValueElem := reflect.ValueOf(r.val).Elem()
	for i, col := range cols {
		field, ok := r.model.ColumnMap[col]
		if !ok {
			return errs.NewErrUnknownColumn(col)
		}
		tpValueElem.FieldByName(field.GoName).Set(valElems[i])
	}
	return err
}
