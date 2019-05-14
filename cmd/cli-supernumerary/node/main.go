package main

import (
	log "git.ronaksoftware.com/ronak/toolbox/logger"
	"os"

	"git.ronaksoftware.com/ronak/riversdk/cmd/cli-supernumerary/config"
	"git.ronaksoftware.com/ronak/riversdk/cmd/cli-supernumerary/pkg/shared"
	"go.uber.org/zap"
)

var (
	_Log log.Logger
)

func init() {
	if _, err := os.Stat("_cache/"); os.IsNotExist(err) {
		os.Mkdir("_cache/", os.ModePerm)
	}
}

func main() {
	_Log = log.NewConsoleLogger()
	_Log.SetLevel(log.Debug)

	cfg, err := config.NewNodeConfig()
	if err != nil {
		panic(err)
	}
	_Log.Info("Config",
		zap.String("BundleID", cfg.BundleID),
		zap.String("InstanceID", cfg.InstanceID),
		zap.String("NatsURL", cfg.NatsURL),
	)

	shared.InitMetrics(cfg.BundleID, cfg.InstanceID)

	// Run metrics
	go shared.Metrics.Run(2374)

	_, err = NewNode(cfg)
	if err != nil {
		panic(err)
	}

	// wait forever
	select {}
}
