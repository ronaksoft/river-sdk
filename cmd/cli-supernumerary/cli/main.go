package main

import (
	"github.com/nats-io/go-nats"
	"go.uber.org/zap"
	"gopkg.in/abiosoft/ishell.v2"
)

var (
	_Shell *ishell.Shell
	_NATS  *nats.Conn
	_Log                     *zap.Logger
	_LogLevel                zap.AtomicLevel
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
}

func main() {
	_LogLevel = zap.NewAtomicLevelAt(zap.DebugLevel)
	cfg := zap.NewProductionConfig()
	cfg.Level = _LogLevel
	_Log, _ = cfg.Build()

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
