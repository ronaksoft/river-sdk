package riversdk

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/binary"
	"fmt"
	"git.ronaksoft.com/river/msg/msg"
	"git.ronaksoft.com/ronak/riversdk/internal/logs"
	mon "git.ronaksoft.com/ronak/riversdk/internal/monitoring"
	fileCtrl "git.ronaksoft.com/ronak/riversdk/pkg/ctrl_file"
	networkCtrl "git.ronaksoft.com/ronak/riversdk/pkg/ctrl_network"
	queueCtrl "git.ronaksoft.com/ronak/riversdk/pkg/ctrl_queue"
	syncCtrl "git.ronaksoft.com/ronak/riversdk/pkg/ctrl_sync"
	"git.ronaksoft.com/ronak/riversdk/pkg/domain"
	messageHole "git.ronaksoft.com/ronak/riversdk/pkg/message_hole"
	"git.ronaksoft.com/ronak/riversdk/pkg/repo"
	"git.ronaksoft.com/ronak/riversdk/pkg/salt"
	"git.ronaksoft.com/ronak/riversdk/pkg/uiexec"
	"github.com/monnand/dhkx"
	"go.uber.org/zap"
	"math/big"
	"runtime"
	"strings"
	"sync"
	"time"
)

// ExecuteCommand is a wrapper function to pass the request to the queueController, to be passed to networkController for final
// delivery to the server. SDK uses GetCurrentTeam() to detect the targeted team of the request
func (r *River) ExecuteCommand(constructor int64, commandBytes []byte, delegate RequestDelegate) (requestID int64, err error) {
	return r.executeCommand(GetCurrTeam(), constructor, commandBytes, delegate)
}

// ExecuteCommandWithTeam is similar to ExecuteTeam but explicitly defines the target team
func (r *River) ExecuteCommandWithTeam(teamID, accessHash, constructor int64, commandBytes []byte, delegate RequestDelegate) (requestID int64, err error) {
	return r.executeCommand(GetTeam(teamID, uint64(accessHash)), constructor, commandBytes, delegate)
}

