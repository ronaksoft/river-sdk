package riversdk

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"git.ronaksoftware.com/ronak/riversdk/cmd"
	"git.ronaksoftware.com/ronak/riversdk/filemanager"

	"git.ronaksoftware.com/ronak/riversdk/domain"
	"git.ronaksoftware.com/ronak/riversdk/network"
	"git.ronaksoftware.com/ronak/riversdk/queue"
	"git.ronaksoftware.com/ronak/riversdk/synchronizer"

	"git.ronaksoftware.com/ronak/riversdk/repo"

	"git.ronaksoftware.com/ronak/riversdk/log"
	"git.ronaksoftware.com/ronak/riversdk/msg"

	"github.com/monnand/dhkx"
	"go.uber.org/zap"

	_ "github.com/mattn/go-sqlite3"
)

var (
	_ServerKeys ServerKeys
)

// SetConfig
// This function must be called before any other function, otherwise it panics
func (r *River) SetConfig(conf *RiverConfig) {
	r.registerCommandHandlers()
	r.delegates = make(map[int64]RequestDelegate)

	// init delegates
	r.mainDelegate = conf.MainDelegate

	// set loglevel
	log.SetLogLevel(conf.LogLevel)

	// check logger
	if conf.Logger != nil {
		// set logger
		r.logger = conf.Logger
		log.SetLogger(func(logLevel int, msg string) {
			r.logger.Log(logLevel, msg)
		})
	}

	// init UI Executer
	cmd.InitUIExecuter()

	// Initialize Database
	os.MkdirAll(conf.DbPath, os.ModePerm)
	conf.DbPath = strings.TrimRight(conf.DbPath, "/ ")

	// Initialize DB replaced with ORM
	var err error
	err = repo.InitRepo("sqlite3", fmt.Sprintf("%s/%s.db", conf.DbPath, conf.DbID))
	if err != nil {
		log.LOG_Fatal("River::SetConfig() faild to initialize DB context",
			zap.String("Error", err.Error()),
		)
	}

	// // Enable log mode
	// if conf.LogLevel == int(zapcore.DebugLevel) {
	// 	repo.Ctx().LogMode(true)
	// }

	// init riverConfigs this should be after connect to DB
	r.loadSystemConfig()

	// load DeviceToken
	r.loadDeviceToken()

	// Initialize realtime requests
	r.realTimeRequest = map[int64]bool{
		msg.C_MessagesSetTyping: true,
		//msg.C_AuthRecall:        true,
		//msg.C_InitConnect:       true,
		//msg.C_InitCompleteAuth:  true,
	}

	// Initialize filemanager
	filemanager.SetRootFolders(conf.DocumentAudioDirectory, conf.DocumentFileDirectory, conf.DocumentPhotoDirectory, conf.DocumentVideoDirectory)
	filemanager.InitFileManager(r.onFileUploadCompleted, r.onFileProgressChanged, r.onFileDownloadCompleted)

	// Initialize Network Controller
	r.networkCtrl = network.NewNetworkController(
		network.NetworkConfig{
			ServerEndpoint: conf.ServerEndpoint,
			PingTime:       time.Duration(conf.PingTimeSec) * time.Second,
			PongTimeout:    time.Duration(conf.PongTimeoutSec) * time.Second,
		},
	)
	r.networkCtrl.SetNetworkStatusChangedCallback(func(newQuality domain.NetworkStatus) {
		if r.mainDelegate != nil && r.mainDelegate.OnNetworkStatusChanged != nil {
			r.mainDelegate.OnNetworkStatusChanged(int(newQuality))
		}
	})

	// Initialize queueController
	var h domain.DeferredRequestHandler
	if r.mainDelegate != nil {
		h = r.mainDelegate.OnDeferredRequests
	} else {
		h = nil
	}
	if q, err := queue.NewQueueController(r.networkCtrl, conf.QueuePath, h); err != nil {
		log.LOG_Fatal("River::SetConfig() faild to initialize Queue",
			zap.String("Error", err.Error()),
		)
	} else {
		r.queueCtrl = q
	}

	// Initialize Sync Controller
	r.syncCtrl = synchronizer.NewSyncController(
		synchronizer.SyncConfig{
			ConnInfo:    r.ConnInfo,
			NetworkCtrl: r.networkCtrl,
			QueueCtrl:   r.queueCtrl,
		},
	)

	// call external delegate on sync status changed
	r.syncCtrl.SetSyncStatusChangedCallback(func(newStatus domain.SyncStatus) {
		if r.mainDelegate != nil && r.mainDelegate.OnSyncStatusChanged != nil {
			r.mainDelegate.OnSyncStatusChanged(int(newStatus))
		}
	})
	// call external delegate on OnUpdate
	r.syncCtrl.SetOnUpdateCallback(func(constructorID int64, b []byte) {
		if r.mainDelegate != nil && r.mainDelegate.OnUpdates != nil {
			r.mainDelegate.OnUpdates(constructorID, b)
		}
	})

	// Initialize Server Keys
	if jsonBytes, err := ioutil.ReadFile(conf.ServerKeysFilePath); err != nil {
		log.LOG_Fatal("River::SetConfig() faild to open server keys",
			zap.String("Error", err.Error()),
		)
	} else if err := _ServerKeys.UnmarshalJSON(jsonBytes); err != nil {
		log.LOG_Fatal("River::SetConfig() faild to unmarshal server keys",
			zap.String("Error", err.Error()),
		)
	}

	// Initialize River Connection
	log.LOG_Info("River::SetConfig() Load/Create New River Connection")

	if r.ConnInfo.UserID != 0 {
		r.syncCtrl.SetUserID(r.ConnInfo.UserID)
	}

	// Update Network Controller
	r.networkCtrl.SetErrorHandler(r.onGeneralError)
	r.networkCtrl.SetMessageHandler(r.onReceivedMessage)
	r.networkCtrl.SetUpdateHandler(r.onReceivedUpdate)
	r.networkCtrl.SetOnConnectCallback(r.callAuthRecall_RegisterDevice)
	r.networkCtrl.SetAuthorization(r.ConnInfo.AuthID, r.ConnInfo.AuthKey[:])
	filemanager.Ctx().SetAuthorization(r.ConnInfo.AuthID, r.ConnInfo.AuthKey[:])

	filemanager.Ctx().LoadFileStatusQueue()
}

