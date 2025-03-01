package orm

import (
	"imola/orm/internal/errs"
)

var (
	DialectMySQL      Dialect = mysqlDialect{}
	DialectSQLite     Dialect = sqliteDialect{}
	DialectPostgreSQL Dialect = postgreDialect{}
)

type Dialect interface {
	// quoter 返回一个引号，引用列名，表的引号
	quoter() byte
	buildOnDuplicateKey(b *builder, odk *OnDuplicateKey) error
}

type standardSQL struct {
}

func (s standardSQL) quoter() byte {
	panic("implement me")
}

func (s standardSQL) buildOnDuplicateKey(b *builder, odk *OnDuplicateKey) error {
	panic("implement me")
}

type mysqlDialect struct {
	standardSQL
}

func (s mysqlDialect) quoter() byte {
	return '`'
}

func (s mysqlDialect) buildOnDuplicateKey(b *builder, odk *OnDuplicateKey) error {
	b.sb.WriteString(" ON DUPLICATE KEY UPDATE ")
	for idx, assign := range odk.assigns {
		if idx > 0 {
			b.sb.WriteByte(',')
		}
		switch asi := assign.(type) {
		case Assignment:
			fd, ok := b.model.FieldMap[asi.col]
			if !ok {
				return errs.NewErrUnknownColumn(asi.col)
			}
			b.quote(fd.Column)
			b.sb.WriteString("=?")
			b.AddArg(asi.val)
		case Column:
			fd, ok := b.model.FieldMap[asi.name]
			if !ok {
				return errs.NewErrUnknownField(asi.name)
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

type postgreDialect struct {
	standardSQL
}
