package valuer

import (
	"database/sql"
	"github.com/dongma/imola/orm/internal/errs"
	"github.com/dongma/imola/orm/model"
	"reflect"
)

type reflectValue struct {
	model *model.Model
	// 对应于T的指针
	val reflect.Value
}

// 确保Value签名发生变化后，这边可以得到通知，会发生compile error
var _ Creator = NewReflectValue

func NewReflectValue(model *model.Model, val any) Value {
	return &reflectValue{
		model: model,
		val:   reflect.ValueOf(val).Elem(),
	}
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
		// first_name->FirstName，拿到表字段对应go struct字段
		field, ok := r.model.ColumnMap[col]
		if !ok {
			return errs.NewErrUnknownColumn(col)
		}
		// 用反射创建一个实例 (原本类型的指针类型), 例如: fd.type = int, 那么val是*int
		val := reflect.New(field.Typ)
		vals = append(vals, val.Interface())
		// 因为val是*int，那么val.Elem()的结果就是int
		valElems = append(valElems, val.Elem())
	}

	// select id, first_name, age, last_name，根据Scan的用法，其参数都是指针类型。 Scan之后，就会将sql结果写入
	err = rows.Scan(vals...)
	if err != nil {
		return err
	}

	// 将vals值塞到tp里面
	tpValueElem := r.val
	for i, col := range cols {
		field, ok := r.model.ColumnMap[col]
		if !ok {
			return errs.NewErrUnknownColumn(col)
		}
		tpValueElem.FieldByName(field.GoName).Set(valElems[i])
	}
	return err
}

func (r reflectValue) Field(name string) (any, error) {
	return r.val.FieldByName(name).Interface(), nil
}
