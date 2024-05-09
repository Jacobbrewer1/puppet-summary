package vault

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	vault "github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/api/auth/approle"
	"github.com/spf13/viper"
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

type client struct {
	v        *vault.Client
	authInfo *vault.Secret
	vip      *viper.Viper
}

func NewClient(v *viper.Viper) (Client, error) {
	config := vault.DefaultConfig()
	config.Address = v.GetString("vault.host")

	c, err := vault.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize Vault client: %w", err)
	}

	clientImpl := &client{
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

func (c *client) login() (*vault.Secret, error) {
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

func (c *client) renewAuthInfo() {
	authTokenWatcher, err := c.v.NewLifetimeWatcher(&vault.LifetimeWatcherInput{
		Secret: c.authInfo,
	})
	if err != nil {
		slog.Error("unable to initialize auth token lifetime watcher", slog.String("error", err.Error()))
		os.Exit(1) // Kill the app to get new credentials
	}

	go authTokenWatcher.Start()
	defer authTokenWatcher.Stop()

	res, err := c.monitorWatcher(context.Background(), "authInfo", authTokenWatcher)
	if err != nil {
		slog.Error("unable to monitor watcher", slog.String("error", err.Error()))
		os.Exit(1) // Kill the app to get new credentials
	}

	onExpire := func() (*vault.Secret, error) {
		authInfo, err := c.login()
		if err != nil {
			return nil, fmt.Errorf("unable to login to Vault: %w", err)
		}

		return authInfo, nil
	}

	err = c.handleWatcherResult(res, func() {
		newAuthInfo, err := onExpire()
		if err != nil {
			slog.Error("unable to handle watcher result", slog.String("error", err.Error()))
			os.Exit(1) // Kill the app to get new credentials
		}

		c.authInfo = newAuthInfo
	})
	if err != nil {
		slog.Error("unable to handle watcher result", slog.String("error", err.Error()))
		os.Exit(1) // Kill the app to get new credentials
	}
}

func (c *client) handleWatcherResult(result renewResult, onExpire ...func()) error {
	switch {
	case result&exitRequested != 0:
		return nil
	case result&expiring != 0:
		if len(onExpire) == 0 {
			return fmt.Errorf("no onExpire functions provided")
		}
		for _, f := range onExpire {
			f()
		}
		return nil
	default:
		slog.Debug("no action required", slog.Int("result", int(result)))
		return nil
	}
}

func (c *client) monitorWatcher(ctx context.Context, name string, watcher *vault.LifetimeWatcher) (renewResult, error) {
	for {
		select {
		case <-ctx.Done():
			return exitRequested, nil

		// DoneCh will return if renewal fails, or if the remaining lease
		// duration is under a built-in threshold and either renewing is not
		// extending it or renewing is disabled.  In both cases, the caller
		// should attempt a re-read of the secret. Clients should check the
		// return value of the channel to see if renewal was successful.
		case err := <-watcher.DoneCh():
			// Leases created by a token get revoked when the token is revoked.
			return expiring, fmt.Errorf("renewal failed: %w", err)

		// RenewCh is a channel that receives a message when a successful
		// renewal takes place and includes metadata about the renewal.
		case info := <-watcher.RenewCh():
			slog.Info("renewal successful", slog.String("renewed_at", info.RenewedAt.String()),
				slog.String("secret", name), slog.String("lease_duration", fmt.Sprintf("%ds", info.Secret.LeaseDuration)))
		}
	}
}

func (c *client) GetSecrets(path string) (*Secrets, error) {
	secret, err := c.v.Logical().Read(path)
	if err != nil {
		return nil, fmt.Errorf("unable to read secrets: %w", err)
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
func (c *client) RenewLease(ctx context.Context, name string, credentials *vault.Secret, renewFunc renewalFunc) error {
	slog.Info("renewing lease", slog.String("secret", name))

	currentCreds := credentials

	for {
		res, err := c.leaseRenew(ctx, name, currentCreds)
		if err != nil {
			return fmt.Errorf("unable to renew lease: %w", err)
		}

		err = c.handleWatcherResult(res, func() {
			newCreds, err := renewFunc()
			if err != nil {
				slog.Error("unable to renew credentials", slog.String("error", err.Error()))
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

func (c *client) leaseRenew(ctx context.Context, name string, credentials *vault.Secret) (renewResult, error) {
	credentialsWatcher, err := c.v.NewLifetimeWatcher(&vault.LifetimeWatcherInput{
		Secret: credentials,
	})
	if err != nil {
		return renewError, fmt.Errorf("unable to initialize credentials lifetime watcher: %w", err)
	}

	go credentialsWatcher.Start()
	defer credentialsWatcher.Stop()

	res, err := c.monitorWatcher(ctx, name, credentialsWatcher)
	if err != nil {
		return renewError, fmt.Errorf("unable to monitor watcher: %w", err)
	}

	return res, nil
}
