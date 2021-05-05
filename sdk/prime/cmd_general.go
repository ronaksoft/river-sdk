package riversdk

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/binary"
	"fmt"
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/internal/logs"
	messageHole "git.ronaksoft.com/river/sdk/internal/message_hole"
	mon "git.ronaksoft.com/river/sdk/internal/monitoring"
	"git.ronaksoft.com/river/sdk/internal/repo"
	"git.ronaksoft.com/river/sdk/internal/salt"
	"git.ronaksoft.com/river/sdk/internal/uiexec"
	"github.com/monnand/dhkx"
	"github.com/ronaksoft/rony"
	"github.com/ronaksoft/rony/registry"
	"github.com/ronaksoft/rony/tools"
	"go.uber.org/zap"
	"math/big"
	"runtime"
	"sync"
	"time"
)

// ExecuteCommand is a wrapper function to pass the request to the queueController, to be passed to networkController for final
// delivery to the server. SDK uses GetCurrentTeam() to detect the targeted team of the request
func (r *River) ExecuteCommand(constructor int64, commandBytes []byte, delegate RequestDelegate) (requestID int64, err error) {
	return r.executeCommand(domain.GetCurrTeamID(), domain.GetCurrTeamAccess(), constructor, commandBytes, delegate)
}

// ExecuteCommandWithTeam is similar to ExecuteTeam but explicitly defines the target team
func (r *River) ExecuteCommandWithTeam(teamID, accessHash, constructor int64, commandBytes []byte, delegate RequestDelegate) (requestID int64, err error) {
	return r.executeCommand(teamID, uint64(accessHash), constructor, commandBytes, delegate)
}

