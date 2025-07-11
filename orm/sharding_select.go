package orm

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/dongma/imola/orm/internal/errs"
	isql "github.com/dongma/imola/orm/sql"
	"golang.org/x/sync/errgroup"
	"strings"
)

type ShardingSelect[T any] struct {
	builder
	table   *T
	where   []isql.Predicate
	columns []Selectable

	sess Session
	db   *ShardingDB
	// 这边需要一个查询特征的东西
	isDistinct bool
	orderBy    []string
	offset     int
	limit      int
}

type ShardingQuery struct {
	SQL  string
	Args []any
	DB   string
}

// Build k是sharding key, fn就是分库分表的算法
func (s *ShardingSelect[T]) Build() ([]*ShardingQuery, error) {
	// 对映射的模型做处理，T对应哪个实体
	if s.model == nil {
		var err error
		s.model, err = s.r.Get(new(T))
		if err != nil {
			return nil, err
		}
	}
	dsts, err := s.findDsts()
	if err != nil {
		return nil, err
	}
	res := make([]*ShardingQuery, 0, len(dsts))
	for _, dst := range dsts {
		q, err := s.build(dst.DB, dst.Table)
		if err != nil {
			return nil, err
		}
		s.sb = strings.Builder{}
		res = append(res, q)
	}
	return res, nil
}

func (s *ShardingSelect[T]) build(db, tbl string) (*ShardingQuery, error) {
	s.sb.WriteString("SELECT ")
	if err := s.buildColumns(); err != nil {
		return nil, err
	}
	s.sb.WriteString(" FROM ")
	s.sb.WriteString(fmt.Sprintf("%s.%s", db, tbl))

	if len(s.where) > 0 {
		s.sb.WriteString(" WHERE ")
		p := s.where[0]
		for i := 0; i < len(s.where); i++ {
			p = p.And(s.where[i])
		}
		if err := s.buildExpression(p); err != nil {
			return nil, err
		}
	}
	s.sb.WriteByte(';')
	return &ShardingQuery{
		SQL:  s.sb.String(),
		Args: s.args,
		DB:   db,
	}, nil
}

// findDsts []Dst所有候选的目标节点，error是否出错
func (s *ShardingSelect[T]) findDsts() ([]Dst, error) {
	// 在这里深入（递归）到where部分
	if len(s.where) > 0 {
		s.sb.WriteString(" WHERE ")
		p := s.where[0]
		for i := 1; i < len(s.where); i++ {
			p = p.And(s.where[i])
		}
		// 在这里，空切片意味着，不需要发请求到任何节点
		// where user_id = 123 and user_id = 124
		return s.findDstByPredicate(p)
	}
	// 这边是要广播
	panic("implement me")
}

// WHERE id = 11 AND user_id = 123
// WHERE user_id = 123 AND id = 11
// WHERE user_id = 123 AND user_id IN (123, 124)
// WHERE user_id = 123 AND user_id = 124
// findDstByPredicate 根据where条件，查找要将请求发送到哪些数据库节点上
func (s *ShardingSelect[T]) findDstByPredicate(p isql.Predicate) ([]Dst, error) {
	var res []Dst
	switch p.Op {
	case isql.OpAnd:
		// 空切片意味着广播，case1: right有一个，case2: right是广播
		right, err := s.findDstByPredicate(p.Right.(isql.Predicate))
		if err != nil {
			return nil, err
		}
		if len(right) == 0 {
			// 说明广播，case2 进来这里
			return s.findDstByPredicate(p.Right.(isql.Predicate))
		}
		// case1: left是广播, 进入到这里
		left, err := s.findDstByPredicate(p.Left.(isql.Predicate))
		if len(left) == 0 {
			return right, nil
		}
		// 求交集, case3 进来这里
		return s.merge(left, right), nil
	case isql.OpOr:
	//case sql.OpLt:
	//case sql.OpRt:
	case isql.OpEq:
		left, ok := p.Left.(Column)
		if ok {
			// where id = 123
			right, ok := p.Right.(isql.Value)
			if !ok {
				return nil, fmt.Errorf("太复杂的查询，暂时不支持")
			}
			if s.model.SK == left.Name && ok {
				db, tbl := s.model.Sf(right.Val)
				res = append(res, Dst{DB: db, Table: tbl})
			}
		}
	default:
		return nil, fmt.Errorf("orm: 不知道怎么处理操作符")
	}
	return res, nil
}

