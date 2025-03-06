package orm

import (
	"context"
	"database/sql"
	"imola/orm/internal/errs"
	"imola/orm/model"
)

type OnDuplicateKeyBuilder[T any] struct {
	i               *Inserter[T]
	conflictColumns []string
}

type OnDuplicateKey struct {
	assigns         []Assignable
	conflictColumns []string
}

func (o *OnDuplicateKeyBuilder[T]) ConflictColumns(cols ...string) *OnDuplicateKeyBuilder[T] {
	o.conflictColumns = cols
	return o
}

func (o *OnDuplicateKeyBuilder[T]) Update(assigns ...Assignable) *Inserter[T] {
	o.i.onDuplicateKey = &OnDuplicateKey{
		assigns:         assigns,
		conflictColumns: o.conflictColumns,
	}
	return o.i
}

type Assignable interface {
	assign()
}

type Inserter[T any] struct {
	builder
	sess           Session
	values         []*T
	columns        []string
	onDuplicateKey *OnDuplicateKey
}

func NewInserter[T any](sess Session) *Inserter[T] {
	core := sess.getCore()
	return &Inserter[T]{
		builder: builder{
			core:   core,
			quoter: core.dialect.quoter(),
		},
		sess: sess,
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
	if i.model == nil {
		mdl, err := i.r.Get(i.values[0])
		i.model = mdl
		if err != nil {
			return nil, err
		}
	}
	// 1、拼接表名
	i.quote(i.model.TableName)

	// 2、显示的指定列的顺序, 不能使用 FieldMap因为在Go中其顺序是乱序
	fields := i.model.Fields
	if len(i.columns) > 0 {
		fields = make([]*model.Field, 0, len(i.columns))
		for _, fd := range i.columns {
			fdMeta, ok := i.model.FieldMap[fd]
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
	i.args = make([]any, 0, len(i.model.Fields))

	for j, val := range i.values {
		if j > 0 {
			i.sb.WriteByte(',')
		}
		i.sb.WriteByte('(')
		unsafe := i.creator(i.model, val)
		for idx, field := range fields {
			if idx > 0 {
				i.sb.WriteByte(',')
			}
			i.sb.WriteByte('?')
			// 读出来参数, 使用反射实现 (改成了unsafe的方式)
			arg, err := unsafe.Field(field.GoName)
			if err != nil {
				return nil, err
			}
			i.AddArg(arg)
		}
		i.sb.WriteByte(')')
	}

	if i.onDuplicateKey != nil {
		err := i.dialect.buildOnDuplicateKey(&i.builder, i.onDuplicateKey)
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

func (i *Inserter[T]) Exec(ctx context.Context) Result {
	var err error
	i.model, err = i.r.Get(new(T))
	if err != nil {
		return Result{
			Err: err,
		}
	}
	res := exec(ctx, i.sess, i.core, &QueryContext{
		Type:    "INSERT",
		Builder: i,
		Model:   i.model,
	})
	var sqlRes sql.Result
	if res.Result != nil {
		sqlRes = res.Result.(sql.Result)
	}
	return Result{
		Err: res.Err,
		Res: sqlRes,
	}
}

var _ Handler = (&Inserter[int]{}).execHandler

func (i *Inserter[T]) execHandler(ctx context.Context, qc *QueryContext) *QueryResult {
	query, err := i.Build()
	if err != nil {
		return &QueryResult{
			Err: err,
			Result: Result{
				Err: err,
			},
		}
	}
	res, err := i.sess.execContext(ctx, query.SQL, query.Args...)
	return &QueryResult{
		Err: err,
		Result: Result{
			Err: err,
			Res: res,
		},
	}
}
