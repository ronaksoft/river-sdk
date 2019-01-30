package main

import (
	"fmt"
	"sync"
	"time"

	"git.ronaksoftware.com/ronak/riversdk/loadtester/scenario"
	"git.ronaksoftware.com/ronak/riversdk/log"
	ishell "gopkg.in/abiosoft/ishell.v2"
)

var cmdPrint = &ishell.Cmd{
	Name: "Print",
	Func: func(c *ishell.Context) {
		if _Reporter != nil {
			fnClearScreeen()
			_Reporter.Print()
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
		log.LOG_Info(fmt.Sprintf(_Reporter.String()))

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
		log.LOG_Info(fmt.Sprintf(_Reporter.String()))

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
		log.LOG_Info(fmt.Sprintf(_Reporter.String()))

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
		log.LOG_Info(fmt.Sprintf(_Reporter.String()))

	},
}

var cmdSendMessage = &ishell.Cmd{
	Name: "SendMessage",
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
			JobQueue <- Job{PhoneNo: phoneNo, Wait: wg, Scenario: scenario.NewSendMessage(true)}

		}
		wg.Wait()
		elapsed := time.Since(sw)

		fnClearScreeen()
		log.LOG_Info(fmt.Sprintf("Execution Time : %v", elapsed))
		log.LOG_Info(fmt.Sprintf(_Reporter.String()))

	},
}
