package main

import (
	"git.ronaksoftware.com/ronak/riversdk/log"
	ishell "gopkg.in/abiosoft/ishell.v2"
)

var CLI = &ishell.Cmd{
	Name: "CLI",
}

var SetLogLevel = &ishell.Cmd{
	Name: "SetLogLevel",
	Func: func(c *ishell.Context) {
		idx := c.MultiChoice([]string{
			"Debug", "Info", "Warn", "Error",
		}, "Level")
		log.SetLogLevel(idx - 1)
	},
}
var SetReporter = &ishell.Cmd{
	Name: "SetReporter",
	Func: func(c *ishell.Context) {
		idx := c.MultiChoice([]string{
			"Logs", "Statistics",
		}, "Mode")
		_Reporter.SetIsActive(idx > 0)
	},
}

func init() {
	CLI.AddCmd(SetReporter)
	CLI.AddCmd(SetLogLevel)
}
