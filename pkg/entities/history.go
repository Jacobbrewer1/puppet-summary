package entities

import "github.com/Jacobbrewer1/puppet-summary/pkg/codegen/apis/summary"

type PuppetHistory struct {
	// Date is the date of the run.
	Date string `json:"date" bson:"date"`

	// Changed is the number of resources that changed.
	Changed int `json:"changed" bson:"changed"`

	// Unchanged is the number of resources that were unchanged.
	Unchanged int `json:"unchanged" bson:"unchanged"`

	// Failed is the number of resources that failed.
	Failed int `json:"failed" bson:"failed"`
}

func (p *PuppetHistory) AddCount(state summary.State, count int) {
	switch state {
	case summary.State_CHANGED:
		p.Changed += count
	case summary.State_UNCHANGED:
		p.Unchanged += count
	case summary.State_FAILED:
		p.Failed += count
	}
}
