package entities

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

func (p *PuppetHistory) AddCount(state State, count int) {
	switch state {
	case StateChanged:
		p.Changed += count
	case StateUnchanged:
		p.Unchanged += count
	case StateFailed:
		p.Failed += count
	}
}
