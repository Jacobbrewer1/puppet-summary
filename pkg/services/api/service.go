package api

import (
	"github.com/Jacobbrewer1/puppet-summary/pkg/codegen/apis/summary"
	"github.com/Jacobbrewer1/puppet-summary/pkg/dataaccess"
	"github.com/Jacobbrewer1/puppet-summary/pkg/services/purge"
)

type service struct {
	// r is the repository used by the service.
	r dataaccess.Database

	// purger is the purge service used by the service.
	purger purge.Purger
}

func NewService(r dataaccess.Database, purger purge.Purger) summary.ServerInterface {
	return &service{
		r:      r,
		purger: purger,
	}
}