// Get deviceToken
func (r *River) loadDeviceToken() {
	r.DeviceToken = new(msg.AccountRegisterDevice)
	str, err := repo.Ctx().System.LoadString(domain.CN_DEVICE_TOKEN)
	if err != nil {
		log.LOG_Info("River::loadDeviceToken() failed to fetch DeviceToken",
			zap.String("Error", err.Error()),
		)
		return
	}
	err = json.Unmarshal([]byte(str), r.DeviceToken)
	if err != nil {
		log.LOG_Info("River::loadDeviceToken() failed to unmarshal DeviceToken",
			zap.String("Error", err.Error()),
		)
	}
}

func (r *River) callAuthRecall_RegisterDevice() {
	req := msg.AuthRecall{}
	reqBytes, _ := req.Marshal()
	if r.syncCtrl.UserID != 0 {
		// send auth recall until it succeed
		for {
			// this is priority command that should not passed to queue
			// after auth recall answer got back the queue should send its requests in order to get related updates
			err := r.queueCtrl.ExecuteRealtimeCommand(
				uint64(domain.SequentialUniqueID()),
				msg.C_AuthRecall,
				reqBytes,
				nil,
				r.onAuthRecalled,
				true,
				false,
			)
			if err == nil {
				break
			} else {
				time.Sleep(1 * time.Second)
			}
		}

		if r.DeviceToken == nil || r.DeviceToken.Token == "" {
			log.LOG_Info("callAuthRecall_RegisterDevice() Device Token is not set")
			return
		}

		// register device to receive notification
		reqBytes, _ = r.DeviceToken.Marshal()
		for {
			err := r.queueCtrl.ExecuteRealtimeCommand(
				uint64(domain.SequentialUniqueID()),
				msg.C_AccountRegisterDevice,
				reqBytes,
				nil,
				nil,
				true,
				false,
			)
			if err == nil {
				break
			} else {
				time.Sleep(1 * time.Second)
			}
		}
	}
}

