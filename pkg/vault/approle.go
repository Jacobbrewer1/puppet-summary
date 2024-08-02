package vault

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/Jacobbrewer1/puppet-summary/pkg/logging"
	vault "github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/api/auth/approle"
	"github.com/spf13/viper"
)

type appRoleClient struct {
	v        *vault.Client
	authInfo *vault.Secret
	vip      *viper.Viper
}

func NewClientAppRole(v *viper.Viper) (Client, error) {
	config := vault.DefaultConfig()
	config.Address = v.GetString("vault.address")

	c, err := vault.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize Vault client: %w", err)
	}

	clientImpl := &appRoleClient{
		v:   c,
		vip: v,
	}

	authInfo, err := clientImpl.login()
	if err != nil {
		return nil, fmt.Errorf("unable to login to Vault: %w", err)
	}

	clientImpl.authInfo = authInfo

	go clientImpl.renewAuthInfo()

	return clientImpl, nil
}

func (c *appRoleClient) login() (*vault.Secret, error) {
	vip := c.vip
	approleSecretID := &approle.SecretID{
		FromString: vip.GetString("vault.app_role_secret_id"),
	}

	// Authenticate with Vault with the AppRole auth method
	appRoleAuth, err := approle.NewAppRoleAuth(
		vip.GetString("vault.app_role_id"),
		approleSecretID,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create AppRole auth: %w", err)
	}

	authInfo, err := c.v.Auth().Login(context.Background(), appRoleAuth)
	if err != nil {
		return nil, fmt.Errorf("unable to authenticate with Vault: %w", err)
	}
	if authInfo == nil {
		return nil, errors.New("authentication with Vault failed")
	}

	return authInfo, nil
}

func (c *appRoleClient) renewAuthInfo() {
	err := c.RenewLease(context.Background(), "auth", c.authInfo, func() (*vault.Secret, error) {
		authInfo, err := c.login()
		if err != nil {
			return nil, fmt.Errorf("unable to renew auth info: %w", err)
		}

		c.authInfo = authInfo

		return authInfo, nil
	})
	if err != nil {
		slog.Error("unable to renew auth info", slog.String(logging.KeyError, err.Error()))
		os.Exit(1)
	}
}

func (c *appRoleClient) GetSecrets(path string) (*Secrets, error) {
	secret, err := c.v.Logical().Read(path)
	if err != nil {
		return nil, fmt.Errorf("unable to read secrets: %w", err)
	} else if secret == nil {
		return nil, ErrSecretNotFound
	}
	return &Secrets{secret}, nil
}

// RenewLease Once you've set the token for your Vault client, you will need to
// periodically renew it. Likewise, the database credentials lease will expire
// at some point and also needs to be renewed periodically.
//
// A function like this one should be run as a goroutine to avoid blocking.
// Production applications may also need to be more tolerant of failures and
// retry on errors rather than exiting.
//
// Additionally, enterprise Vault users should be aware that due to eventual
// consistency, the API may return unexpected errors when running Vault with
// performance standbys or performance replication, despite the client having
// a freshly renewed token. See the link below for several ways to mitigate
// this which are outside the scope of this code sample.
//
// ref: https://www.vaultproject.io/docs/enterprise/consistency#vault-1-7-mitigations
func (c *appRoleClient) RenewLease(ctx context.Context, name string, credentials *vault.Secret, renewFunc renewalFunc) error {
	slog.Info("renewing lease", slog.String("secret", name))

	currentCreds := credentials

	for {
		res, err := c.leaseRenew(ctx, name, currentCreds)
		if err != nil {
			return fmt.Errorf("unable to renew lease: %w", err)
		} else if res&exitRequested != 0 {
			// Context was cancelled. Program is exiting.
			slog.Debug("exit requested", slog.String("secret", name))
			return nil
		}

		err = handleWatcherResult(res, func() {
			newCreds, err := renewFunc()
			if err != nil {
				slog.Error("unable to renew credentials", slog.String(logging.KeyError, err.Error()))
				os.Exit(1) // Forces new credentials to be fetched
			}

			currentCreds = newCreds
		})
		if err != nil {
			return fmt.Errorf("unable to handle watcher result: %w", err)
		}

		slog.Info("lease renewed", slog.String("secret", name))
	}
}

func (c *appRoleClient) leaseRenew(ctx context.Context, name string, credentials *vault.Secret) (renewResult, error) {
	credentialsWatcher, err := c.v.NewLifetimeWatcher(&vault.LifetimeWatcherInput{
		Secret:    credentials,
		Increment: 3600,
	})
	if err != nil {
		return renewError, fmt.Errorf("unable to initialize credentials lifetime watcher: %w", err)
	}

	go credentialsWatcher.Start()
	defer credentialsWatcher.Stop()

	res, err := monitorWatcher(ctx, name, credentialsWatcher)
	if err != nil {
		return renewError, fmt.Errorf("unable to monitor watcher: %w", err)
	}

	return res, nil
}
