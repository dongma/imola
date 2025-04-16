package orm

import (
	"context"
	"database/sql"
)

type ShardingDB struct {
	DBs map[string]*MasterSlavesDB
}

type MasterSlavesDB struct {
	Master *sql.DB
	Slaves []*sql.DB
}

func (m *MasterSlavesDB) query(ctx context.Context, sql string, args ...any) (*sql.Rows, error) {
	// 这边决定两件事：
	// 1、决定走master还是走slave
	// 2、如果走slave，怎么负载均衡
	db := m.Slaves[0]
	return db.QueryContext(ctx, sql, args...)
}
