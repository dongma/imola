package orm

import (
	"context"
	"reflect"
	"strings"
)

type Selector[T any] struct {
	table string
}

func (s Selector[T]) Build() (*Query, error) {
	var sBuilder strings.Builder
	sBuilder.WriteString("SELECT * FROM ")
	if s.table == "" {
		var t T
		typ := reflect.TypeOf(t)
		sBuilder.WriteByte('`')
		sBuilder.WriteString(typ.Name())
		sBuilder.WriteByte('`')
	} else {
		segs := strings.Split(s.table, ".")
		sBuilder.WriteByte('`')
		sBuilder.WriteString(segs[0])
		sBuilder.WriteByte('`')
		sBuilder.WriteByte('.')
		sBuilder.WriteByte('`')
		sBuilder.WriteString(segs[1])
		sBuilder.WriteByte('`')
	}
	sBuilder.WriteByte(';')
	return &Query{
		SQL: sBuilder.String(),
	}, nil
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