// onAuthRecalled update cluster info
func (r *River) onAuthRecalled(m *msg.MessageEnvelope) {
	if m.Constructor == msg.C_AuthRecalled {
		x := new(msg.AuthRecalled)
		err := x.Unmarshal(m.Message)
		if err != nil {
			log.LOG_Warn("onAuthRecalled()", zap.Error(err))
			return
		}

		// save cluster info
		filemanager.SetAvailableClusters(x.AvailableClusters)
	}
}

func (r *River) registerCommandHandlers() {
	r.localCommands = make(map[int64]domain.LocalMessageHandler)
	r.localCommands[msg.C_MessagesGetDialogs] = r.messagesGetDialogs
	r.localCommands[msg.C_MessagesGetDialog] = r.messagesGetDialog
	r.localCommands[msg.C_MessagesGetHistory] = r.messageGetHistory
	r.localCommands[msg.C_MessagesSend] = r.messagesSend
	r.localCommands[msg.C_ClientSendMessageMedia] = r.clientSendMessageMedia
	r.localCommands[msg.C_ContactsGet] = r.contactGet
	r.localCommands[msg.C_MessagesReadHistory] = r.messageReadHistory
	r.localCommands[msg.C_UsersGet] = r.usersGet
	r.localCommands[msg.C_MessagesGet] = r.messagesGet
	r.localCommands[msg.C_AccountUpdateUsername] = r.accountUpdateUsername
	r.localCommands[msg.C_AccountUpdateProfile] = r.accountUpdateProfile
	r.localCommands[msg.C_AccountRegisterDevice] = r.accountRegisterDevice
	r.localCommands[msg.C_AccountUnregisterDevice] = r.accountUnregisterDevice
	r.localCommands[msg.C_AccountSetNotifySettings] = r.accountSetNotifySettings
	r.localCommands[msg.C_GroupsEditTitle] = r.groupsEditTitle
	r.localCommands[msg.C_MessagesClearHistory] = r.messagesClearHistory
	r.localCommands[msg.C_MessagesDelete] = r.messagesDelete
	r.localCommands[msg.C_GroupsAddUser] = r.groupAddUser
	r.localCommands[msg.C_GroupsDeleteUser] = r.groupDeleteUser
	r.localCommands[msg.C_GroupsGetFull] = r.groupsGetFull
	r.localCommands[msg.C_GroupsUpdateAdmin] = r.groupUpdateAdmin
	r.localCommands[msg.C_ContactsImport] = r.contactsImport

}

// Start
func (r *River) Start() error {
	// Start Controllers
	if err := r.networkCtrl.Start(); err != nil {
		log.LOG_Debug("River::Start()",
			zap.String("Error", err.Error()),
		)
		return err
	}
	r.queueCtrl.Start()
	r.syncCtrl.Start()

	// Connect to Server
	go r.networkCtrl.Connect()

	return nil
}

// Stop
func (r *River) Stop() {
	// Disconnect from Server
	r.networkCtrl.Disconnect()

	// Stop Controllers
	r.syncCtrl.Stop()
	r.queueCtrl.Stop()
	r.networkCtrl.Stop()
	cmd.GetUIExecuter().Stop()

	// Close database connection
	err := repo.Ctx().Close()
	log.LOG_Debug("River::Stop() faild to close DB context",
		zap.String("Error", err.Error()),
	)
}

