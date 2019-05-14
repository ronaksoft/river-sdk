package config

import (
	"errors"
	ronak "git.ronaksoftware.com/ronak/toolbox"
	"os"
)

// NodeConfig environment variables  to configs each docker container
type NodeConfig struct {
	BundleID   string
	InstanceID string
	NatsURL    string
}

// NewNodeConfig reads environment variables
func NewNodeConfig() (*NodeConfig, error) {
	cfg := &NodeConfig{
		BundleID:   os.Getenv(EnvBundleID),
		InstanceID: ronak.RandomID(24),
		NatsURL:    os.Getenv(EnvNatsUrl),
	}

	if cfg.NatsURL == "" {
		return nil, errors.New("invalid nats endpoint")
	}

	return cfg, nil
}
