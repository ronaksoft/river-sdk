package main

import (
	"encoding/json"
	"fmt"
	log "git.ronaksoftware.com/ronak/toolbox/logger"
	"io/ioutil"
	"os"

	"git.ronaksoftware.com/ronak/riversdk/cmd/cli-supernumerary/config"
	"git.ronaksoftware.com/ronak/riversdk/cmd/cli-supernumerary/pkg/actor"
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
	loadCachedActors()
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