// take a copy of commandBytes b4 IOS/Android GC/OS collect/alter them
func deepCopy(commandBytes []byte) []byte {
	length := len(commandBytes)
	buff := make([]byte, length)
	copy(buff, commandBytes)
	// Deep Copy
	// for i := 0; i < length; i++ {
	// 	buff[i] = commandBytes[i]
	// }
	return buff
}

// ExecuteCommand
// This is a wrapper function to pass the request to the queueController, to be passed to networkController for final
// delivery to the server.
func (r *River) ExecuteCommand(constructor int64, commandBytes []byte, delegate RequestDelegate, blockingMode, serverForce bool) (requestID int64, err error) {
	// deleteMe
	cmdID := fmt.Sprintf("%v : ", time.Now().UnixNano())
	log.LOG_Debug(cmdID + "SDK::ExecuteCommand() 1 ExecuteCommand Started req:" + msg.ConstructorNames[constructor])
	defer log.LOG_Debug(cmdID + "SDK::ExecuteCommand() 7 ExecuteCommand Ended req:" + msg.ConstructorNames[constructor])

	commandBytesDump := deepCopy(commandBytes)

	if _, ok := msg.ConstructorNames[constructor]; !ok {
		return 0, domain.ErrInvalidConstructor
	}
	waitGroup := new(sync.WaitGroup)
	requestID = domain.SequentialUniqueID()

	log.LOG_Debug("River::ExecuteCommand()",
		zap.String("Constructor", msg.ConstructorNames[constructor]),
	)

	// if function is in blocking mode set the waitGroup to block until the job is done, otherwise
	// save 'delegate' into delegates list to be fetched later.
	if blockingMode {
		log.LOG_Debug(cmdID + "SDK::ExecuteCommand() 2 waitGroup.Add(1) / defer")
		waitGroup.Add(1)
		defer waitGroup.Wait()
	} else if delegate != nil {
		r.delegateMutex.Lock()
		r.delegates[requestID] = delegate
		r.delegateMutex.Unlock()
	}
	timeoutCallback := func() {

		log.LOG_Debug(cmdID + "SDK::ExecuteCommand() 3 timeout called")
		if blockingMode {
			defer waitGroup.Done()
		}
		err = domain.ErrRequestTimeout
		delegate.OnTimeout(err)
		r.releaseDelegate(requestID)
		log.LOG_Debug(cmdID + "SDK::ExecuteCommand() 4 timeout ended")

	}
	successCallback := func(envelope *msg.MessageEnvelope) {

		log.LOG_Debug(cmdID + "SDK::ExecuteCommand() 5 succes called")
		if blockingMode {
			defer waitGroup.Done()
		}
		b, _ := envelope.Marshal()
		delegate.OnComplete(b)
		r.releaseDelegate(requestID)

		log.LOG_Debug(cmdID + "SDK::ExecuteCommand() 6 success ended")

	}
	if !serverForce {
		_, isRealTimeRequest := r.realTimeRequest[constructor]
		if isRealTimeRequest {
			err := r.queueCtrl.ExecuteRealtimeCommand(
				uint64(requestID),
				constructor,
				commandBytesDump,
				timeoutCallback,
				successCallback,
				blockingMode,
				true,
			)
			if err != nil && delegate != nil && delegate.OnTimeout != nil {
				delegate.OnTimeout(err)
			}
		} else {
			// else pass the request to queue
			_, ok := r.localCommands[constructor]
			if ok {

				execBlock := func() {
					r.executeLocalCommand(
						uint64(requestID),
						constructor,
						commandBytesDump,
						timeoutCallback,
						successCallback,
					)
				}
				if blockingMode {
					execBlock()
				} else {
					go execBlock()
				}

			} else {
				r.executeRemoteCommand(
					uint64(requestID),
					constructor,
					commandBytesDump,
					timeoutCallback,
					successCallback,
				)
			}
		}
	} else {
		r.executeRemoteCommand(
			uint64(requestID),
			constructor,
			commandBytesDump,
			timeoutCallback,
			successCallback,
		)
	}

	return
}

