package orm

import (
	"context"
	"github.com/dongma/imola/orm/internal/errs"
	"github.com/dongma/imola/orm/sql"
)

// Selectable 是一个标记接口，代表查找的列或聚合函数等，例如：select xx 这部分
type Selectable interface {
	Selectable()
}

type Selector[T any] struct {
	builder
	table   TableReference
	where   []sql.Predicate
	columns []Selectable
	// 绑定session会话信息
	sess Session
}

func NewSelector[T any](sess Session) *Selector[T] {
	core := sess.getCore()
	return &Selector[T]{
		builder: builder{
			core:   core,
			quoter: core.dialect.Quoter(),
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

// BuildExpression 在这里处理predicate, 构建p.Left、p.op和p.Right
func (s *Selector[T]) BuildExpression(expr sql.Expression) error {
	switch exp := expr.(type) {
	case nil:
		return nil
	case sql.Predicate:
		_, ok := exp.Left.(sql.Predicate)
		if ok {
			s.sb.WriteByte('(')
		}
		if err := s.BuildExpression(exp.Left); err != nil {
			return err
		}
		if ok {
			s.sb.WriteByte(')')
		}

		if exp.Op != "" {
			s.sb.WriteByte(' ')
			s.sb.WriteString(exp.Op.String())
			s.sb.WriteByte(' ')
		}
		_, ok = exp.Right.(sql.Predicate)
		if ok {
			s.sb.WriteByte('(')
		}
		if err := s.BuildExpression(exp.Right); err != nil {
			return err
		}
		if ok {
			s.sb.WriteByte(')')
		}
	case Column:
		exp.Alias = ""
		return s.buildColumn(exp)
	case sql.Value:
		s.sb.WriteByte('?')
		s.AddArg(exp.Val)
	case sql.RawExpr:
		s.sb.WriteByte('(')
		s.sb.WriteString(exp.Raw)
		s.AddArg(exp.Args...)
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
		case sql.Aggregate:
			s.sb.WriteString(c.Fn)
			s.sb.WriteByte('(')
			err := s.buildColumn(Column{Name: c.Arg})
			if err != nil {
				return err
			}
			s.sb.WriteByte(')')
			// 聚合函数本身的别名
			if c.Alias != "" {
				s.sb.WriteString(" AS `")
				s.sb.WriteString(c.Alias)
				s.sb.WriteByte('`')
			}
		case sql.RawExpr:
			s.sb.WriteString(c.Raw)
			s.AddArg(c.Args...)
		}
	}
	return nil
}

func (s *Selector[T]) buildColumn(col Column) error {
	switch table := col.Table.(type) {
	case nil:
		fd, ok := s.model.FieldMap[col.Name]
		// 字段不对，或者说列不对
		if !ok {
			return errs.NewErrUnknownField(col.Name)
		}
		s.quote(fd.Column)
		if col.Alias != "" {
			s.sb.WriteString(" AS ")
			s.quote(col.Alias)
		}
	case Table:
		m, err := s.r.Get(table.Entity)
		if err != nil {
			return err
		}
		fd, ok := m.FieldMap[col.Name]
		if !ok {
			return errs.NewErrUnknownField(col.Name)
		}
		if table.Alias != "" {
			s.quote(table.Alias)
			s.sb.WriteByte('.')
		}
		s.quote(fd.Column)
		if col.Alias != "" {
			s.sb.WriteString(" AS ")
			s.quote(col.Alias)
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

func (s *Selector[T]) Where(conds ...sql.Predicate) *Selector[T] {
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
			Err: sql.ErrorNoRows,
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
		m, err := s.r.Get(t.Entity)
		if err != nil {
			return err
		}
		s.quote(m.TableName)
		if t.Alias != "" {
			s.sb.WriteString(" AS ")
			s.quote(t.Alias)
		}
	case Join:
		s.sb.WriteByte('(')
		// 构造左边
		err := s.buildTable(t.Left)
		if err != nil {
			return err
		}
		s.sb.WriteByte(' ')
		s.sb.WriteString(t.Typ)
		s.sb.WriteByte(' ')
		// 构造右边
		err = s.buildTable(t.Right)
		if err != nil {
			return err
		}

		// 拼接 USING (xx, xx)
		if len(t.Using) > 0 {
			s.sb.WriteString(" USING (")
			for i, col := range t.Using {
				if i > 0 {
					s.sb.WriteByte(',')
				}
				err = s.buildColumn(Column{Name: col})
				if err != nil {
					return err
				}
			}
			s.sb.WriteByte(')')
		}

		if len(t.On) > 0 {
			s.sb.WriteString(" ON ")
			p := t.On[0]
			for i := 1; i < len(t.On); i++ {
				p = p.And(t.On[i])
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
