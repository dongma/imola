package orm

import (
	"github.com/dongma/imola/orm/internal/errs"
	"github.com/dongma/imola/orm/sql"
)

var (
	DialectMySQL      Dialect = mysqlDialect{}
	DialectSQLite     Dialect = sqliteDialect{}
	DialectPostgreSQL Dialect = postgreDialect{}
)

// Dialect 方言抽象
type Dialect interface {
	// Quoter 返回一个引号，引用列名，表的引号
	Quoter() byte
	BuildOnDuplicateKey(b *builder, odk *OnDuplicateKey) error
}

// standardSQL 实现标准sql方案，其它方言继承standardSQL
type standardSQL struct {
}

func (s standardSQL) Quoter() byte {
	panic("implement me")
}

func (s standardSQL) BuildOnDuplicateKey(b *builder, odk *OnDuplicateKey) error {
	panic("implement me")
}

// mysqlDialect mysql onDuplicateKey的实现
type mysqlDialect struct {
	standardSQL
}

func (s mysqlDialect) Quoter() byte {
	return '`'
}

// BuildOnDuplicateKey mysql数据库对DuplicateKey的支持
func (s mysqlDialect) BuildOnDuplicateKey(b *builder, odk *OnDuplicateKey) error {
	b.sb.WriteString(" ON DUPLICATE KEY UPDATE ")
	for idx, assign := range odk.assigns {
		if idx > 0 {
			b.sb.WriteByte(',')
		}
		switch asi := assign.(type) {
		case sql.Assignment:
			fd, ok := b.model.FieldMap[asi.Col]
			if !ok {
				return errs.NewErrUnknownColumn(asi.Col)
			}
			b.quote(fd.Column)
			b.sb.WriteString("=?")
			b.AddArg(asi.Val)
		case Column:
			fd, ok := b.model.FieldMap[asi.Name]
			if !ok {
				return errs.NewErrUnknownField(asi.Name)
			}
			b.quote(fd.Column)
			b.sb.WriteString("=VALUES(")
			b.quote(fd.Column)
			b.sb.WriteByte(')')
		default:
			return errs.NewErrUnsupportedAssignable(assign)
		}
	}
	return nil
}

type sqliteDialect struct {
	standardSQL
}

func (s sqliteDialect) Quoter() byte {
	return '`'
}

// BuildOnDuplicateKey sqlite数据库对DuplicateKey的支持
func (s sqliteDialect) BuildOnDuplicateKey(b *builder, odk *OnDuplicateKey) error {
	b.sb.WriteString(" ON CONFLICT(")
	for i, col := range odk.conflictColumns {
		if i > 0 {
			b.sb.WriteByte(',')
		}
		err := b.buildColumn(col)
		if err != nil {
			return err
		}
	}
	b.sb.WriteString(") DO UPDATE SET ")
	for idx, assign := range odk.assigns {
		if idx > 0 {
			b.sb.WriteByte(',')
		}
		switch asi := assign.(type) {
		case sql.Assignment:
			fd, ok := b.model.FieldMap[asi.Col]
			if !ok {
				return errs.NewErrUnknownColumn(asi.Col)
			}
			b.quote(fd.Column)
			b.sb.WriteString("=?")
			b.AddArg(asi.Val)
		case Column:
			fd, ok := b.model.FieldMap[asi.Name]
			if !ok {
				return errs.NewErrUnknownField(asi.Name)
			}
			b.quote(fd.Column)
			b.sb.WriteString("=excluded.")
			b.quote(fd.Column)
		default:
			return errs.NewErrUnsupportedAssignable(assign)
		}
	}
	return nil
}

type postgreDialect struct {
	standardSQL
}
