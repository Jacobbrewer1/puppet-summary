package entities

import "time"

// PuppetRun is the structure which is used to list a summary of puppet
// runs on the front-page.
type PuppetRun struct {
	ID        string      `json:"id" bson:"id"`
	Fqdn      string      `json:"fqdn" bson:"fqdn"`
	Env       Environment `json:"env" bson:"env"`
	State     State       `json:"state" bson:"state"`
	ExecTime  Datetime    `json:"exec_time" bson:"exec_time"`
	Runtime   Duration    `json:"runtime" bson:"runtime"`
	TimeSince Duration    `json:"time_since" bson:"time_since"`
}

func (p *PuppetRun) CalculateTimeSince() {
	p.TimeSince = Duration(time.Since(p.ExecTime.Time()))
}