func (r *River) executeLocalCommand(requestID uint64, constructor int64, commandBytes []byte, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	log.LOG_Debug("River::executeLocalCommand()",
		zap.String("Constructor", msg.ConstructorNames[constructor]),
	)

	in := new(msg.MessageEnvelope)
	out := new(msg.MessageEnvelope)
	in.Constructor = constructor
	in.Message = commandBytes
	in.RequestID = requestID
	out.RequestID = in.RequestID
	// double check
	if applier, ok := r.localCommands[constructor]; ok {
		applier(in, out, timeoutCB, successCB)
	}
}

func (r *River) executeRemoteCommand(requestID uint64, constructor int64, commandBytes []byte, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	log.LOG_Debug("River::executeRemoteCommand()",
		zap.String("Constructor", msg.ConstructorNames[constructor]),
	)
	r.queueCtrl.ExecuteCommand(requestID, constructor, commandBytes, timeoutCB, successCB, true)
}

func (r *River) releaseDelegate(requestID int64) {
	log.LOG_Debug("River::releaseDelegate()",
		zap.Int64("RequestID", requestID),
	)
	r.delegateMutex.Lock()
	if _, ok := r.delegates[requestID]; ok {
		delete(r.delegates, requestID)
	}
	r.delegateMutex.Unlock()
}

