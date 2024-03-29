package riversdk

import (
    "crypto/rand"
    "crypto/rsa"
    "encoding/binary"
    "math/big"
    "os"
    "runtime"
    "sync"
    "time"

    "github.com/monnand/dhkx"
    "github.com/ronaksoft/river-msg/go/msg"
    "github.com/ronaksoft/river-sdk/internal/domain"
    "github.com/ronaksoft/river-sdk/internal/hole"
    "github.com/ronaksoft/river-sdk/internal/logs"
    mon "github.com/ronaksoft/river-sdk/internal/monitoring"
    "github.com/ronaksoft/river-sdk/internal/repo"
    "github.com/ronaksoft/river-sdk/internal/request"
    "github.com/ronaksoft/river-sdk/internal/salt"
    "github.com/ronaksoft/rony"
    "github.com/ronaksoft/rony/registry"
    "github.com/ronaksoft/rony/tools"
    "go.uber.org/zap"
)

func (r *River) Execute(cb request.Callback) (err error) {
    return r.executeCommand(cb)
}

// ExecuteCommand is a wrapper function to pass the request to the queueController, to be passed to networkController for final
// delivery to the server. SDK uses GetCurrentTeam() to detect the targeted team of the request
func (r *River) ExecuteCommand(constructor int64, commandBytes []byte, delegate RequestDelegate) (requestID int64, err error) {
    requestID = domain.SequentialUniqueID()
    err = r.executeCommand(
        request.DelegateAdapter(
            domain.GetCurrTeamID(), domain.GetCurrTeamAccess(), uint64(requestID), constructor, commandBytes, delegate, nil,
        ),
    )
    return requestID, err
}

// ExecuteCommandWithTeam is similar to ExecuteTeam but explicitly defines the target team
func (r *River) ExecuteCommandWithTeam(
        teamID, accessHash, constructor int64, commandBytes []byte, delegate RequestDelegate,
) (requestID int64, err error) {
    requestID = domain.SequentialUniqueID()
    err = r.executeCommand(
        request.DelegateAdapter(
            teamID, uint64(accessHash), uint64(requestID), constructor, commandBytes, delegate, nil,
        ),
    )
    return requestID, err
}

func (r *River) executeCommand(reqCB request.Callback) (err error) {
    if registry.ConstructorName(reqCB.Constructor()) == "" {
        err = domain.ErrInvalidConstructor
        return
    }

    var (
        waitGroup    = &sync.WaitGroup{}
        blockingMode = reqCB.Flags()&request.Blocking != 0
        serverForce  = reqCB.Flags()&request.ServerForced != 0
    )

    logger.Debug("executes command",
        zap.Uint64("ReqID", reqCB.RequestID()),
        zap.String("C", registry.ConstructorName(reqCB.Constructor())),
        zap.String("Flags", request.DelegateFlagToString(reqCB.Flags())),
    )

    // if function is in blocking mode set the waitGroup to block until the job is done, otherwise
    // save 'delegate' into delegates list to be fetched later.
    if blockingMode {
        waitGroup.Add(1)
        defer waitGroup.Wait()
    }

    // If the constructor is a local command then
    handler, ok := r.localCommands[reqCB.Constructor()]
    if ok && !serverForce {
        go func() {
            r.executeLocalCommand(handler, reqCB)
            if blockingMode {
                waitGroup.Done()
            }
        }()
        return
    }

    go func() {
        r.executeRemoteCommand(reqCB)
        if blockingMode {
            waitGroup.Done()
        }
    }()
    return
}
func (r *River) executeLocalCommand(handler request.LocalHandler, reqCB request.Callback) {
    logger.Info("execute local command",
        zap.Uint64("ReqID", reqCB.RequestID()),
        zap.String("C", registry.ConstructorName(reqCB.Constructor())),
        zap.String("Flags", request.DelegateFlagToString(reqCB.Flags())),
    )

    out := &rony.MessageEnvelope{
        RequestID: reqCB.RequestID(),
    }
    out.Header = append(out.Header, reqCB.Envelope().Header...)

    handler(reqCB)
}
func (r *River) executeRemoteCommand(reqCB request.Callback) {
    logger.Info("execute remote command",
        zap.Uint64("ReqID", reqCB.RequestID()),
        zap.String("C", registry.ConstructorName(reqCB.Constructor())),
        zap.String("Flags", request.DelegateFlagToString(reqCB.Flags())),
    )

    var (
        directToNet    = r.realTimeCommands[reqCB.Constructor()]
        waitForNetwork = true
    )

    if reqCB.Flags()&request.SkipWaitForNetwork != 0 {
        waitForNetwork = false
        directToNet = true

        go func() {
            <-time.After(reqCB.Timeout())
            reqCB := request.GetCallback(reqCB.RequestID())
            if reqCB != nil {
                reqCB.OnTimeout()
                r.CancelRequest(int64(reqCB.RequestID()))
            }
        }()
    }
    if reqCB.Flags()&request.Realtime != 0 {
        directToNet = true
    }

    if waitForNetwork {
        r.networkCtrl.WaitForNetwork(true)
    }

    // If the constructor is a realtime command, then just send it to the server
    if directToNet {
        r.networkCtrl.WebsocketCommand(reqCB)
    } else {
        r.queueCtrl.EnqueueCommand(reqCB)
    }
}

