package scenario

import (
	"fmt"
	"sync"

	"git.ronaksoftware.com/ronak/riversdk/msg"

	"git.ronaksoftware.com/ronak/riversdk/loadtester/shared"
)

// Scenario is something that screenwriter create
type Scenario struct {
	wait sync.WaitGroup
}

// Play execute scenario
func (s *Scenario) Play(act shared.Acter) {
	panic("Not implemented")
}

// Wait until play overs
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

// Play puts actor in scenario to play its role and wait to show overs
func Play(act shared.Acter, scenario shared.Screenwriter) {
	scenario.Play(act)
	scenario.Wait()
}
