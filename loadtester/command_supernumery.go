package main

import (
	"git.ronaksoftware.com/ronak/riversdk/loadtester/controller"
	"git.ronaksoftware.com/ronak/riversdk/loadtester/supernumerary"
	"git.ronaksoftware.com/ronak/riversdk/logs"
	"go.uber.org/zap"
	ishell "gopkg.in/abiosoft/ishell.v2"
)

var (
	su *supernumerary.Supernumerary
)

var cmdSupernumerary = &ishell.Cmd{
	Name: "Supernumerary",
}

func init() {
	cmdSupernumerary.AddCmd(suStart)
	cmdSupernumerary.AddCmd(suStop)
	cmdSupernumerary.AddCmd(suCreateAuthKey)
	cmdSupernumerary.AddCmd(suRegister)
	cmdSupernumerary.AddCmd(suLogin)
	cmdSupernumerary.AddCmd(suSetTicker)
}

var suStart = &ishell.Cmd{
	Name: "Start",
	Func: func(c *ishell.Context) {

		from := fnStartPhone(c)
		to := fnEndPhone(c)

		logs.Info("Disabling packet logger ...")
		controller.StopLogginPackets()
		logs.Info("Disabling packet logger ... Done")

		logs.Info("Initializing Supernumerary ...")
		s, err := supernumerary.NewSupernumerary(from, to)
		logs.Info("Initializing Supernumerary ... Done")

		if err != nil {
			logs.Error("Failed to initialize ", zap.Error(err))
			return
		}
		su = s
	},
}

var suStop = &ishell.Cmd{
	Name: "Stop",
	Func: func(c *ishell.Context) {

		if su == nil {
			logs.Warn("Supernumerary not started")
			return
		}

		logs.Info("Stopping Supernumerary ...")
		su.Stop()
		logs.Info("Stopping Supernumerary ... Done")
		su = nil
	},
}

var suCreateAuthKey = &ishell.Cmd{
	Name: "CreateAuthKey",
	Func: func(c *ishell.Context) {

		if su == nil {
			logs.Warn("Supernumerary not started")
			return
		}

		logs.Info("CreateAuthKey ...")
		su.CreateAuthKey()
		logs.Info("CreateAuthKey ... Done")
	},
}

var suRegister = &ishell.Cmd{
	Name: "Register",
	Func: func(c *ishell.Context) {

		if su == nil {
			logs.Warn("Supernumerary not started")
			return
		}

		logs.Info("Register ...")
		su.Register()
		logs.Info("Register ... Done")
	},
}

var suLogin = &ishell.Cmd{
	Name: "Login",
	Func: func(c *ishell.Context) {

		if su == nil {
			logs.Warn("Supernumerary not started")
			return
		}

		logs.Info("Login ...")
		su.Login()
		logs.Info("Login ... Done")
	},
}

var suSetTicker = &ishell.Cmd{
	Name: "SetTicker",
	Func: func(c *ishell.Context) {

		if su == nil {
			logs.Warn("Supernumerary not started")
			return
		}

		duration := fnGetDuration(c)
		tickerAction := supernumerary.TickerAction(fnGetTickerAction(c))
		logs.Info("SetTickerApplier ...")
		su.SetTickerApplier(duration, tickerAction)
		logs.Info("SetTickerApplier ... Done")

	},
}
