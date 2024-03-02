package entities

import (
	"time"

	"github.com/Jacobbrewer1/puppet-summary/pkg/codegen/apis/summary"
)

// PuppetRun is the structure which is used to list a summary of puppet
// runs on the front-page.
type PuppetRun struct {
	ID        string              `json:"id,omitempty" bson:"id"`
	Fqdn      string              `json:"fqdn" bson:"fqdn"`
	Env       summary.Environment `json:"env" bson:"env"`
	State     summary.State       `json:"state" bson:"state"`
	ExecTime  Datetime            `json:"exec_time" bson:"exec_time"`
	Runtime   Duration            `json:"runtime" bson:"runtime"`
	TimeSince Duration            `json:"-" bson:"time_since"`
}

func (p *PuppetRun) CalculateTimeSince() {
	p.TimeSince = Duration(time.Since(p.ExecTime.Time()))
}