// CreateAuthKey creates an AuthID and AuthKey to be used for transporting messages between client and server
func (r *River) CreateAuthKey() (err error) {
    logger.Info("CreateAuthKey()")

    // Wait for network
    r.networkCtrl.WaitForNetwork(false)

    sk, err := r.getServerKeys()
    if err != nil {
        logger.Warn("got error on SystemGetServers")
        return
    }

    clientNonce, serverNonce, serverPubFP, serverDHFP, serverPQ, err := r.initConnect()
    if err != nil {
        logger.Warn("got error on InitConnect", zap.Error(err))
        return
    }
    logger.Info("passed the 1st step of CreateAuthKey",
        zap.Uint64("ServerNonce", serverNonce),
        zap.Uint64("ServerPubFP", serverPubFP),
        zap.Uint64("ServerPQ", serverPQ),
    )

    err = r.initCompleteAuth(sk, clientNonce, serverNonce, serverPubFP, serverDHFP, serverPQ)
    logger.Info("passed the 2nd step of CreateAuthKey")

    // double set AuthID
    r.networkCtrl.SetAuthorization(r.ConnInfo.AuthID, r.ConnInfo.AuthKey[:])

    r.ConnInfo.Save()

    return
}
func (r *River) getServerKeys() (sk *msg.SystemKeys, err error) {
    logger.Info("GetServerKeys")
    r.networkCtrl.WebsocketCommand(
        request.NewCallback(
            0, 0,
            domain.NextRequestID(),
            msg.C_SystemGetServerKeys,
            &msg.SystemGetServerKeys{},
            func() {
                err = domain.ErrRequestTimeout
            },
            func(res *rony.MessageEnvelope) {
                logger.Debug("GetServerKeys() Success Callback Called")
                switch res.Constructor {
                case msg.C_SystemKeys:
                    sk = &msg.SystemKeys{}
                    err = sk.Unmarshal(res.Message)
                    if err != nil {
                        logger.Error("couldn't unmarshal SystemKeys response", zap.Error(err))
                        return
                    }

                    logger.Debug("received SystemKeys",
                        zap.Int("Keys", len(sk.RSAPublicKeys)),
                        zap.Int("DHGroups", len(sk.DHGroups)),
                    )
                case rony.C_Error:
                    err = domain.ParseServerError(res.Message)
                default:
                    err = domain.ErrInvalidConstructor
                }
            },
            nil,
            false,
            0, domain.WebsocketRequestTimeout,
        ),
    )

    return

}
func (r *River) initConnect() (clientNonce, serverNonce, serverPubFP, serverDHFP, serverPQ uint64, err error) {
    logger.Info("CreateAuthKey() 1st Step Started :: InitConnect")
    r.networkCtrl.WebsocketCommand(
        request.NewCallback(
            0, 0,
            domain.NextRequestID(),
            msg.C_InitConnect,
            &msg.InitConnect{
                ClientNonce: domain.NextRequestID(),
            },
            func() {
                err = domain.ErrRequestTimeout
            },
            func(res *rony.MessageEnvelope) {
                logger.Debug("CreateAuthKey() Success Callback Called")
                switch res.Constructor {
                case msg.C_InitResponse:
                    x := new(msg.InitResponse)
                    err = x.Unmarshal(res.Message)
                    if err != nil {
                        logger.Error("CreateAuthKey() Success Callback", zap.Error(err))
                    }
                    clientNonce = x.ClientNonce
                    serverNonce = x.ServerNonce
                    serverPubFP = x.RSAPubKeyFingerPrint
                    serverDHFP = x.DHGroupFingerPrint
                    serverPQ = x.PQ
                    logger.Debug("CreateAuthKey() InitResponse Received",
                        zap.Uint64("ServerNonce", serverNonce),
                        zap.Uint64("ClientNonce", clientNonce),
                        zap.Uint64("ServerDhFingerPrint", serverDHFP),
                        zap.Uint64("ServerFingerPrint", serverPubFP),
                    )
                case rony.C_Error:
                    err = domain.ParseServerError(res.Message)
                default:
                    err = domain.ErrInvalidConstructor
                }
            },
            nil,
            false, 0, domain.WebsocketRequestTimeout,
        ),
    )
    return
}
func (r *River) initCompleteAuth(sk *msg.SystemKeys, clientNonce, serverNonce, serverPubFP, serverDHFP, serverPQ uint64) (err error) {
    logger.Info("CreateAuthKey() 2nd Step Started :: InitCompleteAuth")

    dhGroup, err := r.getDhGroup(sk, int64(serverDHFP))
    if err != nil {
        return err
    }
    dhPrime := big.NewInt(0)
    dhPrime.SetString(dhGroup.Prime, 16)

    dh := dhkx.CreateGroup(dhPrime, big.NewInt(int64(dhGroup.Gen)))
    clientDhKey, _ := dh.GeneratePrivateKey(rand.Reader)

    req2 := &msg.InitCompleteAuth{
        ServerNonce:    serverNonce,
        ClientNonce:    clientNonce,
        ClientDHPubKey: clientDhKey.Bytes(),
    }

    p, q := domain.SplitPQ(big.NewInt(int64(serverPQ)))
    if p.Cmp(q) < 0 {
        req2.P = p.Uint64()
        req2.Q = q.Uint64()
    } else {
        req2.P = q.Uint64()
        req2.Q = p.Uint64()
    }

    logger.Debug("CreateAuthKey() PQ Split",
        zap.Uint64("P", req2.P),
        zap.Uint64("Q", req2.Q),
    )

    serverPubKey, err := r.getPublicKey(sk, int64(serverPubFP))
    if err != nil {
        return err
    }
    n := big.NewInt(0)
    n.SetString(serverPubKey.N, 10)
    rsaPublicKey := rsa.PublicKey{
        N: n,
        E: int(serverPubKey.E),
    }

    q2Internal := &msg.InitCompleteAuthInternal{
        SecretNonce: []byte(domain.RandomID(16)),
    }
    decrypted, _ := q2Internal.Marshal()
    encrypted, err := rsa.EncryptPKCS1v15(rand.Reader, &rsaPublicKey, decrypted)
    if err != nil {
        logger.Error("CreateAuthKey() got error on EncryptPKCS1v15()", zap.Error(err))
    }
    req2.EncryptedPayload = encrypted
    r.networkCtrl.WebsocketCommand(
        request.NewCallback(
            0, 0,
            domain.NextRequestID(),
            msg.C_InitCompleteAuth,
            req2,
            func() {
                err = domain.ErrRequestTimeout
            },
            func(res *rony.MessageEnvelope) {
                switch res.Constructor {
                case msg.C_InitAuthCompleted:
                    x := new(msg.InitAuthCompleted)
                    _ = x.Unmarshal(res.Message)
                    switch x.Status {
                    case msg.InitAuthCompleted_OK:
                        serverDhKey, err := dh.ComputeKey(dhkx.NewPublicKey(x.ServerDHPubKey), clientDhKey)
                        if err != nil {
                            logger.Error("CreateAuthKey() -> ComputeKey()", zap.Error(err))
                            return
                        }
                        // r.ConnInfo.AuthKey = serverDhKey.Bytes()
                        copy(r.ConnInfo.AuthKey[:], serverDhKey.Bytes())

                        // authKeyHash, _ := domain.Sha256(r.ConnInfo.AuthKey[:])
                        var authKeyHash [32]byte
                        tools.MustSha256(r.ConnInfo.AuthKey[:], authKeyHash[:0])
                        r.ConnInfo.AuthID = int64(binary.LittleEndian.Uint64(authKeyHash[24:32]))

                        var (
                            secret     []byte
                            secretHash [32]byte
                        )
                        secret = append(secret, q2Internal.SecretNonce...)
                        secret = append(secret, byte(msg.InitAuthCompleted_OK))
                        secret = append(secret, authKeyHash[:8]...)
                        tools.MustSha256(secret, secretHash[:0])
                        if x.SecretHash != binary.LittleEndian.Uint64(secretHash[24:32]) {
                            logger.Warn("CreateAuthKey secret hashes do not match",
                                zap.Uint64("Server", x.SecretHash),
                                zap.Uint64("Client", binary.LittleEndian.Uint64(secretHash[24:32])),
                            )
                            return
                        }
                    case msg.InitAuthCompleted_RETRY:
                        // TODO:: Retry with new DHKey
                    case msg.InitAuthCompleted_FAIL:
                        err = domain.ErrAuthFailed
                        return
                    }
                    r.networkCtrl.SetAuthorization(r.ConnInfo.AuthID, r.ConnInfo.AuthKey[:])
                case rony.C_Error:
                    err = domain.ParseServerError(res.Message)
                    return
                default:
                    err = domain.ErrInvalidConstructor
                    return
                }
            },
            nil,
            false, 0, domain.WebsocketRequestTimeout,
        ),
    )
    return
}
func (r *River) getPublicKey(pk *msg.SystemKeys, keyFP int64) (*msg.RSAPublicKey, error) {
    logger.Info("Public Key loaded",
        zap.Int64("keyFP", keyFP),
    )
    for _, pk := range pk.RSAPublicKeys {
        if pk.FingerPrint == keyFP {

            return pk, nil
        }
    }
    return nil, domain.ErrNotFound
}
func (r *River) getDhGroup(pk *msg.SystemKeys, keyFP int64) (*msg.DHGroup, error) {
    logger.Info("DHGroup Key loaded",
        zap.Int64("keyFP", keyFP),
    )
    for _, dh := range pk.DHGroups {
        if dh.FingerPrint == keyFP {
            return dh, nil
        }
    }
    return nil, domain.ErrNotFound
}