// CreateAuthKey
// This function creates an AuthID and AuthKey to be used for transporting messages between client and server
func (r *River) CreateAuthKey() (err error) {
	log.LOG_Debug("River::CreateAuthKey()")
	// wait untill network connects
	for r.networkCtrl.Quality() == domain.DISCONNECTED || r.networkCtrl.Quality() == domain.CONNECTING {
		time.Sleep(200)
	}

	var clientNonce, serverNonce, serverPubFP, serverDHFP, serverPQ uint64
	// 1. Send InitConnect to Server
	req1 := new(msg.InitConnect)
	req1.ClientNonce = uint64(domain.SequentialUniqueID())
	req1Bytes, _ := req1.Marshal()
	waitGroup := new(sync.WaitGroup)
	waitGroup.Add(1)

	log.LOG_Info("River::CreateAuthKey() 1st Step Started :: InitConnect")

	r.executeRemoteCommand(
		//r.executeRealtimeCommand(
		uint64(domain.SequentialUniqueID()),
		msg.C_InitConnect,
		req1Bytes,
		func() {
			defer waitGroup.Done()
			err = domain.ErrRequestTimeout
		},
		func(res *msg.MessageEnvelope) {
			defer waitGroup.Done()
			log.LOG_Debug("River::CreateAuthKey() Success Callback Called")
			switch res.Constructor {
			case msg.C_InitResponse:
				x := new(msg.InitResponse)
				err = x.Unmarshal(res.Message)
				if err != nil {
					log.LOG_Debug("River::CreateAuthKey() Success Callback",
						zap.String("Error", err.Error()),
					)
				}
				clientNonce = x.ClientNonce
				serverNonce = x.ServerNonce
				serverPubFP = x.RSAPubKeyFingerPrint
				serverDHFP = x.DHGroupFingerPrint
				serverPQ = x.PQ
				log.LOG_Debug("River::CreateAuthKey() InitResponse Received",
					zap.Uint64("ServerNonce", serverNonce),
					zap.Uint64("ClientNounce", clientNonce),
					zap.Uint64("ServerDhFingerPrint", serverDHFP),
					zap.Uint64("ServerFingerPrint", serverPubFP),
				)
			case msg.C_Error:
				err = domain.ServerError(res.Message)
			default:
				err = domain.ErrInvalidConstructor
			}
		},
	)

	// Wait for 1st step to complete
	waitGroup.Wait()
	if err != nil {
		log.LOG_Debug("River::CreateAuthKey() InitConnect",
			zap.String("Error", err.Error()),
		)
		return
	} else {
		log.LOG_Info("River::CreateAuthKey() 1st Step Finished")
	}

	// 2. Send InitCompleteAuth
	req2 := new(msg.InitCompleteAuth)
	req2.ServerNonce = serverNonce
	req2.ClientNonce = clientNonce

	// Generate DH Pub Key
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
	log.LOG_Debug("River::CreateAuthKey() PQ Split",
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
		log.LOG_Debug("River::CreateAuthKey() -> EncryptPKCS1v15()",
			zap.String("Error", err.Error()),
		)
	}
	req2.EncryptedPayload = encrypted
	req2Bytes, _ := req2.Marshal()

	waitGroup.Add(1)
	log.LOG_Info("River::CreateAuthKey() 2nd Step Started :: InitConnect")
	r.executeRemoteCommand(
		//r.executeRealtimeCommand(
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
				x.Unmarshal(res.Message)
				switch x.Status {
				case msg.InitAuthCompleted_OK:
					serverDhKey, err := dh.ComputeKey(dhkx.NewPublicKey(x.ServerDHPubKey), clientDhKey)
					if err != nil {
						log.LOG_Debug("River::CreateAuthKey() -> ComputeKey()",
							zap.String("Error", err.Error()),
						)
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
				r.ConnInfo.Save()
				r.networkCtrl.SetAuthorization(r.ConnInfo.AuthID, r.ConnInfo.AuthKey[:])
				filemanager.Ctx().SetAuthorization(r.ConnInfo.AuthID, r.ConnInfo.AuthKey[:])
			case msg.C_Error:
				err = domain.ServerError(res.Message)
				return
			default:
				err = domain.ErrInvalidConstructor
				return
			}
		},
	)

	// Wait for 2nd step to complete
	waitGroup.Wait()

	// double set AuthID
	r.networkCtrl.SetAuthorization(r.ConnInfo.AuthID, r.ConnInfo.AuthKey[:])
	filemanager.Ctx().SetAuthorization(r.ConnInfo.AuthID, r.ConnInfo.AuthKey[:])
	// inform external UI that authKey generated
	if r.mainDelegate != nil {
		if r.mainDelegate.OnAuthKeyCreated != nil {
			r.mainDelegate.OnAuthKeyCreated(r.ConnInfo.AuthID)
		}
	}

	// call authRecall to receive data from websocket
	r.callAuthRecall_RegisterDevice()

	return
}

func (r *River) onGeneralError(e *msg.Error) {
	// TODO:: calll external handler
	log.LOG_Info("River::onGeneralError()",
		zap.String("Code", e.Code),
		zap.String("Item", e.Items),
	)

	if r.mainDelegate != nil && r.mainDelegate.OnGeneralError != nil {
		buff, _ := e.Marshal()
		r.mainDelegate.OnGeneralError(buff)
	}
}

// AddRealTimeRequest
func (r *River) AddRealTimeRequest(constructor int64) {
	r.realTimeRequest[constructor] = true
}

// RemoveRealTimeRequest
func (r *River) RemoveRealTimeRequest(constructor int64) {
	delete(r.realTimeRequest, constructor)
}

