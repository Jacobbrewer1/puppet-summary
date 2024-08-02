package vault

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/Jacobbrewer1/puppet-summary/pkg/logging"
	vault "github.com/hashicorp/vault/api"
	auth "github.com/hashicorp/vault/api/auth/userpass"
	"github.com/spf13/viper"
)

type userPassClient struct {
	v        *vault.Client
	authInfo *vault.Secret
	vip      *viper.Viper
}

func NewClientUserPass(v *viper.Viper) (Client, error) {
	config := vault.DefaultConfig()
	config.Address = v.GetString("vault.address")

	c, err := vault.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize Vault client: %w", err)
	}

	clientImpl := &userPassClient{
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

func (c *userPassClient) renewAuthInfo() {
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

func (c *userPassClient) login() (*vault.Secret, error) {
	// WARNING: A plaintext password like this is obviously insecure.
	// See the hashicorp/vault-examples repo for full examples of how to securely
	// log in to Vault using various auth methods. This function is just
	// demonstrating the basic idea that a *vault.Secret is returned by
	// the login call.
	userpassAuth, err := auth.NewUserpassAuth(c.vip.GetString("vault.auth.username"), &auth.Password{FromString: c.vip.GetString("vault.auth.password")})
	if err != nil {
		return nil, fmt.Errorf("unable to initialize userpass auth method: %w", err)
	}

	authInfo, err := c.v.Auth().Login(context.Background(), userpassAuth)
	if err != nil {
		return nil, fmt.Errorf("unable to login to userpass auth method: %w", err)
	}
	if authInfo == nil {
		return nil, fmt.Errorf("no auth info was returned after login")
	}

	return authInfo, nil
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
func (c *userPassClient) RenewLease(ctx context.Context, name string, credentials *vault.Secret, renewFunc renewalFunc) error {
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

func (c *userPassClient) leaseRenew(ctx context.Context, name string, credentials *vault.Secret) (renewResult, error) {
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

func (c *userPassClient) GetSecrets(path string) (*Secrets, error) {
	secret, err := c.v.Logical().Read(path)
	if err != nil {
		return nil, fmt.Errorf("unable to read secrets: %w", err)
	} else if secret == nil {
		return nil, ErrSecretNotFound
	}
	return &Secrets{secret}, nil
}