// ResetAuthKey reset authorization information, useful in logout
func (r *River) ResetAuthKey() {
    r.networkCtrl.SetAuthorization(0, nil)
    r.ConnInfo.AuthID = 0
    r.ConnInfo.AuthKey = [256]byte{}
    r.ConnInfo.Save()
}

// CancelRequest remove given requestID callbacks&delegates and if its not processed by queue we skip it on queue distributor
func (r *River) CancelRequest(requestID int64) {
    // Cancel Request
    r.queueCtrl.CancelRequest(uint64(requestID))
}

// DeletePendingMessage removes pending message from DB
func (r *River) DeletePendingMessage(id int64) (isSuccess bool) {
    pmsg, _ := repo.PendingMessages.GetByID(id)
    if pmsg == nil {
        return
    }
    if pmsg.FileID != 0 {
        r.fileCtrl.CancelUploadRequest(pmsg.FileID)
    }

    err := repo.PendingMessages.Delete(id)
    isSuccess = err == nil
    return
}

// GetSyncStatus returns SyncController status
func (r *River) GetSyncStatus() int32 {
    return int32(r.syncCtrl.GetSyncStatus())
}

// Logout drop queue & database , etc ...
func (r *River) Logout(notifyServer bool, reason int) error {
    _, err, _ := domain.SingleFlight.Do("Logout", func() (interface{}, error) {
        logger.Info("Logout Called")

        // unregister device if token exist
        if notifyServer {
            // send logout request to server
            waitGroup := &sync.WaitGroup{}
            waitGroup.Add(1)
            r.syncCtrl.Logout(waitGroup, 3)
            waitGroup.Wait()
            logger.Info("sent a AuthLogout request to server, received response")
        }

        if r.mainDelegate != nil {
            r.mainDelegate.OnSessionClosed(reason)
            logger.Info("called SessionClosed delegate")
        }

        // Stop Controllers
        r.syncCtrl.Stop()
        r.queueCtrl.Stop()
        r.fileCtrl.Stop()
        r.networkCtrl.Stop()
        logger.Info("stopped all the controllers")

        repo.DropAll()
        logger.Info("reset our database")

        r.ConnInfo.FirstName = ""
        r.ConnInfo.LastName = ""
        r.ConnInfo.Phone = ""
        r.ConnInfo.UserID = 0
        r.ConnInfo.Username = ""
        r.ConnInfo.Bio = ""
        r.ConnInfo.Save()
        logger.Info("reset our connection info")

        err := r.AppStart()
        if err != nil {
            return nil, err
        }
        logger.Info("started the app again")

        r.networkCtrl.Connect()
        logger.Info("start connecting to server")
        return nil, err
    })
    return err
}

