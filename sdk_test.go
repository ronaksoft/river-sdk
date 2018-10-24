package riversdk

import (
    "bufio"
    "fmt"
    "os"
    "testing"
    "time"

    "git.ronaksoftware.com/customers/river/messages"
    "github.com/dustin/go-humanize"
    "github.com/jmoiron/sqlx"
    "go.uber.org/zap"
)

func dummyMessageHandler(m *msg.MessageEnvelope) {
    log.LOG.Info("MessageEnvelope Handler",
        zap.Uint64("REQ_ID", m.RequestID),
        zap.String("CONSTRUCTOR", msg.ConstructorNames[m.Constructor]),
        zap.String("LENGTH", humanize.Bytes(uint64(len(m.Message)))),
    )
}

func dummyUpdateHandler(u *msg.UpdateContainer) {
    log.LOG.Info("UpdateContainer Handler",
        zap.Int64("MIN_UPDATE_ID", u.MinUpdateID),
        zap.Int64("MAX_UPDATE_ID", u.MaxUpdateID),
        zap.Int32("LENGTH", u.Length),
    )
    for _, update := range u.Updates {
        log.LOG.Info("UpdateEnvelope",
            zap.String("CONSTRUCTOR", msg.ConstructorNames[update.Constructor]),
        )
    }
}

func dummyQualityUpdateHandler(q NetworkStatus) {
    log.LOG.Info("Network Quality Updated",
        zap.Int("New Quality", int(q)),
    )
}

func newTestNetwork(pintTime time.Duration) *networkController {
    ctlConfig := networkConfig{
        ServerEndpoint: "ws://river.im",
        PingTime:       pintTime,
    }

    ctl := newNetworkController(ctlConfig)
    return ctl
}

func newTestQueue(network *networkController) *queueController {
    q, err := newQueueController(network, "./_data/queue", nil)
    if err != nil {
        log.LOG.Fatal(err.Error())
    }
    return q
}

func newTestDB() *sqlx.DB {
    // Initialize Database directory (LevelDB)
    dbPath := "./_data/"
    dbID := "test"
    os.MkdirAll(dbPath, os.ModePerm)
    if db, err := sqlx.Open("sqlite3", fmt.Sprintf("%s/%s.db", dbPath, dbID)); err != nil {
        log.LOG.Fatal(err.Error())
    } else {
        setDB(db)
        return db
    }
    return nil
}

func newTestSync(queue *queueController) *syncController {
    s := newSyncController(
        syncConfig{
            QueueCtrl: queue,
        },
    )
    return s
}

func TestNewSyncController(t *testing.T) {
    setDB(newTestDB())
    networkCtrl := newTestNetwork(20 * time.Second)
    queueCtrl := newTestQueue(networkCtrl)
    syncCtrl := newTestSync(queueCtrl)

    networkCtrl.SetMessageHandler(queueCtrl.messageHandler)
    networkCtrl.SetUpdateHandler(syncCtrl.updateHandler)
    networkCtrl.SetNetworkStatusChangedCallback(dummyQualityUpdateHandler)

    networkCtrl.Start()
    queueCtrl.Start()
    syncCtrl.Start()
    networkCtrl.Connect()

    for i := 0; i < 100; i++ {
        requestID := _RandomUint64()
        initConn := new(msg.InitConnect)
        initConn.ClientNonce = _RandomUint64()
        messageEnvelope := new(msg.MessageEnvelope)
        messageEnvelope.Constructor = msg.C_InitConnect
        messageEnvelope.RequestID = requestID
        messageEnvelope.Message, _ = initConn.Marshal()
        req := request{
            Timeout:         2 * time.Second,
            MessageEnvelope: messageEnvelope,
        }
        log.LOG.Info("prepare request",
            zap.Uint64("ReqID", messageEnvelope.RequestID),
        )
        queueCtrl.addToWaitingList(&req)
    }

    time.Sleep(30 * time.Second)

}

func TestNewNetworkController(t *testing.T) {
    ctl := newTestNetwork(time.Second)
    ctl.SetMessageHandler(dummyMessageHandler)
    ctl.SetUpdateHandler(dummyUpdateHandler)
    ctl.SetNetworkStatusChangedCallback(dummyQualityUpdateHandler)
    if err := ctl.Start(); err != nil {
        log.LOG.Info(err.Error())
        return
    }
    ctl.Connect()

    bufio.NewReader(os.Stdin).ReadBytes('\n')
    ctl.Disconnect()
    ctl.Stop()
}

func TestNewRiver(t *testing.T) {
    log.LOG.Info("Creating New River SDK Instance")
    r := new(River)
    r.SetConfig(&RiverConfig{
        DbPath:             "./_data/",
        DbID:               "test",
        ServerKeysFilePath: "./keys.json",
        ServerEndpoint:     "ws://river.im",
    })

    r.Start()
    if r.ConnInfo.AuthID == 0 {
        log.LOG.Info("AuthKey has not been created yet.")
        if err := r.CreateAuthKey(); err != nil {
            t.Error(err.Error())
            return
        }
        log.LOG.Info("AuthKey Created.")
    }
}
