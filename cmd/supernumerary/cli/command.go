package main

import (
	"encoding/json"

	"git.ronaksoftware.com/ronak/riversdk/loadtester/supernumerary"
	"git.ronaksoftware.com/ronak/riversdk/logs"
	"git.ronaksoftware.com/ronak/riversdk/supernumerary/config"
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
		logs.Info("Publishing Start ...")
		err := _NATS.Publish(config.SUBJECT_START, data)
		if err != nil {
			logs.Error("Error Start", zap.Error(err))
		}

		logs.Info("Publishing Start ... Done")
	},
}

var cmdStop = &ishell.Cmd{
	Name: "Stop",
	Func: func(c *ishell.Context) {
		logs.Info("Publishing Stop ...")
		err := _NATS.Publish(config.SUBJECT_STOP, []byte(config.SUBJECT_STOP))
		if err != nil {
			logs.Error("Error Stop", zap.Error(err))
		}

		logs.Info("Publishing Stop ... Done")
	},
}

var cmdCreateAuthKey = &ishell.Cmd{
	Name: "CreateAuthKey",
	Func: func(c *ishell.Context) {
		logs.Info("Publishing CreateAuthKey ...")
		err := _NATS.Publish(config.SUBJECT_CREATEAUTHKEY, []byte(config.SUBJECT_CREATEAUTHKEY))
		if err != nil {
			logs.Error("Error CreateAuthKey", zap.Error(err))
		}
		logs.Info("Publishing CreateAuthKey ... Done")
	},
}

var cmdRegister = &ishell.Cmd{
	Name: "Register",
	Func: func(c *ishell.Context) {
		logs.Info("Publishing Register ...")
		err := _NATS.Publish(config.SUBJECT_RIGISTER, []byte(config.SUBJECT_RIGISTER))
		if err != nil {
			logs.Error("Error Register", zap.Error(err))
		}
		logs.Info("Publishing Register ... Done")
	},
}

var cmdLogin = &ishell.Cmd{
	Name: "Login",
	Func: func(c *ishell.Context) {
		logs.Info("Publishing Login ...")
		err := _NATS.Publish(config.SUBJECT_LOGIN, []byte(config.SUBJECT_LOGIN))
		if err != nil {
			logs.Error("Error Login", zap.Error(err))
		}

		logs.Info("Publishing Login ... Done")
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
			logs.Error("Error Ticker", zap.Error(err))
		}

		logs.Info("Publishing Ticker ... Done")
	},
}
