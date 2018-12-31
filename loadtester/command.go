package main

import (
	"fmt"

	"git.ronaksoftware.com/ronak/riversdk/loadtester/actor"
	"git.ronaksoftware.com/ronak/riversdk/loadtester/scenario"
	"git.ronaksoftware.com/ronak/riversdk/loadtester/shared"
	"git.ronaksoftware.com/ronak/riversdk/log"
	"go.uber.org/zap"
	ishell "gopkg.in/abiosoft/ishell.v2"
)

var (
	_Actors []shared.Acter
)

var cmdRegister = &ishell.Cmd{
	Name: "Register",
	Func: func(c *ishell.Context) {
		// get suffix start phoneNo
		// get suffix end phoneNo
		// start registering

		startNo := fnStartPhone(c)
		endNo := fnEndPhone(c)
		s := scenario.NewRegister()
		phoneNo := ""
		for startNo <= endNo {
			phoneNo = fmt.Sprintf("237400%07d", startNo)
			startNo++
			act, err := actor.NewActor(phoneNo)
			if err != nil {
				log.LOG_Error(fmt.Sprintf("NewActor(%s)", phoneNo), zap.String("Error", err.Error()))
				continue
			}
			// run sync
			// scenario.Play(act, s)

			// run async
			s.Play(act)
		}
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
		s := scenario.NewLogin()
		phoneNo := ""
		for startNo <= endNo {
			phoneNo = fmt.Sprintf("237400%07d", startNo)
			startNo++
			act, err := actor.NewActor(phoneNo)
			if err != nil {
				log.LOG_Error(fmt.Sprintf("NewActor(%s)", phoneNo), zap.String("Error", err.Error()))
				continue
			}
			// run sync
			// scenario.Play(act, s)

			// run async
			s.Play(act)
		}

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
		s := scenario.NewImportContact()
		phoneNo := ""
		for startNo <= endNo {
			phoneNo = fmt.Sprintf("237400%07d", startNo)
			startNo++
			act, err := actor.NewActor(phoneNo)
			act.SetPhoneList([]string{phoneNoToImportAsContact})
			if err != nil {
				log.LOG_Error(fmt.Sprintf("NewActor(%s)", phoneNo), zap.String("Error", err.Error()))
				continue
			}
			// run sync
			// scenario.Play(act, s)

			// run async
			s.Play(act)
		}
	},
}

var cmdSendMessage = &ishell.Cmd{
	Name: "SendMessage",
	Func: func(c *ishell.Context) {
		// get suffix start phoneNo
		// get suffix end phoneNo
		// start sending to actors peers

		startNo := fnStartPhone(c)
		endNo := fnEndPhone(c)
		s := scenario.NewSendMessage()
		phoneNo := ""
		for startNo <= endNo {
			phoneNo = fmt.Sprintf("237400%07d", startNo)
			startNo++
			act, err := actor.NewActor(phoneNo)
			if err != nil {
				log.LOG_Error(fmt.Sprintf("NewActor(%s)", phoneNo), zap.String("Error", err.Error()))
				continue
			}
			// run sync
			// scenario.Play(act, s)

			// run async
			s.Play(act)
		}
	},
}