// UpdateContactInfo update contact name
func (r *River) UpdateContactInfo(teamID int64, userID int64, firstName, lastName string) error {
    return repo.Users.UpdateContactInfo(teamID, userID, firstName, lastName)
}

func (r *River) GetScrollStatus(peerID int64, peerType int32) int64 {
    return repo.MessagesExtra.GetScrollID(domain.GetCurrTeamID(), peerID, peerType, 0)
}

func (r *River) SetScrollStatus(peerID, msgID int64, peerType int32) {
    repo.MessagesExtra.SaveScrollID(domain.GetCurrTeamID(), peerID, peerType, 0, msgID)

}

func (r *River) GetServerTimeUnix() int64 {
    return domain.Now().Unix()
}

// AppForeground must be called every time apps come into foreground.
func (r *River) AppForeground(online bool) {
    logger.Info("APP is Foreground", zap.Bool("Online", online))
    statusOnline = online

    // Set the time we come to foreground
    mon.SetForegroundTime()

    if r.networkCtrl.Connected() {
        err := r.networkCtrl.Ping(domain.RandomUint64(), domain.WebsocketPingTimeout)
        if err != nil {
            logger.Info("AppForeground:: Ping failed, we reconnect", zap.Error(err))
            r.networkCtrl.Reconnect()
        } else {
            r.syncCtrl.Sync()
        }
    } else {
        logger.Info("AppForeground:: Network was disconnected we reconnect")
        r.networkCtrl.Reconnect()
    }
    if online {
        r.syncCtrl.UpdateStatus(statusOnline)
    }
}