// merge 合并左右的候选节点列表,left和right (求交集)
func (s *ShardingSelect[T]) merge(left, right []Dst) []Dst {
	res := make([]Dst, len(left)+len(right))
	for _, r := range right {
		exist := false
		for _, l := range left {
			if r.DB == l.DB && r.Table == l.Table {
				exist = true
			}
		}
		if exist {
			res = append(res, r)
		}
	}
	return res
}

func (s *ShardingSelect[T]) buildColumns() error {
	if len(s.columns) == 0 {
		// 没有指定列
		s.sb.WriteByte('*')
		return nil
	}
	for i, col := range s.columns {
		if i > 0 {
			s.sb.WriteByte(',')
		}
		switch c := col.(type) {
		case Column:
			err := s.buildColumn(c.Name)
			if err != nil {
				return err
			}
		case isql.Aggregate:
			// 聚合函数，编译有问题，暂时先不提供
			/*s.sb.WriteString(c.Fn)
			s.sb.WriteByte('(')
			err := s.buildColumn(Column{name: c.arg})
			if err != nil {
				return err
			}
			s.sb.WriteByte(')')
			// 聚合函数本身的别名
			if c.Alias != "" {
				s.sb.WriteString(" AS `")
				s.sb.WriteString(c.Alias)
				s.sb.WriteByte('`')
			} */
		case isql.RawExpr:
			s.sb.WriteString(c.Raw)
			s.AddArg(c.Args...)
		}
	}
	return nil
}

func (s *ShardingSelect[T]) buildExpression(expr isql.Expression) error {
	switch expr := expr.(type) {
	case nil:
	case isql.Predicate:
		// 在这里处理好p, p.left构建好、p.right构建好
		_, ok := expr.Left.(isql.Predicate)
		if ok {
			s.sb.WriteByte('(')
		}
		if err := s.buildExpression(expr.Left); err != nil {
			return err
		}
		if ok {
			s.sb.WriteByte(')')
		}

		if expr.Op != "" {
			s.sb.WriteByte(' ')
			s.sb.WriteString(expr.Op.String())
			s.sb.WriteByte(' ')
		}

		// 处理右边的断言
		_, ok = expr.Right.(isql.Predicate)
		if ok {
			s.sb.WriteByte('(')
		}
		if err := s.buildExpression(expr.Right); err != nil {
			return err
		}
		if ok {
			s.sb.WriteByte(')')
		}
	case Column:
		// 这种写法很隐晦
		expr.Alias = ""
		return s.buildColumn(expr.Name)
	case isql.Value:
		s.sb.WriteByte('?')
		s.AddArg(expr.Val)
	case isql.RawExpr:
		s.sb.WriteByte('(')
		s.sb.WriteString(expr.Raw)
		s.AddArg(expr.Args...)
		s.sb.WriteByte(')')
	default:
		return errs.NewErrUnsupportedExpression(expr)
	}
	return nil
}

type Dst struct {
	DB    string
	Table string
}

func (s *ShardingSelect[T]) GetMulti(ctx context.Context) ([]*T, error) {
	qs, err := s.Build()
	if err != nil {
		return nil, err
	}
	var resSlice []*sql.Rows
	var eg errgroup.Group
	for _, query := range qs {
		q := query
		eg.Go(func() error {
			db, ok := s.db.DBs[q.DB]
			if !ok {
				// 可能是用户配置不对，也可能是你的框架不对
				return errors.New("orm: 非法的目标库")
			}
			// 要决策用master还是slave
			rows, err := db.query(ctx, q.SQL, q.Args...)
			if err == nil {
				resSlice = append(resSlice, rows)
			}
			return err
		})
	}

	err = eg.Wait()
	if err != nil {
		return nil, err
	}
	// 你已经把所有的结果取过来了，在这里合并结果集
	var res []*T
	for _, rows := range resSlice {
		for rows.Next() {
			t := new(T)
			val := s.creator(s.model, t)
			err = val.SetColumn(rows)
			if err != nil {
				return nil, err
			}
			res = append(res, t)
		}
	}
	return res, nil
}
