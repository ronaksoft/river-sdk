package main

import (
	"git.ronaksoftware.com/ronak/riversdk/loadtester/pcap_parser"
	"git.ronaksoftware.com/ronak/riversdk/loadtester/report"
	"git.ronaksoftware.com/ronak/riversdk/log"
	"go.uber.org/zap"
	ishell "gopkg.in/abiosoft/ishell.v2"
)

var cmdPcap = &ishell.Cmd{
	Name: "Pcap",
}

var cmdParse = &ishell.Cmd{
	Name: "Parse",
	Func: func(c *ishell.Context) {
		for {
			c.Print("pcap file path:")
			filePath := c.ReadLine()

			res, err := pcap_parser.Parse(filePath)
			if err != nil {
				log.LOG_Error("Error", zap.Error(err))
			}

			rpt := report.NewPcapReport()
			for r := range res {
				rpt.Feed(r)
			}
			log.LOG_Info("Parse result \n\n" + rpt.String())
		}
	},
}

var cmdParse_wsutil = &ishell.Cmd{
	Name: "Parse_wsutil",
	Func: func(c *ishell.Context) {
		for {
			c.Print("pcap file path:")
			filePath := c.ReadLine()

			res, err := pcap_parser.Parse(filePath)
			if err != nil {
				log.LOG_Error("Error", zap.Error(err))
			}

			rpt := report.NewPcapReport()
			for r := range res {
				rpt.Feed(r)
			}
			log.LOG_Info("Parse result \n\n" + rpt.String())
		}
	},
}

func init() {
	cmdPcap.AddCmd(cmdParse)
	cmdPcap.AddCmd(cmdParse_wsutil)

}