func (r *River) executeCommand(
	teamID int64, teamAccess uint64, constructor int64, commandBytes []byte, delegate RequestDelegate,
) (requestID int64, err error) {
	if registry.ConstructorName(constructor) == "" {
		return 0, domain.ErrInvalidConstructor
	}

	commandBytesDump := deepCopy(commandBytes)

	waitGroup := new(sync.WaitGroup)
	requestID = domain.SequentialUniqueID()
	logs.Debug("River executes command",
		zap.String("C", registry.ConstructorName(constructor)),
	)

	blockingMode := delegate.Flags()&RequestBlocking != 0
	serverForce := delegate.Flags()&RequestServerForced != 0

	// if function is in blocking mode set the waitGroup to block until the job is done, otherwise
	// save 'delegate' into delegates list to be fetched later.
	if blockingMode {
		waitGroup.Add(1)
		defer waitGroup.Wait()
	} else {
		r.delegateMutex.Lock()
		r.delegates[uint64(requestID)] = delegate
		r.delegateMutex.Unlock()
	}

	// Timeout Callback
	timeoutCallback := func() {
		err = domain.ErrRequestTimeout
		delegate.OnTimeout(err)
		releaseDelegate(r, uint64(requestID))
		if blockingMode {
			waitGroup.Done()
		}
	}

	// Success Callback
	successCallback := func(envelope *rony.MessageEnvelope) {
		b, _ := envelope.Marshal()
		delegate.OnComplete(b)
		releaseDelegate(r, uint64(requestID))
		if blockingMode {
			waitGroup.Done()
		}
	}

	// If this request must be sent to the server then executeRemoteCommand
	if serverForce {
		executeRemoteCommand(teamID, teamAccess, r, uint64(requestID), constructor, commandBytesDump, timeoutCallback, successCallback)
		return
	}

	// If the constructor is a local command then
	handler, ok := r.localCommands[constructor]
	if ok {
		if blockingMode {
			executeLocalCommand(teamID, teamAccess, handler, uint64(requestID), constructor, commandBytesDump, timeoutCallback, successCallback)
		} else {
			go executeLocalCommand(teamID, teamAccess, handler, uint64(requestID), constructor, commandBytesDump, timeoutCallback, successCallback)
		}
		return
	}

	// If we reached here, then execute the remote commands
	executeRemoteCommand(teamID, teamAccess, r, uint64(requestID), constructor, commandBytesDump, timeoutCallback, successCallback)

	return
}
func executeLocalCommand(
	teamID int64, teamAccess uint64,
	handler domain.LocalMessageHandler,
	requestID uint64, constructor int64, commandBytes []byte,
	timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler,
) {
	logs.Debug("We execute local command",
		zap.String("C", registry.ConstructorName(constructor)),
	)

	in := &rony.MessageEnvelope{
		Header:      domain.TeamHeader(teamID, teamAccess),
		Constructor: constructor,
		Message:     commandBytes,
		RequestID:   requestID,
	}
	out := &rony.MessageEnvelope{
		Header:    domain.TeamHeader(teamID, teamAccess),
		RequestID: requestID,
	}
	handler(in, out, timeoutCB, successCB)
}
func executeRemoteCommand(
	teamID int64, teamAccess uint64,
	r *River,
	requestID uint64, constructor int64, commandBytes []byte,
	timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler,
) {
	logs.Debug("We execute remote command",
		zap.String("C", registry.ConstructorName(constructor)),
	)

	blocking := false
	dontWaitForNetwork := false
	d, ok := getDelegate(r, requestID)
	if ok {
		blocking = d.Flags()&RequestBlocking != 0
		dontWaitForNetwork = d.Flags()&RequestDontWaitForNetwork != 0
	}

	if dontWaitForNetwork {
		go func() {
			select {
			case <-time.After(domain.WebsocketRequestTimeout):
				reqCB := domain.GetRequestCallback(requestID)
				if reqCB == nil {
					break
				}

				if reqCB.TimeoutCallback != nil {
					if reqCB.IsUICallback {
						uiexec.ExecTimeoutCB(reqCB.TimeoutCallback)
					} else {
						reqCB.TimeoutCallback()
					}
				}

				r.CancelRequest(int64(requestID))
			}
		}()
	}

	// If the constructor is a realtime command, then just send it to the server
	if _, ok := r.realTimeCommands[constructor]; ok {
		r.networkCtrl.WebsocketCommand(&rony.MessageEnvelope{
			Header:      domain.TeamHeader(teamID, teamAccess),
			Constructor: constructor,
			RequestID:   requestID,
			Message:     commandBytes,
		}, timeoutCB, successCB, blocking, true)
	} else {
		r.queueCtrl.EnqueueCommand(
			&rony.MessageEnvelope{
				Header:      domain.TeamHeader(teamID, teamAccess),
				Constructor: constructor,
				RequestID:   requestID,
				Message:     commandBytes,
			},
			timeoutCB, successCB, true,
		)
	}
}
func deepCopy(commandBytes []byte) []byte {
	// Takes a copy of commandBytes b4 IOS/Android GC/OS collect/alter them
	length := len(commandBytes)
	buff := make([]byte, length)
	copy(buff, commandBytes)
	return buff
}
func releaseDelegate(r *River, requestID uint64) {
	logs.Debug("River releases delegate",
		zap.Uint64("ReqID", requestID),
	)
	r.delegateMutex.Lock()
	if _, ok := r.delegates[requestID]; ok {
		delete(r.delegates, requestID)
	}
	r.delegateMutex.Unlock()
}
func getDelegate(r *River, requestID uint64) (RequestDelegate, bool) {
	r.delegateMutex.Lock()
	d, ok := r.delegates[requestID]
	r.delegateMutex.Unlock()
	return d, ok
}

