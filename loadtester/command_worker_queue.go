package main

import (
	"fmt"
	"sync"

	"git.ronaksoftware.com/ronak/riversdk/loadtester/shared"

	"git.ronaksoftware.com/ronak/riversdk/loadtester/actor"
	"git.ronaksoftware.com/ronak/riversdk/loadtester/scenario"
	"git.ronaksoftware.com/ronak/riversdk/log"
	"go.uber.org/zap"
)

type Job struct {
	PhoneNo                    string
	Wait                       *sync.WaitGroup
	Scenario                   shared.Screenwriter
	PhoneListToImportAsContact string
}

var JobQueue chan Job
var StopQueue chan bool

func init() {
	JobQueue = make(chan Job, shared.MaxQueueBuffer)
	StopQueue = make(chan bool)
}

func startDispatcher() {
	for index := 0; index < shared.MaxWorker; index++ {
		go startWorker()
	}
	log.LOG_Info(fmt.Sprintf("Started %d workers successfully", shared.MaxWorker))
}

func stopDispatcher() {
	go func() {
		for index := 0; index < shared.MaxWorker; index++ {
			StopQueue <- true
		}
		log.LOG_Info(fmt.Sprintf("Stopped %d workers successfully", shared.MaxWorker))
	}()
}

func startWorker() {
	for {
		select {
		case <-StopQueue:
			return
		case job := <-JobQueue:
			act, err := actor.NewActor(job.PhoneNo)
			// import phone list if exist
			if job.PhoneListToImportAsContact != "" {
				act.SetPhoneList([]string{job.PhoneListToImportAsContact})
			}

			if err != nil {
				log.LOG_Error(fmt.Sprintf("NewActor(%s)", job.PhoneNo), zap.String("Error", err.Error()))
				continue
			}
			_Reporter.Register(act)
			scenario.Play(act, job.Scenario)
			job.Wait.Done()

		}
	}
}