// called when network flushes received messages
func (r *River) onReceivedMessage(msgs []*msg.MessageEnvelope) {

	// sort messages by reauestID
	sort.Slice(msgs, func(i, j int) bool {
		return msgs[i].RequestID < msgs[j].RequestID
	})

	// sync localDB with responses in the background
	go r.syncCtrl.MessageHandler(msgs)

	// check requestCallbacks and call callbacks
	count := len(msgs)
	for idx := 0; idx < count; idx++ {
		m := msgs[idx]
		cb := domain.GetRequestCallback(m.RequestID)
		if cb != nil {
			// if there was any listener maybe request already timedout
			log.LOG_Warn("River::onReceivedMessage() Callback Found")

			select {
			case cb.ResponseChannel <- m:
				log.LOG_Warn("River::onReceivedMessage() passed to callback listener")
			default:
				log.LOG_Warn("River::onReceivedMessage() there is no callback listener")
			}
			domain.RemoveRequestCallback(m.RequestID)
		} else {
			log.LOG_Debug("River::onReceivedMessage() callback does not exists",
				zap.Uint64(domain.LK_REQUEST_ID, m.RequestID),
			)
		}
	}

}

// called when network flushes received updates
func (r *River) onReceivedUpdate(upds []*msg.UpdateContainer) {
	updateContainer := new(msg.UpdateContainer)

	log.LOG_Debug("SDK::onReceivedUpdate()",
		zap.Int("Received Container Count :", len(upds)),
	)

	minID := int64(^uint64(0) >> 1)
	maxID := int64(0)

	// remove duplicated users and updates and pass it to sync controller
	userIDs := domain.MInt64B{}
	groupIDs := domain.MInt64B{}
	updateIDs := domain.MInt64B{}
	users := make([]*msg.User, 0)
	groups := make([]*msg.Group, 0)
	updates := make([]*msg.UpdateEnvelope, 0)

	currentUpdateID := r.syncCtrl.UpdateID()
	for _, val := range upds {
		if val.MinUpdateID < minID && val.MinUpdateID > 0 {
			minID = val.MinUpdateID
		}
		if val.MaxUpdateID > maxID && val.MaxUpdateID > 0 {
			maxID = val.MaxUpdateID
		}

		for _, u := range val.Updates {
			if u.UpdateID > 0 && u.UpdateID <= currentUpdateID {
				log.LOG_Debug("SDK::onReceivedUpdate() XXXXXXXXXXXXXXXXXXXXXXXXXXXXX Outdated update ",
					zap.Int64("CurrentUpdateID", currentUpdateID),
					zap.Int64("UpdateID", u.UpdateID),
				)
				continue
			}
			if _, ok := updateIDs[u.UpdateID]; !ok {
				updateIDs[u.UpdateID] = true
				updates = append(updates, u)
			}
		}
		// get distinct users
		for _, u := range val.Users {
			if _, ok := userIDs[u.ID]; !ok {
				userIDs[u.ID] = true
				users = append(users, u)
			}
		}
		// get distinct groups
		for _, g := range val.Groups {
			if _, ok := groupIDs[g.ID]; !ok {
				groupIDs[g.ID] = true
				groups = append(groups, g)
			}
		}

	}
	// on typing min max is equal to zero so meh
	if minID > maxID {
		minID = 0
	}
	log.LOG_Debug("SDK::onReceivedUpdate()",
		zap.Int("Received Updates Count :", len(updates)),
		zap.Int64("UpdateID :", r.syncCtrl.UpdateID()),
		zap.Int64("MaxID :", maxID),
		zap.Int64("MinID :", minID),
	)

	// check max UpdateID if its greater than snapshot sync threshold discard recived updates and execute sanpshot sync
	if maxID-r.syncCtrl.UpdateID() > domain.SnapshotSync_Threshold {
		log.LOG_Debug("SDK::onReceivedUpdate() snapshot threshold reached")
		r.syncCtrl.CheckSyncState()
		return
	}

	// No need to wait here till DB gets synced cuz UI will have required data
	go func(u []*msg.User, g []*msg.Group) {
		// Save Groups
		repo.Ctx().Groups.SaveMany(g)
		// Save Users
		repo.Ctx().Users.SaveMany(u)
	}(users, groups)

	// sort updates
	sort.Slice(updates, func(i, j int) bool {
		return updates[i].UpdateID < updates[j].UpdateID
	})

	updateContainer.Updates = updates
	updateContainer.Users = users
	updateContainer.Groups = groups
	updateContainer.Length = int32(len(updates))
	updateContainer.MinUpdateID = minID
	updateContainer.MaxUpdateID = maxID
	r.syncCtrl.UpdateHandler(updateContainer)
}

