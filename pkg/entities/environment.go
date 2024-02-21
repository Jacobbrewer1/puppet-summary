package entities

type Environment string

const (
	EnvProduction  Environment = "PRODUCTION"
	EvnStaging     Environment = "STAGING"
	EnvDevelopment Environment = "DEVELOPMENT"

	// EnvAll is used to get all environments.
	EnvAll Environment = "ALL"
)

func (e Environment) String() string {
	return string(e)
}

func (e Environment) IsIn(environments ...Environment) bool {
	for _, environment := range environments {
		if environment == e {
			return true
		}
	}
	return false
}

func (e Environment) Valid() bool {
	return e.IsIn(
		EnvProduction,
		EvnStaging,
		EnvDevelopment,
	)
}
