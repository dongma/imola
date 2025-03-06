//go:build e2e

package integration

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"imola/orm"
	"testing"
	"time"
)

type SelectSuite struct {
	Suite
}

func TestMysqlSelect(t *testing.T) {
	suite.Run(t, &SelectSuite{
		Suite{
			driver: "mysql",
			dsn:    "root:root@tcp(localhost:13306)/integration_test",
		},
	})
}

func (s *SelectSuite) SetupSuite() {
	s.Suite.SetupSuite()
	res := orm.NewInserter[SimpleStruct](s.db).Values(
		NewSimpleStruct(100),
	).Exec(context.Background())
	require.NoError(s.T(), res.Err)
}

func (s *SelectSuite) TestGet() {
	testCases := []struct {
		name    string
		s       *orm.Selector[SimpleStruct]
		wantRes *SimpleStruct
		wantErr error
	}{
		{
			name:    "get data",
			s:       orm.NewSelector[SimpleStruct](s.db).Where(orm.C("Id").Eq(100)),
			wantRes: NewSimpleStruct(100),
		},
		{
			name:    "no row",
			s:       orm.NewSelector[SimpleStruct](s.db).Where(orm.C("Id").Eq(200)),
			wantErr: orm.ErrorNoRows,
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
			defer cancel()
			res, err := tc.s.Get(ctx)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantRes, res)
		})
	}
}
