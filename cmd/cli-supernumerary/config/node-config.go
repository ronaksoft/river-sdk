package config

import (
	"errors"
	"os"
	"strconv"

	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	"go.uber.org/zap"
)

// NodeConfig environment variables  to configs each docker container
type NodeConfig struct {
	BundleID   string
	InstanceID string
	NatsURL    string
	StartPhone int64
	EndPhone   int64
}

// NewNodeConfig reads environment variables
func NewNodeConfig() (*NodeConfig, error) {
	cfg := &NodeConfig{
		BundleID:   os.Getenv(ENV_BOUNDLE_ID),
		InstanceID: os.Getenv(ENV_INSTANCE_ID),
		NatsURL:    os.Getenv(ENV_NATS_URL),
		StartPhone: 0,
		EndPhone:   0,
	}

	if cfg.NatsURL == "" {
		return nil, errors.New("invalid nats endpoint")
	}
	// startPhone
	s := os.Getenv(ENV_START_PHONE)
	sp, err := strconv.Atoi(s)
	if err != nil {
		logs.Error("NewNodeConfig() ENV_START_PHONE", zap.Error(err))
		return nil, err
	}
	cfg.StartPhone = int64(sp)

	// endPhone
	e := os.Getenv(ENV_END_PHONE)
	ep, err := strconv.Atoi(e)
	if err != nil {
		logs.Error("NewNodeConfig() ENV_END_PHONE", zap.Error(err))
		return nil, err
	}
	cfg.EndPhone = int64(ep)

	return cfg, nil
}
