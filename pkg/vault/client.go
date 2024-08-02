package vault

import (
	"context"

	vault "github.com/hashicorp/vault/api"
)

var (
	ErrSecretNotFound = vault.ErrSecretNotFound
)

type renewalFunc func() (*vault.Secret, error)

type Secrets struct {
	*vault.Secret
}

type Client interface {
	// GetSecrets returns a map of secrets for the given path.
	GetSecrets(path string) (*Secrets, error)

	// RenewLease renews the lease of the given credentials.
	RenewLease(ctx context.Context, name string, credentials *vault.Secret, renewFunc renewalFunc) error
}
