package main

import "git.ronaksoftware.com/ronak/riversdk/supernumerary/config"

var (
	_node *Node
)

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
