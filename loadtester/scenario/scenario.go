package scenario

import (
	"fmt"
	"sync"

	"git.ronaksoftware.com/ronak/riversdk/msg"

	"git.ronaksoftware.com/ronak/riversdk/loadtester/actor"
)

type Scenario struct {
	wait sync.WaitGroup
}

func (s *Scenario) Execute(act *actor.Actor) {
	panic("Not implemented")
}

func (s *Scenario) Wait() {
	s.wait.Wait()
}

func (s *Scenario) failed(str string) {
	fmt.Println("Failed : " + str)
	s.wait.Done()
}

func (s *Scenario) completed(str string) {
	fmt.Println("Success : " + str)
	s.wait.Done()
}

func (s *Scenario) isErrorResponse(resp *msg.MessageEnvelope) bool {
	if resp.Constructor == msg.C_Error {
		x := new(msg.Error)
		x.Unmarshal(resp.Message)
		fmt.Println("Response Error : ", x.Code, x.Items)
		s.wait.Done()
		return true
	}
	return false
}
