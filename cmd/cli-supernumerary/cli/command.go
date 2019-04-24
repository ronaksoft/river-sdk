package main

import (
	"encoding/json"

	"git.ronaksoftware.com/ronak/riversdk/cmd/cli-loadtester/supernumerary"
	"git.ronaksoftware.com/ronak/riversdk/cmd/cli-supernumerary/config"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	"go.uber.org/zap"
	ishell "gopkg.in/abiosoft/ishell.v2"
)

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
		err := _NATS.Publish(config.SUBJECT_START, data)
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
		err := _NATS.Publish(config.SUBJECT_STOP, []byte(config.SUBJECT_STOP))
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
		err := _NATS.Publish(config.SUBJECT_CREATEAUTHKEY, []byte(config.SUBJECT_CREATEAUTHKEY))
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
		err := _NATS.Publish(config.SUBJECT_RIGISTER, []byte(config.SUBJECT_RIGISTER))
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
		err := _NATS.Publish(config.SUBJECT_LOGIN, []byte(config.SUBJECT_LOGIN))
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
		err := _NATS.Publish(config.SUBJECT_TICKER, data)
		if err != nil {
			_Log.Error("Error Ticker", zap.Error(err))
		}

		_Log.Info("Publishing Ticker ... Done")
	},
}
