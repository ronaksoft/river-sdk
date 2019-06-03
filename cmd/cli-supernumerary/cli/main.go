package main

import (
	"encoding/json"
	"fmt"
	"git.ronaksoftware.com/ronak/riversdk/cmd/cli-supernumerary/config"
	ronak "git.ronaksoftware.com/ronak/toolbox"
	"github.com/nats-io/go-nats"
	"go.uber.org/zap"
	"gopkg.in/abiosoft/ishell.v2"
	"io/ioutil"
	"os"
	"sync"
)

var (
	_Shell     *ishell.Shell
	_NATS      *nats.Conn
	_Nodes     map[string]struct{}
	_NodesLock sync.RWMutex
	_Log       *zap.Logger
	_LogLevel  zap.AtomicLevel
)

func init() {
	_Nodes = make(map[string]struct{})

	// Initialize Shell
	_Shell = ishell.New()
	_Shell.Println("=================================")
	_Shell.Println("## River Supernumerary Console ##")
	_Shell.Println("=================================")

	_Shell.AddCmd(cmdListNodes)
	_Shell.AddCmd(cmdUpdatePhoneRange)
	_Shell.AddCmd(cmdStart)
	_Shell.AddCmd(cmdStop)
	_Shell.AddCmd(cmdCreateAuthKey)
	_Shell.AddCmd(cmdLogin)
	_Shell.AddCmd(cmdRegister)
	_Shell.AddCmd(cmdSetTicker)
	_Shell.AddCmd(cmdResetAuth)
}

func main() {
	_LogLevel = zap.NewAtomicLevelAt(zap.DebugLevel)
	cfg := zap.NewProductionConfig()
	cfg.Level = _LogLevel
	_Log, _ = cfg.Build()

	defaultNatsUrl, _ := ioutil.ReadFile(".river-nats")
	if defaultNatsUrl == nil {
		defaultNatsUrl = ronak.StrToByte("nats://localhost:4222")
	}
	_Shell.Print(fmt.Sprintf("NATS URL (%s):", defaultNatsUrl))
	natsURL := _Shell.ReadLine()
	if natsURL == "" {
		natsURL = ronak.ByteToStr(defaultNatsUrl)
	} else {
		_ = ioutil.WriteFile(".river-nats", ronak.StrToByte(natsURL), os.ModePerm)
	}

	if natsClient, err := nats.Connect(natsURL); err != nil {
		_Shell.Println("Error : " + err.Error())
	} else {
		_NATS = natsClient
	}

	_, err := _NATS.Subscribe(config.SubjectCommander, func(msg *nats.Msg) {
		cmd := config.NodeRegisterCmd{}
		err := json.Unmarshal(msg.Data, &cmd)
		if err != nil {
			_Log.Warn("Error On Received NATS Message",
				zap.Error(err),
			)
			return
		}
		_NodesLock.Lock()
		_Nodes[cmd.InstanceID] = struct{}{}
		_NodesLock.Unlock()
	})
	if err != nil {
		_Log.Fatal(err.Error())
	}
	_Shell.Run()
}
