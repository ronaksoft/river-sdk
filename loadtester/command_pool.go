package main

import (
	"fmt"
	"sync"
	"time"

	"git.ronaksoftware.com/ronak/riversdk/loadtester/scenario"
	"git.ronaksoftware.com/ronak/riversdk/log"
	ishell "gopkg.in/abiosoft/ishell.v2"
)

var cmdRegisterByPool = &ishell.Cmd{
	Name: "RegisterByPool",
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
		if !_Reporter.IsActive() {
			fnClearScreeen()
			log.LOG_Info(fmt.Sprintf("Execution Time : %v", elapsed))
			log.LOG_Info(fmt.Sprintf(_Reporter.String()))
		}

	},
}

var cmdLoginByPool = &ishell.Cmd{
	Name: "LoginByPool",
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
		if !_Reporter.IsActive() {
			fnClearScreeen()
			log.LOG_Info(fmt.Sprintf("Execution Time : %v", elapsed))
			log.LOG_Info(fmt.Sprintf(_Reporter.String()))
		}

	},
}

var cmdImportContactByPool = &ishell.Cmd{
	Name: "ImportContactByPool",
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
		if !_Reporter.IsActive() {
			fnClearScreeen()
			log.LOG_Info(fmt.Sprintf("Execution Time : %v", elapsed))
			log.LOG_Info(fmt.Sprintf(_Reporter.String()))
		}
	},
}

var cmdSendMessageByPool = &ishell.Cmd{
	Name: "SendMessageByPool",
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
		if !_Reporter.IsActive() {
			fnClearScreeen()
			log.LOG_Info(fmt.Sprintf("Execution Time : %v", elapsed))
			log.LOG_Info(fmt.Sprintf(_Reporter.String()))
		}
	},
}
