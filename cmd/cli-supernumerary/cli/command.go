package main

import (
	"encoding/json"
	"fmt"
	ronak "git.ronaksoftware.com/ronak/toolbox"
	"time"

	"git.ronaksoftware.com/ronak/riversdk/cmd/cli-loadtester/supernumerary"
	"git.ronaksoftware.com/ronak/riversdk/cmd/cli-supernumerary/config"
	"go.uber.org/zap"
	"gopkg.in/abiosoft/ishell.v2"
)

var cmdListNodes = &ishell.Cmd{
	Name: "ListNodes",
	Func: func(c *ishell.Context) {
		_NodesLock.RLock()
		c.Println("Total Nodes:", len(_Nodes))
		for instanceID := range _Nodes {
			c.Println(instanceID)
		}
		_NodesLock.RUnlock()
	},
}

var cmdUpdatePhoneRange = &ishell.Cmd{
	Name: "UpdatePhoneRange",
	Func: func(c *ishell.Context) {
		c.Print("Total Phones:")
		totalPhone := ronak.StrToInt32(c.ReadLine())

		_NodesLock.RLock()
		totalNodes := int32(len(_Nodes))
		instanceIDs := make([]string, 0, totalNodes)
		for instanceID := range _Nodes {
			instanceIDs = append(instanceIDs, instanceID)
		}
		_NodesLock.RUnlock()

		phoneRange := totalPhone / totalNodes
		rangeRemaining := totalPhone % totalNodes
		idx := int32(0)
		for instanceID := range instanceIDs {
			startPhone := idx * phoneRange
			endPhone := startPhone + phoneRange
			if idx == totalNodes-1 {
				endPhone += rangeRemaining
			}
			cfg := config.PhoneRangeCfg{
				StartPhone: int64(startPhone),
				EndPhone:   int64(endPhone),
			}
			d, _ := json.Marshal(cfg)
			_, err := _NATS.Request(fmt.Sprintf("%s.%s", instanceID, config.SubjectPhoneRange), d, time.Second*10)
			_Log.Warn("Error On UpdatePhoneRange",
				zap.Error(err),
			)
			idx++
		}

	},
}
var cmdStart = &ishell.Cmd{
	Name: "Start",
	Func: func(c *ishell.Context) {
		serverURL := fnGetServerURL(c)
		fileServerURL := fnGetFileServerURL(c)
		timeout := fnGetTimeout(c)
		cfg := config.StartCfg{
			ServerURL:     serverURL,
			FileServerURL: fileServerURL,
			Timeout:       timeout,
		}

		data, _ := json.Marshal(cfg)
		_Log.Info("Publishing Start ...")
		err := _NATS.Publish(config.SubjectStart, data)
		if err != nil {
			_Log.Error("Error Start", zap.Error(err))
		}

		_Log.Info("Publishing Start ... Done")
	},
}

var cmdStop = &ishell.Cmd{
	Name: "Stop",
	Func: func(c *ishell.Context) {
		_Log.Info("Publishing Stop ...")
		err := _NATS.Publish(config.SubjectStop, []byte(config.SubjectStop))
		if err != nil {
			_Log.Error("Error Stop", zap.Error(err))
		}

		_Log.Info("Publishing Stop ... Done")
	},
}

var cmdCreateAuthKey = &ishell.Cmd{
	Name: "CreateAuthKey",
	Func: func(c *ishell.Context) {
		_Log.Info("Publishing CreateAuthKey ...")
		err := _NATS.Publish(config.SubjectCreateAuthKey, []byte(config.SubjectCreateAuthKey))
		if err != nil {
			_Log.Error("Error CreateAuthKey", zap.Error(err))
		}
		_Log.Info("Publishing CreateAuthKey ... Done")
	},
}

var cmdRegister = &ishell.Cmd{
	Name: "Register",
	Func: func(c *ishell.Context) {
		_Log.Info("Publishing Register ...")
		err := _NATS.Publish(config.SubjectRegister, []byte(config.SubjectRegister))
		if err != nil {
			_Log.Error("Error Register", zap.Error(err))
		}
		_Log.Info("Publishing Register ... Done")
	},
}

var cmdLogin = &ishell.Cmd{
	Name: "Login",
	Func: func(c *ishell.Context) {
		_Log.Info("Publishing Login ...")
		err := _NATS.Publish(config.SubjectLogin, []byte(config.SubjectLogin))
		if err != nil {
			_Log.Error("Error Login", zap.Error(err))
		}

		_Log.Info("Publishing Login ... Done")
	},
}

var cmdSetTicker = &ishell.Cmd{
	Name: "SetTicker",
	Func: func(c *ishell.Context) {
		duration := fnGetDuration(c)
		tickerAction := fnGetTickerAction(c)
		cfg := config.TickerCfg{
			Action:   supernumerary.TickerAction(tickerAction),
			Duration: duration,
		}
		data, _ := json.Marshal(cfg)
		err := _NATS.Publish(config.SubjectTicker, data)
		if err != nil {
			_Log.Error("Error Ticker", zap.Error(err))
		}

		_Log.Info("Publishing Ticker ... Done")
	},
}
