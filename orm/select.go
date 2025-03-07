package orm

import (
	"context"
	"imola/orm/internal/errs"
)

// Selectable 是一个标记接口，代表查找的列或聚合函数等，例如：select xx 这部分
type Selectable interface {
	selectable()
}

type Selector[T any] struct {
	builder
	table   TableReference
	where   []Predicate
	columns []Selectable
	sess    Session
}

func NewSelector[T any](sess Session) *Selector[T] {
	core := sess.getCore()
	return &Selector[T]{
		builder: builder{
			core:   core,
			quoter: core.dialect.quoter(),
		},
		sess: sess,
	}
}

func (s *Selector[T]) Build() (*Query, error) {
	if s.model == nil {
		var err error
		s.model, err = s.r.Get(new(T))
		if err != nil {
			return nil, err
		}
	}
	s.sb.WriteString("SELECT ")
	err := s.buildColumns()
	if err != nil {
		return nil, err
	}

	s.sb.WriteString(" FROM ")
	// 构建数据表或子查询
	if err := s.buildTable(s.table); err != nil {
		return nil, err
	}

	if len(s.where) > 0 {
		s.sb.WriteString(" WHERE ")
		predicate := s.where[0]
		for i := 1; i < len(s.where); i++ {
			predicate = predicate.And(s.where[i])
		}
		if err := s.BuildExpression(predicate); err != nil {
			return nil, err
		}
	}

	s.sb.WriteByte(';')
	return &Query{
		SQL:  s.sb.String(),
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
			s.sb.WriteByte('(')
		}
		if err := s.BuildExpression(exp.left); err != nil {
			return err
		}
		if ok {
			s.sb.WriteByte(')')
		}

		if exp.op != "" {
			s.sb.WriteByte(' ')
			s.sb.WriteString(exp.op.String())
			s.sb.WriteByte(' ')
		}
		_, ok = exp.right.(Predicate)
		if ok {
			s.sb.WriteByte('(')
		}
		if err := s.BuildExpression(exp.right); err != nil {
			return err
		}
		if ok {
			s.sb.WriteByte(')')
		}
	case Column:
		exp.alias = ""
		return s.buildColumn(exp)
	case value:
		s.sb.WriteByte('?')
		s.AddArg(exp.val)
	case RawExpr:
		s.sb.WriteByte('(')
		s.sb.WriteString(exp.raw)
		s.AddArg(exp.args...)
		s.sb.WriteByte(')')
	default:
		return errs.NewErrUnsupportedExpression(expr)
	}
	return nil
}

// buildColumns 为了避免方法膨胀，将构建sql列单独抽成一个方法
func (s *Selector[T]) buildColumns() error {
	if len(s.columns) == 0 {
		s.sb.WriteString("*")
		return nil
	}
	for i, col := range s.columns {
		if i > 0 {
			s.sb.WriteByte(',')
		}
		switch c := col.(type) {
		case Column:
			err := s.buildColumn(c)
			if err != nil {
				return err
			}
		case Aggregate:
			s.sb.WriteString(c.fn)
			s.sb.WriteByte('(')
			err := s.buildColumn(Column{name: c.arg})
			if err != nil {
				return err
			}
			s.sb.WriteByte(')')
			// 聚合函数本身的别名
			if c.alias != "" {
				s.sb.WriteString(" AS `")
				s.sb.WriteString(c.alias)
				s.sb.WriteByte('`')
			}
		case RawExpr:
			s.sb.WriteString(c.raw)
			s.AddArg(c.args...)
		}
	}
	return nil
}

func (s *Selector[T]) buildColumn(col Column) error {
	switch table := col.table.(type) {
	case nil:
		fd, ok := s.model.FieldMap[col.name]
		// 字段不对，或者说列不对
		if !ok {
			return errs.NewErrUnknownField(col.name)
		}
		s.quote(fd.Column)
		if col.alias != "" {
			s.sb.WriteString(" AS ")
			s.quote(col.alias)
		}
	case Table:
		m, err := s.r.Get(table.entity)
		if err != nil {
			return err
		}
		fd, ok := m.FieldMap[col.name]
		if !ok {
			return errs.NewErrUnknownField(col.name)
		}
		if table.alias != "" {
			s.quote(table.alias)
			s.sb.WriteByte('.')
		}
		s.quote(fd.Column)
		if col.alias != "" {
			s.sb.WriteString(" AS ")
			s.quote(col.alias)
		}
	default:
		return errs.NewErrUnsupportedTable(table)
	}
	return nil
}

