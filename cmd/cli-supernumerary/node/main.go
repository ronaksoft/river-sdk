package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"git.ronaksoftware.com/ronak/riversdk/cmd/cli-loadtester/actor"
	"git.ronaksoftware.com/ronak/riversdk/cmd/cli-loadtester/shared"
	"git.ronaksoftware.com/ronak/riversdk/cmd/cli-supernumerary/config"
	"go.uber.org/zap"
)

var (
	_node     *Node
	_Log      *zap.Logger
	_LogLevel zap.AtomicLevel
)

func init() {
	if _, err := os.Stat("_cache/"); os.IsNotExist(err) {
		os.Mkdir("_cache/", os.ModePerm)
	}
	loadCachedActors()
}

func main() {
	_LogLevel = zap.NewAtomicLevelAt(zap.DebugLevel)
	logConfig := zap.NewProductionConfig()
	logConfig.Level = _LogLevel
	_Log, _ = logConfig.Build()

	cfg, err := config.NewNodeConfig()
	if err != nil {
		panic(err)
	}
	_Log.Info("Config",
		zap.String("BundleID", cfg.BundleID),
		zap.String("InstanceID", cfg.InstanceID),
		zap.String("NatsURL", cfg.NatsURL),
		zap.Int64("StartPhone", cfg.StartPhone),
		zap.Int64("EndPhone", cfg.EndPhone),
	)

	shared.InitMetrics(cfg.BundleID, cfg.InstanceID)
	// Run metrics
	go shared.Metrics.Run(2374)

	n, err := NewNode(cfg)
	if err != nil {
		panic(err)
	}
	_node = n

	// wait forever
	select {}
}

func loadCachedActors() {
	fmt.Printf("\n\n Start Loading Cached Actors ... \n\n")
	files, err := ioutil.ReadDir("_cache/")
	if err != nil {
		_Log.Error("Fialed to load cached actors LoadCachedActors()", zap.Error(err))
		return
	}

	counter := 0
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		jsonBytes, err := ioutil.ReadFile("_cache/" + f.Name())
		if err != nil {
			fmt.Println("Failed to load actor Filename : ", f.Name())
			continue
		}
		act := new(actor.Actor)
		err = json.Unmarshal(jsonBytes, act)
		if err == nil {
			shared.CacheActor(act)
			counter++
		} else {
			fmt.Println("Failed to Unmarshal actor Filename :", f.Name())
		}
	}
	fmt.Printf("\n Successfully loaded %d actors \n\n", counter)
}
