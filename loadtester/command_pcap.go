package main

import (
	"fmt"
	"os"

	"git.ronaksoftware.com/ronak/riversdk/loadtester/shared"

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

		c.Print("pcap file path:")
		filePath := c.ReadLine()
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			log.LOG_Error("Error", zap.Error(err))
			return
		}
		res, err := pcap_parser.Parse(filePath)
		if err != nil {
			log.LOG_Error("Error", zap.Error(err))
			return
		}

		rpt := report.NewPcapReport()
		feedErrs := 0
		for r := range res {
			err := rpt.Feed(r)
			if err != nil {
				feedErrs++
				flow := fmt.Sprintf("%v:%d-->%v:%d", r.SrcIP, r.SrcPort, r.DstIP, r.DstPort)
				_, ok := shared.GetCachedActorByAuthID(r.Message.AuthID)
				fmt.Printf("Feed() AuthID : %d \t Exist : %v \t %s \t %s \n", r.Message.AuthID, ok, flow, err.Error())
			}
		}
		fmt.Println(rpt.String())
		fmt.Println("Feed() Errors : ", feedErrs)

	},
}

var cmdParse_wsutil = &ishell.Cmd{
	Name: "Parse_wsutil",
	Func: func(c *ishell.Context) {

		c.Print("pcap file path:")
		filePath := c.ReadLine()
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			log.LOG_Error("Error", zap.Error(err))
			return
		}

		res, err := pcap_parser.Parse(filePath)
		if err != nil {
			log.LOG_Error("Error", zap.Error(err))
			return
		}

		rpt := report.NewPcapReport()
		feedErrs := 0
		for r := range res {
			err := rpt.Feed(r)
			if err != nil {
				feedErrs++
				flow := fmt.Sprintf("%v:%d-->%v:%d", r.SrcIP, r.SrcPort, r.DstIP, r.DstPort)
				_, ok := shared.GetCachedActorByAuthID(r.Message.AuthID)
				fmt.Printf("Feed() AuthID : %d \t Exist : %v \t %s \t %s \n", r.Message.AuthID, ok, flow, err.Error())
			}
		}
		fmt.Println(rpt.String())
		fmt.Println("Feed() Errors : ", feedErrs)

	},
}

func init() {
	cmdPcap.AddCmd(cmdParse)
	cmdPcap.AddCmd(cmdParse_wsutil)

}
