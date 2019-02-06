package main

import (
	"fmt"
	"sync"
	"time"

	"git.ronaksoftware.com/ronak/riversdk/loadtester/controller"
	"git.ronaksoftware.com/ronak/riversdk/loadtester/report"

	"git.ronaksoftware.com/ronak/riversdk/loadtester/scenario"
	"git.ronaksoftware.com/ronak/riversdk/loadtester/shared"
	"git.ronaksoftware.com/ronak/riversdk/log"
	ishell "gopkg.in/abiosoft/ishell.v2"
)

var cmdPrint = &ishell.Cmd{
	Name: "Print",
	Func: func(c *ishell.Context) {
		if _Reporter != nil {
			fnClearScreeen()

			fmt.Println(_Reporter.String())
			fmt.Printf("Failed Requests :\n%s", shared.PrintFailedRequest())
		}
	},
}

var cmdCreateAuthKey = &ishell.Cmd{
	Name: "CreateAuthKey",
	Func: func(c *ishell.Context) {

		startNo := fnStartPhone(c)
		endNo := fnEndPhone(c)
		fnClearScreeen()
		_Reporter.Clear()

		startDispatcher()
		defer stopDispatcher()

		phoneNo := ""
		wg := &sync.WaitGroup{}
		sw := time.Now()
		for startNo <= endNo {
			phoneNo = fmt.Sprintf("237400%07d", startNo)
			startNo++

			// Add To Queue
			wg.Add(1)
			JobQueue <- Job{PhoneNo: phoneNo, Wait: wg, Scenario: scenario.NewCreateAuthKey(true)}

		}
		wg.Wait()
		elapsed := time.Since(sw)

		fnClearScreeen()
		log.LOG_Info(fmt.Sprintf("Execution Time : %v", elapsed))

		fmt.Println(_Reporter.String())
		fmt.Printf("Failed Requests :\n%s", shared.PrintFailedRequest())

	},
}

var cmdRegister = &ishell.Cmd{
	Name: "Register",
	Func: func(c *ishell.Context) {

		startNo := fnStartPhone(c)
		endNo := fnEndPhone(c)
		fnClearScreeen()
		_Reporter.Clear()

		startDispatcher()
		defer stopDispatcher()

		phoneNo := ""
		wg := &sync.WaitGroup{}
		sw := time.Now()
		for startNo <= endNo {
			phoneNo = fmt.Sprintf("237400%07d", startNo)
			startNo++

			// Add To Queue
			wg.Add(1)
			JobQueue <- Job{PhoneNo: phoneNo, Wait: wg, Scenario: scenario.NewRegister(true)}

		}
		wg.Wait()
		elapsed := time.Since(sw)

		fnClearScreeen()
		log.LOG_Info(fmt.Sprintf("Execution Time : %v", elapsed))

		fmt.Println(_Reporter.String())
		fmt.Printf("Failed Requests :\n%s", shared.PrintFailedRequest())

	},
}

var cmdLogin = &ishell.Cmd{
	Name: "Login",
	Func: func(c *ishell.Context) {

		// get suffix start phoneNo
		// get suffix end phoneNo
		// start loging in :/

		startNo := fnStartPhone(c)
		endNo := fnEndPhone(c)
		fnClearScreeen()
		_Reporter.Clear()

		startDispatcher()
		defer stopDispatcher()

		phoneNo := ""
		wg := &sync.WaitGroup{}
		sw := time.Now()
		for startNo <= endNo {
			phoneNo = fmt.Sprintf("237400%07d", startNo)
			startNo++

			// Add To Queue
			wg.Add(1)
			JobQueue <- Job{PhoneNo: phoneNo, Wait: wg, Scenario: scenario.NewLogin(true)}
		}
		wg.Wait()
		elapsed := time.Since(sw)

		fnClearScreeen()
		log.LOG_Info(fmt.Sprintf("Execution Time : %v", elapsed))

		fmt.Println(_Reporter.String())
		fmt.Printf("Failed Requests :\n%s", shared.PrintFailedRequest())

	},
}

var cmdImportContact = &ishell.Cmd{
	Name: "ImportContact",
	Func: func(c *ishell.Context) {
		// get phone number to import
		// get suffix start phoneNo
		// get suffix end phoneNo
		// start importing contacts

		phoneNoToImportAsContact := fnGetPhone(c)
		startNo := fnStartPhone(c)
		endNo := fnEndPhone(c)
		fnClearScreeen()
		_Reporter.Clear()

		startDispatcher()
		defer stopDispatcher()

		phoneNo := ""
		wg := &sync.WaitGroup{}
		sw := time.Now()
		for startNo <= endNo {
			phoneNo = fmt.Sprintf("237400%07d", startNo)
			startNo++

			// Add To Queue
			wg.Add(1)
			JobQueue <- Job{PhoneNo: phoneNo, Wait: wg, Scenario: scenario.NewImportContact(true), PhoneListToImportAsContact: phoneNoToImportAsContact}

		}
		wg.Wait()
		elapsed := time.Since(sw)

		fnClearScreeen()
		log.LOG_Info(fmt.Sprintf("Execution Time : %v", elapsed))

		fmt.Println(_Reporter.String())
		fmt.Printf("Failed Requests :\n%s", shared.PrintFailedRequest())

	},
}

var cmdSendMessage = &ishell.Cmd{
	Name: "SendMessage",
	Func: func(c *ishell.Context) {

		startNo := fnStartPhone(c)
		endNo := fnEndPhone(c)
		fnClearScreeen()
		_Reporter.Clear()
		controller.ClearLoggedPackets()

		startDispatcher()
		defer stopDispatcher()

		phoneNo := ""
		wg := &sync.WaitGroup{}
		sw := time.Now()
		for startNo <= endNo {
			phoneNo = fmt.Sprintf("237400%07d", startNo)
			startNo++

			// Add To Queue
			wg.Add(1)
			JobQueue <- Job{PhoneNo: phoneNo, Wait: wg, Scenario: scenario.NewSendMessage(true)}

		}
		wg.Wait()
		elapsed := time.Since(sw)

		fnClearScreeen()
		log.LOG_Info(fmt.Sprintf("Execution Time : %v", elapsed))

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
		// for _, p := range responseList {
		// 	err := rpt.FeedPacket(p, true)
		// 	if err != nil {
		// 		fmt.Println("rpt.FeedPacket(p, reponses) :", err)
		// 	}
		// }
		fmt.Println(rpt.String())

	},
}
