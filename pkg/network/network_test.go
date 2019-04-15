package network_test

import (
	"fmt"
	"git.ronaksoftware.com/ronak/riversdk/msg"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/network"
	"testing"
	"time"
)

/*
   Creation Time: 2019 - Apr - 15
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2018
*/

var (
	_Log log.Logger
)
func init() {

}
func dummyMessageHandler(messages []*msg.MessageEnvelope) {
	fmt.Println("Message Handler")
}

func dummyUpdateHandler(updates []*msg.UpdateContainer) {

}

func dummyOnConnectHandler() {

}
func dummyErrorHandler(e *msg.Error) {

}

func dummyNetworkChangeHandler(newStatus domain.NetworkStatus) {

}
func TestNewController(t *testing.T) {
	ctrl := network.NewController(network.Config{
		ServerEndpoint: "ws://new.river.im",
		PingTime:       30 * time.Second,
		PongTimeout:    30 * time.Second,
	})
	ctrl.SetMessageHandler(dummyMessageHandler)
	ctrl.SetErrorHandler(dummyErrorHandler)
	ctrl.SetUpdateHandler(dummyUpdateHandler)
	ctrl.SetNetworkStatusChangedCallback(dummyNetworkChangeHandler)
	ctrl.SetOnConnectCallback(dummyOnConnectHandler)

	err := ctrl.Start()
	if err != nil {
		t.Error(err)
		return
	}

	time.Sleep(20 * time.Second)
	ctrl.Stop()
}
