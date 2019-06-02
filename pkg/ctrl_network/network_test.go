package network_test

import (
	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
	"git.ronaksoftware.com/ronak/riversdk/pkg/ctrl_network"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	ronak "git.ronaksoftware.com/ronak/toolbox"
	"go.uber.org/zap"
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

func dummyMessageHandler(messages []*msg.MessageEnvelope) {
	logs.Info("Message Handler")
	for _, m := range messages {
		logs.Info("Message",
			zap.String("Constructor", msg.ConstructorNames[m.Constructor]),
			zap.Uint64("RequestID", m.RequestID),
		)
	}
}

func dummyUpdateHandler(updateContainers []*msg.UpdateContainer) {
	logs.Info("Update Handler")
	for _, uc := range updateContainers {
		for _, u := range uc.Updates {
			logs.Info("Update",
				zap.String("Constructor", msg.ConstructorNames[u.Constructor]),
				zap.Int64("UpdateID", u.UpdateID),
			)
		}
	}
}

func dummyOnConnectHandler() {
	logs.Info("Connected")
}

func dummyErrorHandler(e *msg.Error) {
	logs.Info("Error Handler",
		zap.String("Code", e.Code),
		zap.String("Items", e.Items),
	)
}

func dummyNetworkChangeHandler(newStatus domain.NetworkStatus) {
	logs.Info("Network Status Changed",
		zap.String("New Status", newStatus.ToString()),
	)
}

func authRecall() *msg.MessageEnvelope {
	m := new(msg.AuthRecall)
	m.ClientID = 2374
	b, _ := m.Marshal()
	return &msg.MessageEnvelope{
		Constructor: msg.C_AuthRecall,
		RequestID:   ronak.RandomUint64(),
		Message:     b,
	}
}
func TestNewController(t *testing.T) {
	logs.SetLogLevel(0)
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
	ctrl.Connect(true)
	for i := 0; i < 10; i++ {
		err = ctrl.Send(authRecall(), false)
		if err != nil {
			t.Error(err)
		}
	}
	time.Sleep(20 * time.Second)
	ctrl.Disconnect()
	time.Sleep(3 * time.Second)
	ctrl.Stop()
}
