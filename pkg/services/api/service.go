package api

import (
	"github.com/Jacobbrewer1/puppet-summary/pkg/codegen/apis/summary"
	"github.com/Jacobbrewer1/puppet-summary/pkg/dataaccess"
)

type service struct {
	r dataaccess.Database
}

func NewService(r dataaccess.Database) summary.ServerInterface {
	return &service{
		r: r,
	}
}
