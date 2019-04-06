package main

import (
	"sync"
	"time"

	"git.ronaksoftware.com/ronak/riversdk/loadtester/shared"

	"git.ronaksoftware.com/ronak/riversdk/loadtester/scenario"
	ishell "gopkg.in/abiosoft/ishell.v2"
)

var cmdCreateAuthKey = &ishell.Cmd{
	Name: "CreateAuthKey",
	Func: func(c *ishell.Context) {

		startNo := fnStartPhone(c)
		endNo := fnEndPhone(c)

		// clear
		fnClearScreeen()
		fnClearReports()

		// start workers
		startDispatcher()
		defer stopDispatcher()

		phoneNo := ""
		wg := &sync.WaitGroup{}
		sw := time.Now()
		for startNo <= endNo {
			phoneNo = shared.GetPhone(startNo)
			startNo++

			// Add To Queue
			wg.Add(1)
			JobQueue <- Job{PhoneNo: phoneNo, Wait: wg, Scenario: scenario.NewCreateAuthKey(true)}

		}
		wg.Wait()
		fnPrintReports(time.Since(sw))

	},
}

var cmdRegister = &ishell.Cmd{
	Name: "Register",
	Func: func(c *ishell.Context) {

		startNo := fnStartPhone(c)
		endNo := fnEndPhone(c)

		// clear
		fnClearScreeen()
		fnClearReports()

		// start workers
		startDispatcher()
		defer stopDispatcher()

		phoneNo := ""
		wg := &sync.WaitGroup{}
		sw := time.Now()
		for startNo <= endNo {
			phoneNo = shared.GetPhone(startNo)
			startNo++

			// Add To Queue
			wg.Add(1)
			JobQueue <- Job{PhoneNo: phoneNo, Wait: wg, Scenario: scenario.NewRegister(true)}

		}
		wg.Wait()
		fnPrintReports(time.Since(sw))

	},
}

var cmdLogin = &ishell.Cmd{
	Name: "Login",
	Func: func(c *ishell.Context) {

		startNo := fnStartPhone(c)
		endNo := fnEndPhone(c)

		// clear
		fnClearScreeen()
		fnClearReports()

		// start workers
		startDispatcher()
		defer stopDispatcher()

		phoneNo := ""
		wg := &sync.WaitGroup{}
		sw := time.Now()
		for startNo <= endNo {
			phoneNo = shared.GetPhone(startNo)
			startNo++

			// Add To Queue
			wg.Add(1)
			JobQueue <- Job{PhoneNo: phoneNo, Wait: wg, Scenario: scenario.NewLogin(true)}
		}
		wg.Wait()
		fnPrintReports(time.Since(sw))

	},
}

var cmdImportContact = &ishell.Cmd{
	Name: "ImportContact",
	Func: func(c *ishell.Context) {

		phoneNoToImportAsContact := fnGetPhone(c)
		startNo := fnStartPhone(c)
		endNo := fnEndPhone(c)

		// clear
		fnClearScreeen()
		fnClearReports()

		// start workers
		startDispatcher()
		defer stopDispatcher()

		phoneNo := ""
		wg := &sync.WaitGroup{}
		sw := time.Now()
		for startNo <= endNo {
			phoneNo = shared.GetPhone(startNo)
			startNo++

			// Add To Queue
			wg.Add(1)
			JobQueue <- Job{PhoneNo: phoneNo, Wait: wg, Scenario: scenario.NewImportContact(true), PhoneListToImportAsContact: phoneNoToImportAsContact}

		}
		wg.Wait()
		fnPrintReports(time.Since(sw))

	},
}

var cmdSendMessage = &ishell.Cmd{
	Name: "SendMessage",
	Func: func(c *ishell.Context) {

		startNo := fnStartPhone(c)
		endNo := fnEndPhone(c)

		// clear
		fnClearScreeen()
		fnClearReports()

		// start workers
		startDispatcher()
		defer stopDispatcher()

		phoneNo := ""
		wg := &sync.WaitGroup{}
		sw := time.Now()
		for startNo <= endNo {
			phoneNo = shared.GetPhone(startNo)
			startNo++

			// Add To Queue
			wg.Add(1)
			JobQueue <- Job{PhoneNo: phoneNo, Wait: wg, Scenario: scenario.NewSendMessage(true)}

		}
		wg.Wait()
		fnPrintReports(time.Since(sw))
	},
}

var cmdSendFile = &ishell.Cmd{
	Name: "SendFile",
	Func: func(c *ishell.Context) {

		startNo := fnStartPhone(c)
		endNo := fnEndPhone(c)

		// clear
		fnClearScreeen()
		fnClearReports()

		// start workers
		startDispatcher()
		defer stopDispatcher()

		phoneNo := ""
		wg := &sync.WaitGroup{}
		sw := time.Now()
		for startNo <= endNo {
			phoneNo = shared.GetPhone(startNo)
			startNo++

			// Add To Queue
			wg.Add(1)
			JobQueue <- Job{PhoneNo: phoneNo, Wait: wg, Scenario: scenario.NewSendFile(true)}

		}
		wg.Wait()
		fnPrintReports(time.Since(sw))
	},
}