func (s *Selector[T]) AddArg(vals ...any) {
	if len(vals) == 0 {
		return
	}
	if s.args == nil {
		s.args = make([]any, 0, 8)
	}
	s.args = append(s.args, vals...)
}

func (s *Selector[T]) Where(conds ...Predicate) *Selector[T] {
	s.where = conds
	return s
}

func (s *Selector[T]) From(tbl TableReference) *Selector[T] {
	s.table = tbl
	return s
}

var _ Handler = (&Selector[any]{}).getHandler

func (s *Selector[T]) getHandler(ctx context.Context, qc *QueryContext) *QueryResult {
	query, err := s.Build()
	if err != nil {
		return &QueryResult{
			Err: err,
		}
	}

	rows, err := s.sess.queryContext(ctx, query.SQL, query.Args...)
	// 数据库执行错误，返回err
	if err != nil {
		return &QueryResult{
			Err: err,
		}
	}
	// 确认rows中是否有查询结果，若无，则跑出异常
	if !rows.Next() {
		return &QueryResult{
			Err: ErrorNoRows,
		}
	}

	tp := new(T)
	val := s.creator(s.model, tp)
	err = val.SetColumn(rows)
	return &QueryResult{
		Err:    err,
		Result: tp,
	}
}

func (s *Selector[T]) Get(ctx context.Context) (*T, error) {
	var err error
	s.model, err = s.r.Get(new(T))
	if err != nil {
		return nil, err
	}
	root := s.getHandler
	for i := len(s.mdls) - 1; i >= 0; i-- {
		root = s.mdls[i](root)
	}
	res := root(ctx, &QueryContext{
		Type:    "SELECT",
		Builder: s,
		Model:   s.model,
	})
	if res.Result != nil {
		return res.Result.(*T), res.Err
	}
	return nil, res.Err
}

func (s *Selector[T]) GetMulti(ctx context.Context) ([]*T, error) {
	//TODO implement me
	panic("implement me")
}

func (s *Selector[T]) Select(cols ...Selectable) *Selector[T] {
	s.columns = cols
	return s
}

func (s *Selector[T]) buildTable(table TableReference) error {
	switch t := table.(type) {
	case nil:
		// 这个地方代表完全没有调用from，是最普通的形态
		s.quote(s.model.TableName)
	case Table:
		// 这个地方是拿到指定表的元数据
		m, err := s.r.Get(t.entity)
		if err != nil {
			return err
		}
		s.quote(m.TableName)
		if t.alias != "" {
			s.sb.WriteString(" AS ")
			s.quote(t.alias)
		}
	case Join:
		s.sb.WriteByte('(')
		// 构造左边
		err := s.buildTable(t.left)
		if err != nil {
			return err
		}
		s.sb.WriteByte(' ')
		s.sb.WriteString(t.typ)
		s.sb.WriteByte(' ')
		// 构造右边
		err = s.buildTable(t.right)
		if err != nil {
			return err
		}

		// 拼接 USING (xx, xx)
		if len(t.using) > 0 {
			s.sb.WriteString(" USING (")
			for i, col := range t.using {
				if i > 0 {
					s.sb.WriteByte(',')
				}
				err = s.buildColumn(Column{name: col})
				if err != nil {
					return err
				}
			}
			s.sb.WriteByte(')')
		}

		if len(t.on) > 0 {
			s.sb.WriteString(" ON ")
			p := t.on[0]
			for i := 1; i < len(t.on); i++ {
				p = p.And(t.on[i])
			}
			if err = s.BuildExpression(p); err != nil {
				return err
			}
		}
		s.sb.WriteByte(')')
	default:
		return errs.NewErrUnsupportedTable(table)
	}
	return nil
}
