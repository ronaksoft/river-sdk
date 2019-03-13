package main

import (
	"encoding/json"
	"fmt"

	"git.ronaksoftware.com/ronak/riversdk/logs"
	"git.ronaksoftware.com/ronak/riversdk/supernumerary/config"
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
		_NATS.Publish(config.SUBJECT_START, data)
		logs.Info("Publishing Start ... Done")
	},
}

var cmdStop = &ishell.Cmd{
	Name: "Stop",
	Func: func(c *ishell.Context) {
		logs.Info("Publishing Stop ...")
		_NATS.Publish(config.SUBJECT_STOP, []byte(config.SUBJECT_STOP))
		logs.Info("Publishing Stop ... Done")
	},
}

var cmdCreateAuthKey = &ishell.Cmd{
	Name: "CreateAuthKey",
	Func: func(c *ishell.Context) {
		logs.Info("Publishing CreateAuthKey ...")
		_NATS.Publish(config.SUBJECT_CREATEAUTHKEY, []byte(config.SUBJECT_CREATEAUTHKEY))
		logs.Info("Publishing CreateAuthKey ... Done")
	},
}

var cmdRegister = &ishell.Cmd{
	Name: "Register",
	Func: func(c *ishell.Context) {
		logs.Info("Publishing Register ...")
		_NATS.Publish(config.SUBJECT_RIGISTER, []byte(config.SUBJECT_RIGISTER))
		logs.Info("Publishing Register ... Done")
	},
}

var cmdLogin = &ishell.Cmd{
	Name: "Login",
	Func: func(c *ishell.Context) {
		logs.Info("Publishing Login ...")
		_NATS.Publish(config.SUBJECT_LOGIN, []byte(config.SUBJECT_LOGIN))
		logs.Info("Publishing Login ... Done")
	},
}

var cmdSetTicker = &ishell.Cmd{
	Name: "SetTicker",
	Func: func(c *ishell.Context) {
		duration := fnGetDuration(c)
		tickerAction := fnGetTickerAction(c)
		data := fmt.Sprintf("%d:%d", duration, tickerAction)

		logs.Info("Publishing Ticker ...")
		_NATS.Publish(config.SUBJECT_TICKER, []byte(data))
		logs.Info("Publishing Ticker ... Done")
	},
}
