package connection

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDB struct {
	AppName          string
	ConnectionString string
	Username         string
	Password         string
	Host             string
	Port             string
	Args             string
}

func (m *MongoDB) GenerateConnectionString() {
	cs := "mongodb+srv://"
	if m.Username != "" && m.Password != "" {
		cs += m.Username + ":" + m.Password + "@"
	} else if m.Username != "" {
		cs += m.Username + "@"
	}

	cs += m.Host

	if m.Port != "" {
		cs += ":" + m.Port
	}

	if m.Args != "" {
		cs += "/?" + m.Args
	}

	m.ConnectionString = cs
}

func (m *MongoDB) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(m.ConnectionString))
	if err != nil {
		return fmt.Errorf("error connecting to mongo: %w", err)
	}
	err = client.Ping(ctx, nil)
	if err != nil {
		return fmt.Errorf("error pinging mongo: %w", err)
	}
	return nil
}

func (m *MongoDB) Connect() (*mongo.Client, error) {
	if m.ConnectionString == "" {
		m.GenerateConnectionString()
	}

	if err := m.Ping(); err != nil {
		return nil, fmt.Errorf("error pinging mongo: %w", err)
	}
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)

	opts := options.Client().ApplyURI(m.ConnectionString).SetServerAPIOptions(serverAPI)

	if m.AppName != "" {
		opts.SetAppName(m.AppName)
	}

	client, err := mongo.Connect(context.Background(), opts)
	if err != nil {
		return nil, fmt.Errorf("error connecting to mongo: %w", err)
	}
	return client, nil
}
