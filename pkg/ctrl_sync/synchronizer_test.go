package synchronizer

import (
	"fmt"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo"
	"testing"
	"time"

	"git.ronaksoftware.com/ronak/riversdk/pkg/ctrl_network"
	"git.ronaksoftware.com/ronak/riversdk/pkg/ctrl_queue"
)

func TestController_CheckSalt(t *testing.T) {
	type fields struct {
		//connInfo               domain.RiverConfigurator
		networkCtrl            *network.Controller
		queueCtrl              *queue.Controller
		//onSyncStatusChange     domain.SyncStatusUpdateCallback
		//onUpdateMainDelegate   domain.OnUpdateMainDelegateHandler
		//syncStatus             domain.SyncStatus
		//lastUpdateReceived     time.Time
		//updateID               int64
		//updateAppliers         map[int64]domain.UpdateApplier
		//messageAppliers        map[int64]domain.MessageApplier
		//stopChannel            chan bool
		//userID                 int64
		//deliveredMessagesMutex sync.Mutex
		//deliveredMessages      map[int64]bool
		//updateDifferenceLock   int32
		//syncLock               int32
	}
	nctrl := network.NewController(network.Config{
		ServerEndpoint: "ws://test.river.im",
		PingTime:       30 * time.Second,
		PongTimeout:    30 * time.Second,
	})
	dataDir := fmt.Sprintf("%s/%s", "./_queue", "./_db")
	quctrl, _ := queue.NewController(nctrl, dataDir)
	_ = repo.InitRepo("sqlite3", fmt.Sprintf("%s/%s.db", "./_db", ))
	tests := []struct {
		name   string
		fields fields
	}{
		{"test", fields{networkCtrl: nctrl, queueCtrl: quctrl}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := &Controller{
				networkCtrl:            tt.fields.networkCtrl,
				queueCtrl:              tt.fields.queueCtrl,
			}
			ctrl.CheckSalt()
		})
	}
}
