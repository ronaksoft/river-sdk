package main

import (
	"git.ronaksoftware.com/ronak/riversdk/cmd/cli-supernumerary/config"
	"git.ronaksoftware.com/ronak/riversdk/cmd/cli-supernumerary/pkg/scenario"
	"git.ronaksoftware.com/ronak/riversdk/cmd/cli-supernumerary/pkg/shared"
	ronak "git.ronaksoftware.com/ronak/toolbox"
	log "git.ronaksoftware.com/ronak/toolbox/logger"
	"go.uber.org/zap"
	"os"
	"time"
)

var (
	_Log log.Logger
)

func init() {}

func main() {
	_Log = log.NewConsoleLogger()
	_Log.SetLevel(log.Debug)
	ronak.SetLogger(_Log)
	// supernumerary.SetLogger(_Log)
	scenario.SetLogger(_Log)

	cfg := &NodeConfig{
		BundleID:   os.Getenv(config.EnvBundleID),
		InstanceID: ronak.RandomID(24),
		NatsURL:    os.Getenv(config.EnvNatsUrl),
		RedisPass:  os.Getenv(config.EnvRedisPass),
		RedisHost:  os.Getenv(config.EnvRedisDSN),
	}
	_Log.Info("Config",
		zap.String("BundleID", cfg.BundleID),
		zap.String("InstanceID", cfg.InstanceID),
		zap.String("NatsURL", cfg.NatsURL),
	)

	shared.InitMetrics(cfg.BundleID, cfg.InstanceID)

	// Run metrics
	go shared.Metrics.Run(2374)

	for {
		_, err := NewNode(cfg)
		if err != nil {
			_Log.Warn("Error On NewNode",
				zap.Error(err),
			)
			time.Sleep(time.Second * 5)
		} else {
			break
		}
	}

	// wait forever
	select {}
}
