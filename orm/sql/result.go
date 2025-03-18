package sql

import "database/sql"

type Result struct {
	Err error
	Res sql.Result
}

func (r Result) LastInsertId() (int64, error) {
	if r.Err != nil {
		return 0, r.Err
	}
	return r.LastInsertId()
}

func (r Result) RowsAffected() (int64, error) {
	if r.Err != nil {
		return 0, r.Err
	}
	return r.Res.RowsAffected()
}

func (r Result) Error() error {
	return r.Err
}
