package main

import (
	"time"

	"git.ronaksoftware.com/ronak/riversdk/loadtester/scenario"

	"go.uber.org/zap"

	"git.ronaksoftware.com/ronak/riversdk/loadtester/actor"
	"git.ronaksoftware.com/ronak/riversdk/logs"

	"git.ronaksoftware.com/ronak/riversdk/loadtester/shared"

	ishell "gopkg.in/abiosoft/ishell.v2"
)

var _clientActer shared.Acter
var _clientCreatedOn time.Time

var cmdClient = &ishell.Cmd{
	Name: "Client",
}

var cmdClientStart = &ishell.Cmd{
	Name: "Start",
	Func: func(c *ishell.Context) {

		phoneNo := fnGetPhone(c)
		// clear
		fnClearScreeen()
		fnClearReports()

		// TODO :: register or login
		var err error
		_clientCreatedOn = time.Now()
		_clientActer, err = actor.NewActor(phoneNo)
		if err != nil {
			logs.Error("Faile to create client actor", zap.Error(err))
		}

		// create authKey if actor does not have authID
		if _clientActer.GetAuthID() == 0 {
			sw := scenario.NewCreateAuthKey(false)
			success := scenario.Play(_clientActer, sw)
			if !success {
				logs.Error("Faile at pre requested scenario CreateAuthKey")
				return
			}
			_clientActer.Save()
		}
		// login if actor does not have userID
		if _clientActer.GetUserID() == 0 {
			sw := scenario.NewLogin(false)
			success := scenario.Play(_clientActer, sw)
			if !success {
				logs.Error("Faile at pre requested scenario Login")
				return
			}
			_clientActer.Save()
		}

		sw := scenario.NewAuthRecall(false)
		success := scenario.Play(_clientActer, sw)
		if !success {
			logs.Error("Faile at pre requested scenario AuthRecall")
			return
		}
		_Reporter.SetIsActive(false)
		logs.SetLogLevel(-1)
		logs.Info("Client Actor Started log level changed to debug")
	},
}

var cmdClientStop = &ishell.Cmd{
	Name: "Stop",
	Func: func(c *ishell.Context) {
		if _clientActer == nil {
			logs.Error("Client Actor is not started")
			return
		}
		_clientActer.Stop()
		fnPrintReports(time.Since(_clientCreatedOn))
		_clientActer = nil
	},
}

func init() {
	cmdClient.AddCmd(cmdClientStart)
	cmdClient.AddCmd(cmdClientStop)
}
