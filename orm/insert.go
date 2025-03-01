package orm

import (
	"imola/orm/internal/errs"
	"imola/orm/model"
	"reflect"
)

type OnDuplicateKeyBuilder[T any] struct {
	i *Inserter[T]
}

type OnDuplicateKey struct {
	assigns []Assignable
}

func (o *OnDuplicateKeyBuilder[T]) Update(assigns ...Assignable) *Inserter[T] {
	o.i.onDuplicateKey = &OnDuplicateKey{
		assigns: assigns,
	}
	return o.i
}

type Assignable interface {
	assign()
}

type Inserter[T any] struct {
	builder
	values         []*T
	db             *DB
	columns        []string
	onDuplicateKey *OnDuplicateKey
}

func NewInserter[T any](db *DB) *Inserter[T] {
	return &Inserter[T]{
		builder: builder{
			dialect: db.dialect,
			quoter:  db.dialect.quoter(),
		},
		db: db,
	}
}

func (i *Inserter[T]) OnDuplicateKey() *OnDuplicateKeyBuilder[T] {
	return &OnDuplicateKeyBuilder[T]{
		i: i,
	}
}

// Values 指定插入的数据
func (i *Inserter[T]) Values(vals ...*T) *Inserter[T] {
	i.values = vals
	return i
}

// Columns 将指定的列进行插入
func (i *Inserter[T]) Columns(cols ...string) *Inserter[T] {
	i.columns = cols
	return i
}

// Build 构建插入的insert语句
func (i *Inserter[T]) Build() (*Query, error) {
	if len(i.values) == 0 {
		return nil, errs.ErrInsertZeroRow
	}
	i.sb.WriteString("INSERT INTO ")
	mdl, err := i.db.r.Get(i.values[0])
	i.model = mdl
	if err != nil {
		return nil, err
	}

	// 1、拼接表名
	i.quote(mdl.TableName)

	// 2、显示的指定列的顺序, 不能使用 FieldMap因为在Go中其顺序是乱序
	fields := mdl.Fields
	if len(i.columns) > 0 {
		fields = make([]*model.Field, 0, len(i.columns))
		for _, fd := range i.columns {
			fdMeta, ok := mdl.FieldMap[fd]
			if !ok {
				return nil, errs.NewErrUnknownField(fd)
			}
			fields = append(fields, fdMeta)
		}
	}

	i.sb.WriteByte('(')
	for idx, field := range fields {
		if idx > 0 {
			i.sb.WriteByte(',')
		}
		i.quote(field.Column)
	}
	i.sb.WriteByte(')')

	// 3、拼接Values的部分, 以及构建Args
	i.sb.WriteString(" VALUES ")
	i.args = make([]any, 0, len(mdl.Fields))

	for j, val := range i.values {
		if j > 0 {
			i.sb.WriteByte(',')
		}
		i.sb.WriteByte('(')
		for idx, field := range fields {
			if idx > 0 {
				i.sb.WriteByte(',')
			}
			i.sb.WriteByte('?')
			// 读出来参数, 使用反射实现
			arg := reflect.ValueOf(val).Elem().FieldByName(field.GoName).Interface()
			i.AddArg(arg)
		}
		i.sb.WriteByte(')')
	}

	if i.onDuplicateKey != nil {
		err = i.dialect.buildOnDuplicateKey(&i.builder, i.onDuplicateKey)
		if err != nil {
			return nil, err
		}
	}

	i.sb.WriteByte(';')
	return &Query{
		SQL:  i.sb.String(),
		Args: i.args,
	}, nil
}