// CreateAuthKey creates an AuthID and AuthKey to be used for transporting messages between client and server
func (r *River) CreateAuthKey() (err error) {
	logs.Info("River::CreateAuthKey()")

	// Wait for network
	r.networkCtrl.WaitForNetwork(false)

	sk, err := getServerKeys(r)
	if err != nil {
		logs.Warn("River got error on SystemGetServers")
		return
	}

	err, clientNonce, serverNonce, serverPubFP, serverDHFP, serverPQ := initConnect(r)
	if err != nil {
		logs.Warn("River got error on InitConnect", zap.Error(err))
		return
	}
	logs.Info("River passed the 1st step of CreateAuthKey",
		zap.Uint64("ServerNonce", serverNonce),
		zap.Uint64("ServerPubFP", serverPubFP),
		zap.Uint64("ServerPQ", serverPQ),
	)

	err = initCompleteAuth(r, sk, clientNonce, serverNonce, serverPubFP, serverDHFP, serverPQ)
	logs.Info("River passed the 2nd step of CreateAuthKey")

	// double set AuthID
	r.networkCtrl.SetAuthorization(r.ConnInfo.AuthID, r.ConnInfo.AuthKey[:])

	r.ConnInfo.Save()

	return
}
func getServerKeys(r *River) (sk *msg.SystemKeys, err error) {
	logs.Info("River::GetServerKeys")
	req := &msg.SystemGetServerKeys{}
	reqBytes, _ := req.Marshal()
	waitGroup := new(sync.WaitGroup)
	waitGroup.Add(1)
	executeRemoteCommand(
		0, 0,
		r,
		uint64(domain.SequentialUniqueID()),
		msg.C_SystemGetServerKeys,
		reqBytes,
		func() {
			defer waitGroup.Done()
			err = domain.ErrRequestTimeout
		},
		func(res *rony.MessageEnvelope) {
			defer waitGroup.Done()
			logs.Debug("River::GetServerKeys() Success Callback Called")
			switch res.Constructor {
			case msg.C_SystemKeys:
				sk = &msg.SystemKeys{}
				err = sk.Unmarshal(res.Message)
				if err != nil {
					logs.Error("SyncCtrl couldn't unmarshal SystemKeys response", zap.Error(err))
					return
				}

				logs.Debug("SyncCtrl received SystemKeys",
					zap.Int("Keys", len(sk.RSAPublicKeys)),
					zap.Int("DHGroups", len(sk.DHGroups)),
				)
			case rony.C_Error:
				err = domain.ParseServerError(res.Message)
			default:
				err = domain.ErrInvalidConstructor
			}
		},
	)
	waitGroup.Wait()
	return

}
func initConnect(r *River) (err error, clientNonce, serverNonce, serverPubFP, serverDHFP, serverPQ uint64) {
	logs.Info("River::CreateAuthKey() 1st Step Started :: InitConnect")
	req1 := new(msg.InitConnect)
	req1.ClientNonce = uint64(domain.SequentialUniqueID())
	req1Bytes, _ := req1.Marshal()
	waitGroup := new(sync.WaitGroup)
	waitGroup.Add(1)
	executeRemoteCommand(
		0, 0,
		r,
		uint64(domain.SequentialUniqueID()),
		msg.C_InitConnect,
		req1Bytes,
		func() {
			defer waitGroup.Done()
			err = domain.ErrRequestTimeout
		},
		func(res *rony.MessageEnvelope) {
			defer waitGroup.Done()
			logs.Debug("River::CreateAuthKey() Success Callback Called")
			switch res.Constructor {
			case msg.C_InitResponse:
				x := new(msg.InitResponse)
				err = x.Unmarshal(res.Message)
				if err != nil {
					logs.Error("River::CreateAuthKey() Success Callback", zap.Error(err))
				}
				clientNonce = x.ClientNonce
				serverNonce = x.ServerNonce
				serverPubFP = x.RSAPubKeyFingerPrint
				serverDHFP = x.DHGroupFingerPrint
				serverPQ = x.PQ
				logs.Debug("River::CreateAuthKey() InitResponse Received",
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
	)
	waitGroup.Wait()
	return
}
func initCompleteAuth(r *River, sk *msg.SystemKeys, clientNonce, serverNonce, serverPubFP, serverDHFP, serverPQ uint64) (err error) {
	logs.Info("River::CreateAuthKey() 2nd Step Started :: InitCompleteAuth")
	req2 := new(msg.InitCompleteAuth)
	req2.ServerNonce = serverNonce
	req2.ClientNonce = clientNonce
	dhGroup, err := getDhGroup(sk, int64(serverDHFP))
	if err != nil {
		return err
	}
	dhPrime := big.NewInt(0)
	dhPrime.SetString(dhGroup.Prime, 16)

	dh := dhkx.CreateGroup(dhPrime, big.NewInt(int64(dhGroup.Gen)))
	clientDhKey, _ := dh.GeneratePrivateKey(rand.Reader)
	req2.ClientDHPubKey = clientDhKey.Bytes()

	p, q := domain.SplitPQ(big.NewInt(int64(serverPQ)))
	if p.Cmp(q) < 0 {
		req2.P = p.Uint64()
		req2.Q = q.Uint64()
	} else {
		req2.P = q.Uint64()
		req2.Q = p.Uint64()
	}
	logs.Debug("River::CreateAuthKey() PQ Split",
		zap.Uint64("P", req2.P),
		zap.Uint64("Q", req2.Q),
	)

	q2Internal := new(msg.InitCompleteAuthInternal)
	q2Internal.SecretNonce = []byte(domain.RandomID(16))

	serverPubKey, err := getPublicKey(sk, int64(serverPubFP))
	if err != nil {
		return err
	}
	n := big.NewInt(0)
	n.SetString(serverPubKey.N, 10)
	rsaPublicKey := rsa.PublicKey{
		N: n,
		E: int(serverPubKey.E),
	}
	decrypted, _ := q2Internal.Marshal()
	encrypted, err := rsa.EncryptPKCS1v15(rand.Reader, &rsaPublicKey, decrypted)
	if err != nil {
		logs.Error("River::CreateAuthKey() -> EncryptPKCS1v15()", zap.Error(err))
	}
	req2.EncryptedPayload = encrypted
	req2Bytes, _ := req2.Marshal()

	waitGroup := new(sync.WaitGroup)
	waitGroup.Add(1)
	executeRemoteCommand(
		0, 0,
		r,
		uint64(domain.SequentialUniqueID()),
		msg.C_InitCompleteAuth,
		req2Bytes,
		func() {
			defer waitGroup.Done()
			err = domain.ErrRequestTimeout
		},
		func(res *rony.MessageEnvelope) {
			defer waitGroup.Done()
			switch res.Constructor {
			case msg.C_InitAuthCompleted:
				x := new(msg.InitAuthCompleted)
				_ = x.Unmarshal(res.Message)
				switch x.Status {
				case msg.InitAuthCompleted_OK:
					serverDhKey, err := dh.ComputeKey(dhkx.NewPublicKey(x.ServerDHPubKey), clientDhKey)
					if err != nil {
						logs.Error("River::CreateAuthKey() -> ComputeKey()", zap.Error(err))
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
						fmt.Println(x.SecretHash, binary.LittleEndian.Uint64(secretHash[24:32]))
						err = domain.ErrSecretNonceMismatch
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
	)
	waitGroup.Wait()
	return
}
func getPublicKey(pk *msg.SystemKeys, keyFP int64) (*msg.RSAPublicKey, error) {
	logs.Info("Public Key loaded",
		zap.Int64("keyFP", keyFP),
	)
	for _, pk := range pk.RSAPublicKeys {
		if pk.FingerPrint == keyFP {

			return pk, nil
		}
	}
	return nil, domain.ErrNotFound
}
func getDhGroup(pk *msg.SystemKeys, keyFP int64) (*msg.DHGroup, error) {
	logs.Info("DHGroup Key loaded",
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
	// Remove delegate
	r.delegateMutex.Lock()
	delete(r.delegates, uint64(requestID))
	r.delegateMutex.Unlock()

	// Remove Callback
	domain.RemoveRequestCallback(uint64(requestID))

	// Cancel Request
	r.queueCtrl.CancelRequest(requestID)

}

// DeletePendingMessage removes pending message from DB
func (r *River) DeletePendingMessage(id int64) (isSuccess bool) {
	pmsg := repo.PendingMessages.GetByID(id)
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

// RetryPendingMessage puts pending message again in command queue to re send it
func (r *River) RetryPendingMessage(id int64) bool {
	pmsg := repo.PendingMessages.GetByID(id)
	if pmsg == nil {
		return false
	}
	req := &msg.MessagesSend{
		Body: pmsg.Body,
		Peer: &msg.InputPeer{
			ID:         pmsg.PeerID,
			AccessHash: pmsg.AccessHash,
			Type:       msg.PeerType(pmsg.PeerType),
		},
		RandomID:   pmsg.RequestID,
		ReplyTo:    pmsg.ReplyTo,
		ClearDraft: pmsg.ClearDraft,
		Entities:   pmsg.Entities,
	}
	buff, _ := req.Marshal()
	r.queueCtrl.EnqueueCommand(
		&rony.MessageEnvelope{
			Constructor: msg.C_MessagesSend,
			RequestID:   uint64(req.RandomID),
			Message:     buff,
		},
		nil, nil, true,
	)

	logs.Debug("River::RetryPendingMessage() Request enqueued")
	return true
}

// GetSyncStatus returns SyncController status
func (r *River) GetSyncStatus() int32 {
	return int32(r.syncCtrl.GetSyncStatus())
}

// Logout drop queue & database , etc ...
func (r *River) Logout(notifyServer bool, reason int) error {
	_, err, _ := domain.SingleFlight.Do("Logout", func() (interface{}, error) {
		logs.Info("Logout Called")

		// unregister device if token exist
		if notifyServer {
			// send logout request to server
			waitGroup := &sync.WaitGroup{}
			waitGroup.Add(1)
			r.syncCtrl.Logout(waitGroup, 3)
			waitGroup.Wait()
			logs.Info("We sent a AuthLogout request to server, received response")
		}

		if r.mainDelegate != nil {
			r.mainDelegate.OnSessionClosed(reason)
			logs.Info("We called SessionClosed delegate")
		}

		// Stop Controllers
		r.syncCtrl.Stop()
		r.queueCtrl.Stop()
		r.fileCtrl.Stop()
		r.networkCtrl.Stop()
		logs.Info("We stopped all the controllers")

		repo.DropAll()
		logs.Info("We reset our database")

		r.ConnInfo.FirstName = ""
		r.ConnInfo.LastName = ""
		r.ConnInfo.Phone = ""
		r.ConnInfo.UserID = 0
		r.ConnInfo.Username = ""
		r.ConnInfo.Bio = ""
		r.ConnInfo.Save()
		logs.Info("We reset our connection info")

		err := r.AppStart()
		if err != nil {
			return nil, err
		}
		logs.Info("We started the app again")

		r.networkCtrl.Connect()
		logs.Info("We start connecting to server")
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
	statusOnline = online

	// Set the time we come to foreground
	mon.SetForegroundTime()

	if r.networkCtrl.GetQuality() == domain.NetworkConnected {
		err := r.networkCtrl.Ping(domain.RandomUint64(), domain.WebsocketPingTimeout)
		if err != nil {
			logs.Info("AppForeground:: Ping failed, we reconnect", zap.Error(err))
			r.networkCtrl.Reconnect()
		} else {
			r.syncCtrl.Sync()
		}
	} else {
		logs.Info("AppForeground:: Network was disconnected we reconnect")
		r.networkCtrl.Reconnect()
	}
	if online {
		r.syncCtrl.UpdateStatus(statusOnline)
	}
}

// AppBackground must be called every time apps goes into background
func (r *River) AppBackground() {
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
	statusOnline = true
	runtime.GOMAXPROCS(runtime.NumCPU() * 2)

	logs.Info("River Starting")
	logs.SetSentry(r.ConnInfo.AuthID, r.ConnInfo.UserID, r.sentryDSN)

	// Initialize MessageHole
	messageHole.Init()

	// Initialize DB replaced with ORM
	err := repo.InitRepo(r.dbPath, r.optimizeForLowMemory)
	if err != nil {
		return err
	}

	repo.SetSelfUserID(r.ConnInfo.UserID)

	confBytes, _ := repo.System.LoadBytes("SysConfig")
	if confBytes != nil {
		domain.SysConfig.Reactions = domain.SysConfig.Reactions[:0]
		err := domain.SysConfig.Unmarshal(confBytes)
		if err != nil {
			logs.Warn("We could not unmarshal SysConfig", zap.Error(err))
		}
	}

	// Load the usage stats
	mon.LoadUsage()

	// Update Authorizations
	r.networkCtrl.SetAuthorization(r.ConnInfo.AuthID, r.ConnInfo.AuthKey[:])
	r.syncCtrl.SetUserID(r.ConnInfo.UserID)

	// Load device token
	r.loadDeviceToken()

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
			logs.Info("ReIndexing Users & Groups")
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
	logs.Info("River Started")

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
	// We wait about 5 seconds to make sure user is actually in app
	for {
		time.Sleep(d)
		r.syncCtrl.UpdateStatus(statusOnline)
	}
}
