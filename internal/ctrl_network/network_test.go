package networkCtrl_test

import (
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/ctrl_network"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/internal/request"
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
	requestID   uint64
	ctrl        *networkCtrl.Controller
	messageChan = make(chan []*rony.MessageEnvelope, 100)
	updateChan  = make(chan *msg.UpdateContainer, 100)
)

func dummyMessageReceiver() {
	for msgs := range messageChan {
		for _, m := range msgs {
			testenv.Log().Info("Message",
				zap.String("C", registry.ConstructorName(m.Constructor)),
				zap.Uint64("ReqID", m.RequestID),
			)
		}
	}
}

func dummyUpdateReceiver() {
	for updateContainer := range updateChan {
		testenv.Log().Info("Update Handler")
		for _, u := range updateContainer.Updates {
			testenv.Log().Info("Update",
				zap.String("C", registry.ConstructorName(u.Constructor)),
				zap.Int64("GetUpdateID", u.UpdateID),
			)
		}
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

func init() {
	testenv.Log().SetLogLevel(0)
	ctrl = networkCtrl.New(networkCtrl.Config{
		SeedHosts:   []string{"edge.river.im", "edge.rivermsg.com"},
		CountryCode: "IR",
	})
	go dummyMessageReceiver()
	go dummyUpdateReceiver()
	ctrl.OnGeneralError = dummyErrorHandler
	ctrl.OnNetworkStatusChange = dummyNetworkChangeHandler
	ctrl.OnWebsocketConnect = dummyOnConnectHandler
	ctrl.MessageChan = messageChan
	ctrl.UpdateChan = updateChan
}

func TestNewController(t *testing.T) {
	ctrl.Start()
	ctrl.Connect()
	go func() {
		for {
			ctrl.WebsocketCommand(
				request.NewCallback(
					0, 0, atomic.AddUint64(&requestID, 1), msg.C_SystemGetServerTime, &msg.SystemGetServerTime{},
					func() {
						t.Error(domain.ErrRequestTimeout)
					},
					func(res *rony.MessageEnvelope) {

					},
					nil, false, request.SkipFlusher, 0,
				),
			)
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
			ctrl.WebsocketCommand(
				request.NewCallback(
					0, 0, atomic.AddUint64(&requestID, 1), msg.C_SystemGetServerTime, &msg.SystemGetServerTime{},
					func() {
						t.Error(domain.ErrRequestTimeout)
					},
					func(res *rony.MessageEnvelope) {

					},
					nil, false, request.SkipFlusher, 0,
				),
			)
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
				ctrl := networkCtrl.New(networkCtrl.Config{
					SeedHosts:   []string{"edge.river.im"},
					CountryCode: "IR",
				})

				ctrl.OnGeneralError = dummyErrorHandler
				ctrl.OnNetworkStatusChange = dummyNetworkChangeHandler
				ctrl.OnWebsocketConnect = dummyOnConnectHandler
				ctrl.MessageChan = messageChan
				ctrl.UpdateChan = updateChan
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
