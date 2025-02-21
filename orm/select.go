package orm

import (
	"context"
	"imola/orm/internal/errs"
	"strings"
)

type Selector[T any] struct {
	table    string
	model    *Model
	where    []Predicate
	sBuilder *strings.Builder
	args     []any
}

func (s Selector[T]) Build() (*Query, error) {
	s.sBuilder = &strings.Builder{}
	var err error
	s.model, err = ParseModel(new(T))
	if err != nil {
		return nil, err
	}

	sBuilder := s.sBuilder
	sBuilder.WriteString("SELECT * FROM ")
	if s.table == "" {
		sBuilder.WriteByte('`')
		sBuilder.WriteString(s.model.TableName)
		sBuilder.WriteByte('`')
	} else {
		//segs := strings.Split(s.table, ".")
		//sBuilder.WriteByte('`')
		//sBuilder.WriteString(segs[0])
		//sBuilder.WriteByte('`')
		//sBuilder.WriteByte('.')
		//sBuilder.WriteByte('`')
		//sBuilder.WriteString(segs[1])
		//sBuilder.WriteByte('`')
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
		fd, ok := s.model.Fields[exp.name]
		// 字段不对，或者说列不对
		if !ok {
			return errs.NewErrUnknownField(exp.name)
		}
		s.sBuilder.WriteByte('`')
		s.sBuilder.WriteString(fd.Column)
		s.sBuilder.WriteByte('`')
	case value:
		s.sBuilder.WriteByte('?')
		s.AddArg(exp.val)
	default:
		return errs.NewErrUnsupportedExpression(expr)
	}
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

func (s *Selector[T]) Get(ctx context.Context) (*interface{}, error) {
	//TODO implement me
	panic("implement me")
}

func (s *Selector[T]) GetMulti(ctx context.Context) ([]*interface{}, error) {
	//TODO implement me
	panic("implement me")
}
