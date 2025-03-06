package integration

import (
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"imola/orm"
)

type Suite struct {
	suite.Suite
	db     *orm.DB
	driver string
	dsn    string
}

func (s *Suite) SetupSuite() {
	db, err := orm.Open(s.driver, s.dsn)
	require.NoError(s.T(), err)
	s.db = db
}
