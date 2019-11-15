package networkCtrl_test

import (
	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
	"git.ronaksoftware.com/ronak/riversdk/pkg/ctrl_network"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	ronak "git.ronaksoftware.com/ronak/toolbox"
	"go.uber.org/zap"
	"sync"
	"sync/atomic"
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
	requestID uint64
	ctrl      *networkCtrl.Controller
)

func dummyMessageHandler(messages []*msg.MessageEnvelope) {
	for _, m := range messages {
		logs.Info("Message",
			zap.String("Constructor", msg.ConstructorNames[m.Constructor]),
			zap.Uint64("RequestID", m.RequestID),
		)
	}
}

func dummyUpdateHandler(updateContainer *msg.UpdateContainer) {
	logs.Info("Update Handler")
	for _, u := range updateContainer.Updates {
		logs.Info("Update",
			zap.String("Constructor", msg.ConstructorNames[u.Constructor]),
			zap.Int64("UpdateID", u.UpdateID),
		)
	}

}

func dummyOnConnectHandler() {
	logs.Info("Connected")
}

func dummyErrorHandler(requestID uint64, e *msg.Error) {
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

func getServerTime() *msg.MessageEnvelope {
	m := new(msg.SystemGetServerTime)
	b, _ := m.Marshal()
	return &msg.MessageEnvelope{
		Constructor: msg.C_SystemGetServerTime,
		RequestID:   atomic.AddUint64(&requestID, 1),
		Message:     b,
	}
}

func init() {
	logs.SetLogLevel(-1)
	ctrl = networkCtrl.New(networkCtrl.Config{
		WebsocketEndpoint: "ws://river.ronaksoftware.com",
		HttpEndpoint:      "http://river.ronaksoftware.com",
	})
	ctrl.OnMessage = dummyMessageHandler
	ctrl.OnGeneralError = dummyErrorHandler
	ctrl.OnUpdate = dummyUpdateHandler
	ctrl.OnNetworkStatusChange = dummyNetworkChangeHandler
	ctrl.OnWebsocketConnect = dummyOnConnectHandler
}

func TestNewController(t *testing.T) {
	ctrl.Start()
	ctrl.Connect()
	go func() {
		for {
			err := ctrl.SendWebsocket(getServerTime(), true)
			if err != nil {
				t.Error(err)
			}
			time.Sleep(time.Second)
		}
	}()
	time.Sleep(5000 * time.Second)
	ctrl.Disconnect()
	time.Sleep(5 * time.Second)

	ctrl.Stop()
}

func TestConnect(t *testing.T) {
	ctrl.Start()
	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			for i := 0; i < 10; i++ {
				switch ronak.RandomInt(100) % 3 {
				case 0:
					ctrl.Disconnect()
				case 1, 2:
					ctrl.Reconnect()
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()

	ctrl.Stop()
}

func TestReconnect(t *testing.T) {
	ctrl.Start()
	for i := 0 ; i < 10 ; i++ {
		ctrl.Reconnect()
		time.Sleep(time.Second * 5)
	}
}
