package dataaccess

type StoreType string

const (
	// StoreTypeLocal is a local file system store
	StoreTypeLocal StoreType = "LOCAL"

	// StoreTypeS3 is an S3 store
	StoreTypeS3 StoreType = "S3"

	// StoreTypeGCS is a Google Cloud fileHandler store
	StoreTypeGCS StoreType = "GCS"
)
