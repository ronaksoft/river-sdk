package main

import (
	"strconv"
	"time"

	"git.ronaksoftware.com/ronak/riversdk/loadtester/shared"

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

var SetTimeout = &ishell.Cmd{
	Name: "SetTimeout",
	Func: func(c *ishell.Context) {
		var timeout int64
		for {
			c.Print("Timeout (second): ")
			entery, err := strconv.ParseInt(c.ReadLine(), 10, 64)
			if err == nil {
				timeout = entery
				break
			} else {
				c.Println(err.Error())
			}
		}

		shared.DefaultSendTimeout = time.Duration(timeout) * time.Second
		shared.DefaultTimeout = time.Duration(timeout) * time.Second

	},
}

var SetMaxWorker = &ishell.Cmd{
	Name: "SetMaxWorker",
	Func: func(c *ishell.Context) {
		var maxWorker int = 96
		for {
			c.Print("Max Worker: ")
			entery, err := strconv.ParseInt(c.ReadLine(), 10, 64)
			if err == nil {
				maxWorker = int(entery)
				break
			} else {
				c.Println(err.Error())
			}
		}

		shared.MaxWorker = maxWorker

	},
}

var SetMaxQueueBuffer = &ishell.Cmd{
	Name: "SetMaxQueueBuffer",
	Func: func(c *ishell.Context) {
		var maxQ int = 10000
		for {
			c.Print("Max QueueBuffer (10000): ")
			entery, err := strconv.ParseInt(c.ReadLine(), 10, 64)
			if err == nil {
				maxQ = int(entery)
				break
			} else {
				c.Println(err.Error())
			}
		}

		shared.MaxQueueBuffer = maxQ
	},
}

var SetServerURL = &ishell.Cmd{
	Name: "SetServerURL",
	Func: func(c *ishell.Context) {
		c.Print("URL (ws://test.river.im): ")
		url := c.ReadLine()
		shared.DefaultServerURL = url
	},
}

func init() {
	CLI.AddCmd(SetReporter)
	CLI.AddCmd(SetLogLevel)
	CLI.AddCmd(SetTimeout)
	CLI.AddCmd(SetMaxWorker)
	CLI.AddCmd(SetMaxQueueBuffer)
	CLI.AddCmd(SetServerURL)

}
