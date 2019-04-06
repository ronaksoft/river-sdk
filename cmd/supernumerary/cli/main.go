package main

import (
	"git.ronaksoftware.com/ronak/riversdk/logs"
	"github.com/nats-io/go-nats"
	"gopkg.in/abiosoft/ishell.v2"
)

var (
	_Shell *ishell.Shell
	_NATS  *nats.Conn
)

func init() {

	// Initialize Shell
	_Shell = ishell.New()
	_Shell.Println("=================================")
	_Shell.Println("## River Supernumerary Console ##")
	_Shell.Println("=================================")

	_Shell.AddCmd(cmdStart)
	_Shell.AddCmd(cmdStop)
	_Shell.AddCmd(cmdCreateAuthKey)
	_Shell.AddCmd(cmdLogin)
	_Shell.AddCmd(cmdRegister)
	_Shell.AddCmd(cmdSetTicker)

	logs.SetLogLevel(0) // DBG: -1, INF: 0, WRN: 1, ERR: 2
}

func main() {

	for {
		_Shell.Print("NATS URL (nats://localhost:4222):")
		natsURL := _Shell.ReadLine()
		if natsURL == "" {
			natsURL = "nats://localhost:4222"
		}
		nats, err := nats.Connect(natsURL)
		if err != nil {
			_Shell.Println("Error : " + err.Error())
		} else {
			_NATS = nats
			break
		}
	}

	_Shell.Run()
}
