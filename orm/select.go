package orm

import (
	"context"
	"imola/orm/internal/errs"
	"imola/orm/model"
	"strings"
)

// Selectable 是一个标记接口，代表查找的列或聚合函数等，例如：select xx 这部分
type Selectable interface {
	selectable()
}

type Selector[T any] struct {
	table    string
	model    *model.Model
	where    []Predicate
	sBuilder *strings.Builder
	args     []any
	db       *DB
	columns  []Selectable
}

func NewSelector[T any](db *DB) *Selector[T] {
	return &Selector[T]{
		sBuilder: &strings.Builder{},
		db:       db,
	}
}

func (s Selector[T]) Build() (*Query, error) {
	s.sBuilder = &strings.Builder{}
	var err error
	s.model, err = s.db.r.Get(new(T))
	if err != nil {
		return nil, err
	}

	sBuilder := s.sBuilder
	sBuilder.WriteString("SELECT ")
	err = s.buildColumns()
	if err != nil {
		return nil, err
	}

	if s.table == "" {
		sBuilder.WriteByte('`')
		sBuilder.WriteString(s.model.TableName)
		sBuilder.WriteByte('`')
	} else {
		sBuilder.WriteString(s.table)
	}

	//args := make([]any, 0, 4)
	if len(s.where) > 0 {
		sBuilder.WriteString(" WHERE ")
		predicate := s.where[0]
		for i := 1; i < len(s.where); i++ {
			predicate = predicate.And(s.where[i])
		}
		if err := s.BuildExpression(predicate); err != nil {
			return nil, err
		}
	}

	sBuilder.WriteByte(';')
	return &Query{
		SQL:  sBuilder.String(),
		Args: s.args,
	}, nil
}

// BuildExpression 在这里处理predicate, 构建p.left、p.op和p.right
func (s *Selector[T]) BuildExpression(expr Expression) error {
	switch exp := expr.(type) {
	case nil:
		return nil
	case Predicate:
		_, ok := exp.left.(Predicate)
		if ok {
			s.sBuilder.WriteByte('(')
		}
		if err := s.BuildExpression(exp.left); err != nil {
			return err
		}
		if ok {
			s.sBuilder.WriteByte(')')
		}
		s.sBuilder.WriteByte(' ')
		s.sBuilder.WriteString(exp.op.String())
		s.sBuilder.WriteByte(' ')

		_, ok = exp.right.(Predicate)
		if ok {
			s.sBuilder.WriteByte('(')
		}
		if err := s.BuildExpression(exp.right); err != nil {
			return err
		}
		if ok {
			s.sBuilder.WriteByte(')')
		}
	case Column:
		return s.buildColumn(exp.name)
	case value:
		s.sBuilder.WriteByte('?')
		s.AddArg(exp.val)
	default:
		return errs.NewErrUnsupportedExpression(expr)
	}
	return nil
}

// buildColumns 为了避免方法膨胀，将构建sql列单独抽成一个方法
func (s *Selector[T]) buildColumns() error {
	if len(s.columns) == 0 {
		s.sBuilder.WriteString("*")
		return nil
	}
	for i, col := range s.columns {
		if i > 0 {
			s.sBuilder.WriteByte(',')
		}
		switch c := col.(type) {
		case Column:
			err := s.buildColumn(c.name)
			if err != nil {
				return err
			}
		case Aggregate:
			s.sBuilder.WriteString(c.fn)
			s.sBuilder.WriteByte('(')
			err := s.buildColumn(c.arg)
			if err != nil {
				return err
			}
			s.sBuilder.WriteByte(')')
		}
	}
	return nil
}

func (s *Selector[T]) buildColumn(col string) error {
	fd, ok := s.model.FieldMap[col]
	// 字段不对，或者说列不对
	if !ok {
		return errs.NewErrUnknownField(col)
	}
	s.sBuilder.WriteByte('`')
	s.sBuilder.WriteString(fd.Column)
	s.sBuilder.WriteByte('`')
	return nil
}

func (s *Selector[T]) AddArg(val any) {
	if s.args == nil {
		s.args = make([]any, 0, 8)
	}
	s.args = append(s.args, val)
}

func (s *Selector[T]) Where(conds ...Predicate) *Selector[T] {
	s.where = conds
	return s
}

func (s *Selector[T]) From(tbl string) *Selector[T] {
	s.table = tbl
	return s
}

func (s *Selector[T]) Get(ctx context.Context) (*T, error) {
	query, err := s.Build()
	if err != nil {
		return nil, err
	}

	db := s.db.db
	rows, err := db.QueryContext(ctx, query.SQL, query.Args...)
	// 数据库执行错误，返回err
	if err != nil {
		return nil, err
	}
	// 确认rows中是否有查询结果，若无，则跑出异常
	if !rows.Next() {
		return nil, ErrorNoRows
	}

	tp := new(T)
	val := s.db.creator(s.model, tp)
	err = val.SetColumn(rows)
	return tp, err
}

func (s *Selector[T]) GetMulti(ctx context.Context) ([]*T, error) {
	//TODO implement me
	panic("implement me")
}

func (s *Selector[T]) Select(cols ...Selectable) *Selector[T] {
	s.columns = cols
	return s
}
