package networkCtrl_test

import (
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/ctrl_network"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/internal/testenv"
	"github.com/ronaksoft/rony"
	"github.com/ronaksoft/rony/registry"
	. "github.com/smartystreets/goconvey/convey"
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

func dummyMessageHandler(messages []*rony.MessageEnvelope) {
	for _, m := range messages {
		testenv.Log().Info("Message",
			zap.String("C", registry.ConstructorName(m.Constructor)),
			zap.Uint64("ReqID", m.RequestID),
		)
	}
}

func dummyUpdateHandler(updateContainer *msg.UpdateContainer) {
	testenv.Log().Info("Update Handler")
	for _, u := range updateContainer.Updates {
		testenv.Log().Info("Update",
			zap.String("C", registry.ConstructorName(u.Constructor)),
			zap.Int64("GetUpdateID", u.UpdateID),
		)
	}

}

func dummyOnConnectHandler() error {
	testenv.Log().Info("Connected")
	return nil
}

func dummyErrorHandler(requestID uint64, e *rony.Error) {
	testenv.Log().Info("Error Handler",
		zap.String("Code", e.Code),
		zap.String("Items", e.Items),
	)
}

func dummyNetworkChangeHandler(newStatus domain.NetworkStatus) {
	testenv.Log().Info("Network Status Changed",
		zap.String("New Status", newStatus.ToString()),
	)
}

func authRecall() *rony.MessageEnvelope {
	m := new(msg.AuthRecall)
	m.ClientID = 2374
	b, _ := m.Marshal()
	return &rony.MessageEnvelope{
		Constructor: msg.C_AuthRecall,
		RequestID:   domain.RandomUint64(),
		Message:     b,
	}
}

func getServerTime() *rony.MessageEnvelope {
	m := new(msg.SystemGetServerTime)
	b, _ := m.Marshal()
	return &rony.MessageEnvelope{
		Constructor: msg.C_SystemGetServerTime,
		RequestID:   atomic.AddUint64(&requestID, 1),
		Message:     b,
	}
}

func init() {
	testenv.Log().SetLogLevel(0)
	ctrl = networkCtrl.New(networkCtrl.Config{
		SeedHosts:   []string{"edge.river.im", "edge.rivermsg.com"},
		CountryCode: "IR",
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
			err := ctrl.WebsocketSend(getServerTime(), domain.RequestSkipFlusher)
			if err != nil {
				t.Error(err)
			}
			time.Sleep(time.Second)
		}
	}()
	time.Sleep(time.Second * 5)
	err := ctrl.Ping(domain.RandomUint64(), domain.WebsocketWriteTime)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(5 * time.Second)

	ctrl.Stop()
}

func TestStartStop(t *testing.T) {
	ctrl.Start()
	go func() {
		for {
			err := ctrl.WebsocketSend(getServerTime(), domain.RequestSkipFlusher)
			if err != nil {
				t.Error(err)
			}
			time.Sleep(time.Second)
		}
	}()
	for j := 0; j < 20; j++ {
		time.Sleep(5 * time.Second)
		ctrl.Stop()
		ctrl.Start()
		ctrl.Connect()
	}

	ctrl.Stop()
}

func TestReconnect(t *testing.T) {
	Convey("Reconnect Test", t, func(c C) {
		wg := sync.WaitGroup{}
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				var ctrl *networkCtrl.Controller
				ctrl = networkCtrl.New(networkCtrl.Config{
					SeedHosts:   []string{"edge.river.im"},
					CountryCode: "IR",
				})
				ctrl.OnMessage = dummyMessageHandler
				ctrl.OnGeneralError = dummyErrorHandler
				ctrl.OnUpdate = dummyUpdateHandler
				ctrl.OnNetworkStatusChange = dummyNetworkChangeHandler
				ctrl.OnWebsocketConnect = dummyOnConnectHandler
				ctrl.Start()
				for i := 0; i < 100; i++ {
					ctrl.Reconnect()
					time.Sleep(time.Millisecond * 250)
				}
				wg.Done()
			}()
		}
		wg.Wait()
	})
}

func TestPing(t *testing.T) {
	ctrl.Start()
	ctrl.Connect()
	for i := 0; i < 10; i++ {
		startTime := time.Now()
		err := ctrl.Ping(domain.RandomUint64(), domain.WebsocketWriteTime)
		if err != nil {
			t.Fatal(err)
		}
		testenv.Log().Info("Pinged", zap.Duration("D", time.Since(startTime)))
	}
	time.Sleep(5 * time.Second)
	ctrl.Stop()
}
