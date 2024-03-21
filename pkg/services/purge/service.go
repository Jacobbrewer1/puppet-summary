package purge

import (
	"github.com/Jacobbrewer1/puppet-summary/pkg/dataaccess"
)

type Purger interface {
	PurgePuppetReports(purgeDays int) error

	SetupPurge(purgeDays int) error
}

type service struct {
	db dataaccess.Database
}

func NewService(db dataaccess.Database) Purger {
	return &service{
		db: db,
	}
}