func (r *River) PrintDebuncerStatus() {
	log.LOG_Debug("SDK::PrintDebuncerStatus()")
	r.networkCtrl.PrintDebuncerStatus()
}

func (r *River) onFileProgressChanged(messageID, position, totalSize int64, isDownload bool) {
	percent := float64(position) / float64(totalSize) * float64(100)
	log.LOG_Warn("onFileProgressChanged()",
		zap.Int64("MsgID", messageID),
		zap.Float64("Percent", percent),
	)

	// Notify UI that upload is completed
	if isDownload {
		if r.mainDelegate.OnDownloadProgressChanged != nil {
			r.mainDelegate.OnDownloadProgressChanged(messageID, position, totalSize, percent)
		}
	} else {
		if r.mainDelegate.OnUploadProgressChanged != nil {
			r.mainDelegate.OnUploadProgressChanged(messageID, position, totalSize, percent)
		}
	}

}

func (r *River) onFileUploadCompleted(messageID, fileID int64, clusterID, totalParts int32, req *msg.ClientSendMessageMedia) {
	log.LOG_Warn("onFileUploadCompleted()",
		zap.Int64("messageID", messageID),
		zap.Int64("fileID", fileID),
	)
	// if total parts are grater than zero it means we actually uploaded new file
	// else the doc was already uploaded we called this just to notify ui that upload finished
	if totalParts > 0 {

		// Create SendMessageMedia Request
		x := new(msg.MessagesSendMedia)
		x.Peer = req.Peer
		x.ClearDraft = req.ClearDraft
		x.MediaType = req.MediaType
		x.RandomID = fileID
		x.ReplyTo = req.ReplyTo

		//
		switch x.MediaType {
		case msg.InputMediaTypeEmpty:
			panic("SDK:onFileUploadCompleted() not implemented")
		case msg.InputMediaTypeUploadedPhoto:
			panic("SDK:onFileUploadCompleted() not implemented")
		case msg.InputMediaTypePhoto:
			panic("SDK:onFileUploadCompleted() not implemented")
		case msg.InputMediaTypeContact:
			panic("SDK:onFileUploadCompleted() not implemented")
		case msg.InputMediaTypeUploadedDocument:

			doc := new(msg.InputMediaUploadedDocument)
			doc.MimeType = req.FileMIME
			doc.Attributes = req.Attributes
			doc.Caption = req.Caption
			doc.File = &msg.InputFile{
				ClusterID:   clusterID,
				FileID:      fileID,
				FileName:    req.FileName,
				MD5Checksum: "",
				TotalParts:  totalParts,
			}
			x.MediaData, _ = doc.Marshal()
		case msg.InputMediaTypeDocument:
			panic("SDK:onFileUploadCompleted() not implemented")
		default:
			panic("SDK:onFileUploadCompleted() invalid input media type")
		}
		reqBuff, _ := x.Marshal()
		requestID := uint64(domain.SequentialUniqueID())
		r.queueCtrl.ExecuteCommand(requestID, msg.C_MessagesSendMedia, reqBuff, nil, nil, false)

	}

	// Notify UI that upload is completed
	if r.mainDelegate.OnUploadProgressChanged != nil {
		r.mainDelegate.OnUploadCompleted(messageID, req.FilePath)
	}
}

func (r *River) onFileDownloadCompleted(messageID int64, filePath string) {

	// Notify UI that download is completed
	if r.mainDelegate.OnDownloadCompleted != nil {
		r.mainDelegate.OnDownloadCompleted(messageID, filePath)
	}
}