// AppBackground must be called every time apps goes into background
func (r *River) AppBackground() {
    logger.Info("APP is Background")
    statusOnline = false
    r.syncCtrl.UpdateStatus(false)

    // Compute the time we have been foreground
    mon.IncForegroundTime()

    // Save the usage
    mon.SaveUsage()
}

// AppKill must be called when app is closed
func (r *River) AppKill() {
    r.AppBackground()
}

// AppStart must be called when app is started
func (r *River) AppStart() error {
    statusOnline = false
    runtime.GOMAXPROCS(runtime.NumCPU() * 2)

    logs.SetSentry(r.ConnInfo.AuthID, r.ConnInfo.UserID, r.sentryDSN)
    logger.Info("Starting")

    // Initialize MessageHole
    hole.Init()

    // Initialize DB replaced with ORM
    err := repo.Init(r.dbPath, r.optimizeForLowMemory)
    if err != nil {
        _ = os.RemoveAll(r.dbPath)
        return err
    }

    repo.SetSelfUserID(r.ConnInfo.UserID)

    confBytes, _ := repo.System.LoadBytes("SysConfig")
    if confBytes != nil {
        domain.SysConfig.Reactions = domain.SysConfig.Reactions[:0]
        err := domain.SysConfig.Unmarshal(confBytes)
        if err != nil {
            logger.Warn("could not unmarshal SysConfig", zap.Error(err))
        }
    }

    // Load the usage stats
    mon.LoadUsage()

    // Update Authorizations
    r.networkCtrl.SetAuthorization(r.ConnInfo.AuthID, r.ConnInfo.AuthKey[:])
    r.syncCtrl.SetUserID(r.ConnInfo.UserID)

    // Update the current salt
    salt.UpdateSalt()

    // Start Controllers
    r.networkCtrl.Start()
    r.fileCtrl.Start()
    r.queueCtrl.Start(r.resetQueueOnStartup)
    r.syncCtrl.Start()

    lastReIndexTime, err := repo.System.LoadInt(domain.SkReIndexTime)
    if err != nil || time.Now().Unix()-int64(lastReIndexTime) > domain.Day {
        go func() {
            logger.Info("ReIndexing Users & Groups")
            repo.Users.ReIndex(domain.GetCurrTeamID())
            repo.Groups.ReIndex()
            repo.Messages.ReIndex()
            _ = repo.System.SaveInt(domain.SkReIndexTime, uint64(time.Now().Unix()))
        }()
    }

    domain.StartTime = time.Now()
    domain.WindowLog = func(txt string) {
        r.mainDelegate.AddLog(txt)
    }
    logger.Info("Started")

    // Run update/message processors in background
    go r.messageReceiver()
    go r.updateReceiver()

    // Try to keep the user's status online
    go r.updateStatusJob()

    // Run Garbage Collection In Background
    go func() {
        time.Sleep(10 * time.Second)
        repo.GC()
    }()

    return nil
}

func (r *River) SetTeam(teamID int64, teamAccessHash int64, forceSync bool) {
    domain.SetCurrentTeam(teamID, uint64(teamAccessHash))

    if teamID != 0 {
        r.syncCtrl.TeamSync(teamID, uint64(teamAccessHash), forceSync)
    }
}

func (r *River) Version() string {
    return domain.SDKVersion
}

/*
	Online Status
*/

var statusOnline bool

func (r *River) updateStatusJob() {
    d := time.Duration(domain.SysConfig.OnlineUpdatePeriodInSec-5) * time.Second
    for {
        time.Sleep(d)
        r.syncCtrl.UpdateStatus(statusOnline)
    }
}
