package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"git.ronaksoftware.com/ronak/riversdk/loadtester/actor"
	"git.ronaksoftware.com/ronak/riversdk/loadtester/shared"
	"git.ronaksoftware.com/ronak/riversdk/logs"
	"git.ronaksoftware.com/ronak/riversdk/supernumerary/config"
	"go.uber.org/zap"
)

var (
	_node *Node
)

func init() {
	if _, err := os.Stat("_cache/"); os.IsNotExist(err) {
		os.Mkdir("_cache/", os.ModePerm)
	}
	loadCachedActors()
}

func main() {
	cfg, err := config.NewNodeConfig()
	if err != nil {
		panic(err)
	}
	n, err := NewNode(cfg)
	if err != nil {
		panic(err)
	}
	_node = n

	//wait forever
	select {}
}

func loadCachedActors() {

	fmt.Printf("\n\n Start Loading Cached Actors ... \n\n")

	files, err := ioutil.ReadDir("_cache/")
	if err != nil {
		logs.Error("Fialed to load cached actors LoadCachedActors()", zap.Error(err))
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
