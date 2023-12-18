package entities

// PuppetResource refers to a resource in your puppet modules, a resource has
// a name, along with the file & line-number it was defined in within your
// manifest
type PuppetResource struct {
	Name string `json:"name" bson:"name"`
	Type string `json:"type" bson:"type"`
	File string `json:"file" bson:"file"`
	Line string `json:"line" bson:"line"`
}