func (r *River) executeCommand(
	team *msg.InputTeam, constructor int64, commandBytes []byte, delegate RequestDelegate,
) (requestID int64, err error) {
	if _, ok := msg.ConstructorNames[constructor]; !ok {
		return 0, domain.ErrInvalidConstructor
	}

	commandBytesDump := deepCopy(commandBytes)

	waitGroup := new(sync.WaitGroup)
	requestID = domain.SequentialUniqueID()
	logs.Debug("River executes command",
		zap.String("C", msg.ConstructorNames[constructor]),
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
	successCallback := func(envelope *msg.MessageEnvelope) {
		b, _ := envelope.Marshal()
		delegate.OnComplete(b)
		releaseDelegate(r, uint64(requestID))
		if blockingMode {
			waitGroup.Done()
		}
	}

	// If this request must be sent to the server then executeRemoteCommand
	if serverForce {
		executeRemoteCommand(team, r, uint64(requestID), constructor, commandBytesDump, timeoutCallback, successCallback)
		return
	}

	// If the constructor is a local command then
	handler, ok := r.localCommands[constructor]
	if ok {
		if blockingMode {
			executeLocalCommand(team, handler, uint64(requestID), constructor, commandBytesDump, timeoutCallback, successCallback)
		} else {
			go executeLocalCommand(team, handler, uint64(requestID), constructor, commandBytesDump, timeoutCallback, successCallback)
		}
		return
	}

	// If we reached here, then execute the remote commands
	executeRemoteCommand(team, r, uint64(requestID), constructor, commandBytesDump, timeoutCallback, successCallback)

	return
}
func executeLocalCommand(
	team *msg.InputTeam,
	handler domain.LocalMessageHandler,
	requestID uint64, constructor int64, commandBytes []byte,
	timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler,
) {
	logs.Debug("We execute local command",
		zap.String("C", msg.ConstructorNames[constructor]),
	)

	in := &msg.MessageEnvelope{
		Team:        team,
		Constructor: constructor,
		Message:     commandBytes,
		RequestID:   requestID,
	}
	out := &msg.MessageEnvelope{
		Team:      team,
		RequestID: requestID,
	}
	handler(in, out, timeoutCB, successCB)
}
func executeRemoteCommand(
	team *msg.InputTeam,
	r *River,
	requestID uint64, constructor int64, commandBytes []byte,
	timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler,
) {
	logs.Debug("We execute remote command",
		zap.String("C", msg.ConstructorNames[constructor]),
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
			case <-time.After(domain.WebsocketRequestTime):
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
		r.queueCtrl.RealtimeCommand(&msg.MessageEnvelope{
			Team:        team,
			Constructor: constructor,
			RequestID:   requestID,
			Message:     commandBytes,
		}, timeoutCB, successCB, blocking, true)
	} else {
		r.queueCtrl.EnqueueCommand(
			&msg.MessageEnvelope{
				Team:        team,
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

// CreateAuthKey ...
// This function creates an AuthID and AuthKey to be used for transporting messages between client and server
func (r *River) CreateAuthKey() (err error) {
	logs.Info("River::CreateAuthKey()")

	// Wait for network
	r.networkCtrl.WaitForNetwork(false)

	err, clientNonce, serverNonce, serverPubFP, serverDHFP, serverPQ := initConnect(r)
	if err != nil {
		logs.Warn("River got error on InitConnect", zap.Error(err))
		return
	}
	logs.Info("River passed the 1st step of CreateAuthKey")

	err = initCompleteAuth(r, clientNonce, serverNonce, serverPubFP, serverDHFP, serverPQ)
	logs.Info("River passed the 2nd step of CreateAuthKey")

	// double set AuthID
	r.networkCtrl.SetAuthorization(r.ConnInfo.AuthID, r.ConnInfo.AuthKey[:])

	r.ConnInfo.Save()

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
		&msg.InputTeam{},
		r,
		uint64(domain.SequentialUniqueID()),
		msg.C_InitConnect,
		req1Bytes,
		func() {
			defer waitGroup.Done()
			err = domain.ErrRequestTimeout
		},
		func(res *msg.MessageEnvelope) {
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
					zap.Uint64("ClientNounce", clientNonce),
					zap.Uint64("ServerDhFingerPrint", serverDHFP),
					zap.Uint64("ServerFingerPrint", serverPubFP),
				)
			case msg.C_Error:
				err = domain.ParseServerError(res.Message)
			default:
				err = domain.ErrInvalidConstructor
			}
		},
	)
	waitGroup.Wait()
	return
}
func initCompleteAuth(r *River, clientNonce, serverNonce, serverPubFP, serverDHFP, serverPQ uint64) (err error) {
	logs.Info("River::CreateAuthKey() 2nd Step Started :: InitConnect")
	req2 := new(msg.InitCompleteAuth)
	req2.ServerNonce = serverNonce
	req2.ClientNonce = clientNonce
	dhGroup, err := _ServerKeys.GetDhGroup(int64(serverDHFP))
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

	serverPubKey, err := _ServerKeys.GetPublicKey(int64(serverPubFP))
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
		&msg.InputTeam{},
		r,
		uint64(domain.SequentialUniqueID()),
		msg.C_InitCompleteAuth,
		req2Bytes,
		func() {
			defer waitGroup.Done()
			err = domain.ErrRequestTimeout
		},
		func(res *msg.MessageEnvelope) {
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
					authKeyHash, _ := domain.Sha256(r.ConnInfo.AuthKey[:])
					r.ConnInfo.AuthID = int64(binary.LittleEndian.Uint64(authKeyHash[24:32]))

					var secret []byte
					secret = append(secret, q2Internal.SecretNonce...)
					secret = append(secret, byte(msg.InitAuthCompleted_OK))
					secret = append(secret, authKeyHash[:8]...)
					secretHash, _ := domain.Sha256(secret)

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
			case msg.C_Error:
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

// ResetAuthKey
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

// Delete removes pending message from DB
func (r *River) DeletePendingMessage(id int64) (isSuccess bool) {
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
		&msg.MessageEnvelope{
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
			requestID := domain.SequentialUniqueID()
			req := new(msg.AuthLogout)
			buff, _ := req.Marshal()
			r.queueCtrl.RealtimeCommand(
				&msg.MessageEnvelope{
					Constructor: msg.C_AuthLogout,
					RequestID:   uint64(requestID),
					Message:     buff,
				},
				nil, nil, true, false)
			logs.Info("We sent a AuthLogout request to server")
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
func (r *River) UpdateContactInfo(userID int64, firstName, lastName string) error {
	return repo.Users.UpdateContactInfo(userID, firstName, lastName)
}

// GetScrollStatus
func (r *River) GetScrollStatus(peerID int64, peerType int32) int64 {
	return repo.MessagesExtra.GetScrollID(GetCurrTeamID(), peerID, peerType)
}

// SetScrollStatus
func (r *River) SetScrollStatus(peerID, msgID int64, peerType int32) {
	repo.MessagesExtra.SaveScrollID(GetCurrTeamID(), peerID, peerType, msgID)

}

// GetServerTimeUnix
func (r *River) GetServerTimeUnix() int64 {
	return domain.Now().Unix()
}

// AppForeground
func (r *River) AppForeground() {
	// Set the time we come to foreground
	mon.SetForegroundTime()

	if r.networkCtrl.GetQuality() == domain.NetworkConnected {
		err := r.networkCtrl.Ping(domain.RandomUint64(), domain.WebsocketPingTimeout)
		if err != nil {
			logs.Debug("AppForeground:: Ping failed, we reconnect", zap.Error(err))
			r.networkCtrl.Reconnect()
		} else {
			r.syncCtrl.Sync()
		}
	} else {
		logs.Debug("AppForeground:: Network was disconnected we reconnect")
		r.networkCtrl.Reconnect()
	}
}

// AppBackground
func (r *River) AppBackground() {
	// Compute the time we have been foreground
	mon.IncForegroundTime()

	// Save the usage
	mon.SaveUsage()
}

// AppKill
func (r *River) AppKill() {
	r.AppBackground()
}

// AppStart
func (r *River) AppStart() error {
	runtime.GOMAXPROCS(runtime.NumCPU() * 10)

	logs.Info("River Starting")
	logs.SetSentry(r.ConnInfo.AuthID, r.ConnInfo.UserID)

	// Initialize MessageHole
	messageHole.Init()

	// Initialize DB replaced with ORM
	err := repo.InitRepo(r.dbPath, r.optimizeForLowMemory)
	if err != nil {
		return err
	}

	repo.SetSelfUserID(r.ConnInfo.UserID)

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
	r.queueCtrl.Start(r.resetQueueOnStartup)
	r.syncCtrl.Start()
	r.fileCtrl.Start()

	lastReIndexTime, err := repo.System.LoadInt(domain.SkReIndexTime)
	if err != nil || time.Now().Unix()-int64(lastReIndexTime) > domain.Day {
		go func() {
			logs.Info("ReIndexing Users & Groups")
			repo.Users.ReIndex()
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

	// Run Garbage Collection In Background
	go func() {
		time.Sleep(10 * time.Second)
		repo.GC()
	}()

	domain.SysConfig = &msg.SystemConfig{
		TestMode:                false,
		PhoneCallEnabled:        false,
		ExpireOn:                0,
		GroupMaxSize:            250,
		ForwardedMaxCount:       50,
		OnlineUpdatePeriodInSec: 90,
		EditTimeLimitInSec:      86400,
		RevokeTimeLimitInSec:    86400,
		PinnedDialogsMaxCount:   7,
		UrlPrefix:               0,
		MessageMaxLength:        4096,
		CaptionMaxLength:        4096,
		DCs:                     nil,
		MaxLabels:               20,
		TopPeerDecayRate:        3500000,
		TopPeerMaxStep:          365,
		GifBot:                  "",
		WikiBot:                 "",
	}
	confBytes, _ := repo.System.LoadBytes("SysConfig")
	if confBytes != nil {
		err := domain.SysConfig.Unmarshal(confBytes)
		if err != nil {
			logs.Warn("We could not unmarshal SysConfig", zap.Error(err))
		}
	}

	return nil
}

// SetConfig must be called before any other function, otherwise it panics
func (r *River) SetConfig(conf *RiverConfig) {
	domain.ClientPlatform = conf.ClientPlatform
	domain.ClientVersion = conf.ClientVersion
	domain.ClientOS = conf.ClientOs
	domain.ClientVendor = conf.ClientVendor

	r.optimizeForLowMemory = conf.OptimizeForLowMemory
	r.resetQueueOnStartup = conf.ResetQueueOnStartup
	r.ConnInfo = conf.ConnInfo

	if conf.MaxInFlightDownloads <= 0 {
		conf.MaxInFlightDownloads = 10
	}
	if conf.MaxInFlightUploads <= 0 {
		conf.MaxInFlightUploads = 10
	}

	// Initialize DB Path
	if strings.HasPrefix(conf.DbPath, "file://") {
		conf.DbPath = conf.DbPath[7:]
	}
	conf.DbPath = strings.TrimRight(conf.DbPath, "/ ")
	r.dbPath = fmt.Sprintf("%s/%s.db", conf.DbPath, conf.DbID)

	r.registerCommandHandlers()
	r.delegates = make(map[uint64]RequestDelegate)
	r.mainDelegate = conf.MainDelegate
	r.fileDelegate = conf.FileDelegate

	// set loglevel
	logs.SetLogLevel(conf.LogLevel)

	// set log file path
	if conf.LogDirectory != "" {
		_ = logs.SetLogFilePath(conf.LogDirectory)
	}

	// Initialize realtime requests
	r.realTimeCommands = map[int64]bool{
		msg.C_MessagesSetTyping:   true,
		msg.C_InitConnect:         true,
		msg.C_InitConnectTest:     true,
		msg.C_InitAuthCompleted:   true,
		msg.C_SystemGetConfig:     true,
		msg.C_SystemGetSalts:      true,
		msg.C_SystemGetServerTime: true,
		msg.C_SystemGetPublicKeys: true,
	}

	// Initialize Network Controller
	r.networkCtrl = networkCtrl.New(
		networkCtrl.Config{
			WebsocketEndpoint: conf.ServerEndpoint,
			HttpEndpoint:      conf.FileServerEndpoint,
			CountryCode:       conf.CountryCode,
		},
	)
	r.networkCtrl.OnNetworkStatusChange = func(newQuality domain.NetworkStatus) {
		if r.mainDelegate != nil {
			r.mainDelegate.OnNetworkStatusChanged(int(newQuality))
		}
	}
	r.networkCtrl.OnGeneralError = r.onGeneralError
	r.networkCtrl.OnMessage = r.onReceivedMessage
	r.networkCtrl.OnUpdate = r.onReceivedUpdate
	r.networkCtrl.OnWebsocketConnect = r.onNetworkConnect

	// Initialize FileController
	repo.Files.SetRootFolders(
		conf.DocumentAudioDirectory,
		conf.DocumentFileDirectory,
		conf.DocumentPhotoDirectory,
		conf.DocumentVideoDirectory,
		conf.DocumentCacheDirectory,
	)
	r.fileCtrl = fileCtrl.New(fileCtrl.Config{
		Network:              r.networkCtrl,
		MaxInflightDownloads: conf.MaxInFlightDownloads,
		MaxInflightUploads:   conf.MaxInFlightUploads,
		HttpRequestTimeout:   domain.HttpRequestTime,
		CompletedCB:          r.fileDelegate.OnCompleted,
		ProgressChangedCB:    r.fileDelegate.OnProgressChanged,
		CancelCB:             r.fileDelegate.OnCancel,
		PostUploadProcessCB:  r.postUploadProcess,
	})

	// Initialize queueController
	if q, err := queueCtrl.New(r.fileCtrl, r.networkCtrl, r.dbPath); err != nil {
		logs.Fatal("We couldn't initialize MessageQueue",
			zap.String("Error", err.Error()),
		)
	} else {
		r.queueCtrl = q
	}

	// Initialize Sync Controller
	r.syncCtrl = syncCtrl.NewSyncController(
		syncCtrl.Config{
			ConnInfo:    r.ConnInfo,
			NetworkCtrl: r.networkCtrl,
			QueueCtrl:   r.queueCtrl,
			FileCtrl:    r.fileCtrl,
			SyncStatusChangeCB: func(newStatus domain.SyncStatus) {
				if r.mainDelegate != nil {
					r.mainDelegate.OnSyncStatusChanged(int(newStatus))
				}
			},
			UpdateReceivedCB: func(constructorID int64, b []byte) {
				if r.mainDelegate != nil {
					r.mainDelegate.OnUpdates(constructorID, b)
				}
			},
			AppUpdateCB: func(version string, updateAvailable bool, force bool) {
				if r.mainDelegate != nil {
					r.mainDelegate.AppUpdate(version, updateAvailable, force)
				}
			},
		},
	)

	// Initialize Server Keys
	if err := _ServerKeys.UnmarshalJSON([]byte(conf.ServerKeys)); err != nil {
		logs.Error("We couldn't unmarshal ServerKeys",
			zap.String("Error", err.Error()),
		)
	}

	// Initialize River Connection
	logs.Info("River SetConfig done!")
}

func (r *River) SetTeam(teamID int64, teamAccessHash int64, forceSync bool) {
	_CurrTeamID = teamID
	_CurrTeamAccessHash = uint64(teamAccessHash)

	if _CurrTeamID != 0 {
		r.syncCtrl.TeamSync(_CurrTeamID, _CurrTeamAccessHash, forceSync)
	}
}

func (r *River) Version() string {
	return domain.SDKVersion
}

func (r *River) CountAllUnreadMessages(teamID int64, withMutes bool) int32 {
	u, _, err := repo.Dialogs.CountAllUnread(r.ConnInfo.UserID, teamID, withMutes)
	logs.WarnOnErr("We got error on counting all unread messages", err)
	return u
}

func (r *River) CountAllMentioned(teamID int64) int32 {
	_, m, err := repo.Dialogs.CountAllUnread(r.ConnInfo.UserID, teamID, true)
	logs.WarnOnErr("We got error on counting all mentioned", err)
	return m
}
