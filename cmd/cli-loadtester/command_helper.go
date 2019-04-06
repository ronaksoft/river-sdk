package main

import (
	"fmt"
	"strconv"
	"time"

	"git.ronaksoftware.com/ronak/riversdk/loadtester/controller"
	"git.ronaksoftware.com/ronak/riversdk/loadtester/report"
	"git.ronaksoftware.com/ronak/riversdk/loadtester/shared"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	ishell "gopkg.in/abiosoft/ishell.v2"
)

func fnStartPhone(c *ishell.Context) int64 {
	var tmpNo int64
	for {

		c.Print("Start Phone: ")
		tmp, err := strconv.ParseInt(c.ReadLine(), 10, 64)
		if err == nil {
			if tmp < 10000000 {
				tmpNo = tmp
				break
			}
			c.Println("max 7 digit allowed")
		} else {
			c.Println(err.Error())
		}
	}
	return tmpNo
}

func fnEndPhone(c *ishell.Context) int64 {
	var tmpNo int64
	for {

		c.Print("End Phone: ")
		tmp, err := strconv.ParseInt(c.ReadLine(), 10, 64)
		if err == nil {
			if tmp < 10000000 {
				tmpNo = tmp
				break
			}
			c.Println("max 7 digit allowed")
		} else {
			c.Println(err.Error())
		}
	}
	return tmpNo
}

func fnGetPhone(c *ishell.Context) string {
	var tmpNo string
	for {

		c.Print("Phone: ")
		tmp, err := strconv.ParseInt(c.ReadLine(), 10, 32)
		if err == nil {
			tmpNo = strconv.Itoa(int(tmp))
			break
		} else {
			c.Println(err.Error())
		}
	}
	return tmpNo
}

func fnClearScreeen() {
	fmt.Println("\033[H\033[2J") // clear screen
}

func fnClearReports() {
	_Reporter.Clear()
	controller.ClearLoggedPackets()
}

func fnPrintReports(elapsed time.Duration) {
	fnClearScreeen()
	logs.Info(fmt.Sprintf("Execution Time : %v", elapsed))

	fmt.Println(_Reporter.String())
	fmt.Printf("Failed Requests :\n%s", shared.PrintFailedRequest())

	rpt := report.NewPcapReport()
	requsetList := controller.GetLoggedSentPackets()
	for _, p := range requsetList {
		err := rpt.FeedPacket(p, false)
		if err != nil {
			fmt.Println("rpt.FeedPacket(p, requests) :", err)
		}
	}

	responseList := controller.GetLoggedReceivedPackets()
	n := len(responseList)
	for i := 0; i < n; i++ {
		err := rpt.FeedPacket(responseList[i], true)
		if err != nil {
			fmt.Println("rpt.FeedPacket(p, reponses) :", err)
		}
	}
	fmt.Println(rpt.String())
}

func fnGetDuration(c *ishell.Context) time.Duration {
	var tmpNo time.Duration
	for {
		c.Print("Duration (second): ")
		tmp, err := strconv.ParseInt(c.ReadLine(), 10, 32)
		if err == nil {
			tmpNo = time.Duration(tmp)
			break
		} else {
			c.Println(err.Error())
		}
	}
	return tmpNo * time.Second
}

func fnGetTickerAction(c *ishell.Context) int {
	var action int
	for {
		c.Print("Actions (SendMsg = 1 , SendFile = 2) : ")
		tmp, err := strconv.ParseInt(c.ReadLine(), 10, 32)
		if err == nil {
			if tmp > 0 && tmp < 3 {
				action = int(tmp)
				break
			}
		}
	}
	return action
}
