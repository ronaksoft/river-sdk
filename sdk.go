package riversdk

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"git.ronaksoftware.com/ronak/riversdk/configs"

	"git.ronaksoftware.com/ronak/riversdk/cmd"

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
	_ServerKeys configs.ServerKeys
)

// SetConfig
// This function must be called before any other function, otherwise it panics
func (r *River) SetConfig(conf *RiverConfig) {
	r.registerCommandHandlers()
	r.delegates = make(map[int64]RequestDelegate)

	// init delegates
	setMainDelegate(conf.MainDelegate)

	// init UI Executer
	cmd.InitUIExecuter()

	// Initialize Database
	os.MkdirAll(conf.DbPath, os.ModePerm)
	conf.DbPath = strings.TrimRight(conf.DbPath, "/ ")

	// Initialize DB replaced with ORM
	var err error
	err = repo.InitRepo("sqlite3", fmt.Sprintf("%s/%s.db", conf.DbPath, conf.DbID))
	if err != nil {
		log.LOG.Fatal("River::SetConfig() faild to initialize DB context",
			zap.String("Error", err.Error()),
		)
	}

	// init riverConfigs this should be after connect to DB
	configs.Get()

	// Initialize realtime requests
	r.realTimeRequest = map[int64]bool{
		msg.C_MessagesSetTyping: true,
		//msg.C_AuthRecall:        true,
		//msg.C_InitConnect:       true,
		//msg.C_InitCompleteAuth:  true,
	}

	// Enable log mode
	//repo.Ctx().LogMode(true)

	// Initialize requests that should not passed to queueController

	// Initialize Network Controller
	r.networkCtrl = network.NewNetworkController(
		network.NetworkConfig{
			ServerEndpoint: conf.ServerEndpoint,
			PingTime:       time.Duration(conf.PingTimeSec) * time.Second,
			PongTimeout:    time.Duration(conf.PongTimeoutSec) * time.Second,
		},
	)
	r.networkCtrl.SetNetworkStatusChangedCallback(func(newQuality domain.NetworkStatus) {
		if getMainDelegate() != nil && getMainDelegate().OnNetworkStatusChanged != nil {
			getMainDelegate().OnNetworkStatusChanged(int(newQuality))
		}
	})

	// Initialize queueController
	var h domain.DeferredRequestHandler
	if getMainDelegate() != nil {
		h = getMainDelegate().OnDeferredRequests
	} else {
		h = nil
	}
	if q, err := queue.NewQueueController(r.networkCtrl, conf.QueuePath, h); err != nil {
		log.LOG.Fatal("River::SetConfig() faild to initialize Queue",
			zap.String("Error", err.Error()),
		)
	} else {
		r.queueCtrl = q
	}

	// Initialize Sync Controller
	r.syncCtrl = synchronizer.NewSyncController(
		synchronizer.SyncConfig{
			NetworkCtrl: r.networkCtrl,
			QueueCtrl:   r.queueCtrl,
		},
	)

	// call external delegate on sync status changed
	r.syncCtrl.SetSyncStatusChangedCallback(func(newStatus domain.SyncStatus) {
		if getMainDelegate() != nil && getMainDelegate().OnSyncStatusChanged != nil {
			getMainDelegate().OnSyncStatusChanged(int(newStatus))
		}
	})
	// call external delegate on OnUpdate
	r.syncCtrl.SetOnUpdateCallback(func(constructor int64, buff []byte) {
		if getMainDelegate() != nil && getMainDelegate().OnUpdates != nil {
			getMainDelegate().OnUpdates(constructor, buff)
		}
	})

	// Initialize Server Keys
	if jsonBytes, err := ioutil.ReadFile(conf.ServerKeysFilePath); err != nil {
		log.LOG.Fatal("River::SetConfig() faild to open server keys",
			zap.String("Error", err.Error()),
		)
	} else if err := _ServerKeys.UnmarshalJSON(jsonBytes); err != nil {
		log.LOG.Fatal("River::SetConfig() faild to unmarshal server keys",
			zap.String("Error", err.Error()),
		)
	}

	// Initialize River Connection
	log.LOG.Info("River::SetConfig() Load/Create New River Connection")

	if configs.Get().UserID != 0 {
		r.syncCtrl.SetUserID(configs.Get().UserID)
	}

	// Update Network Controller
	r.networkCtrl.SetErrorHandler(r.onGeneralError)
	r.networkCtrl.SetMessageHandler(r.onReceivedMessage)
	r.networkCtrl.SetUpdateHandler(r.onReceivedUpdate)
	r.networkCtrl.SetOnConnectCallback(r.callAuthRecall)
	r.networkCtrl.SetAuthorization(configs.Get().AuthID, configs.Get().AuthKey[:])
}

func (r *River) callAuthRecall() {
	req := msg.AuthRecall{}
	reqBytes, _ := req.Marshal()
	if r.syncCtrl.UserID != 0 {
		// send auth recall until it succeed
		for {
			// this is priority command that should not passed to queue
			// after auth recall answer got back the queue should send its requests in order to get related updates
			err := r.queueCtrl.ExecuteRealtimeCommand(
				domain.RandomUint64(),
				msg.C_AuthRecall,
				reqBytes,
				nil,
				nil,
				true,
			)
			if err == nil {
				break
			} else {
				time.Sleep(1 * time.Second)
			}
		}
	}
}

func (r *River) registerCommandHandlers() {
	r.localCommands = make(map[int64]domain.LocalMessageHandler)
	r.localCommands[msg.C_MessagesGetDialogs] = r.messagesGetDialogs
	r.localCommands[msg.C_MessagesGetHistory] = r.messageGetHistory
	r.localCommands[msg.C_MessagesSend] = r.messagesSend
	r.localCommands[msg.C_ContactsGet] = r.contactGet
	r.localCommands[msg.C_MessagesReadHistory] = r.messageReadHistory
	r.localCommands[msg.C_UsersGet] = r.usersGet
	r.localCommands[msg.C_MessagesGet] = r.messagesGet
}

// Start
func (r *River) Start() error {
	// Start Controllers
	if err := r.networkCtrl.Start(); err != nil {
		log.LOG.Debug("River::Start()",
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
	log.LOG.Debug("River::Stop() faild to close DB context",
		zap.String("Error", err.Error()),
	)
}
func (r *River) deepCopy(commandBytes []byte) []byte {
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
func (r *River) ExecuteCommand(constructor int64, commandBytes []byte, delegate RequestDelegate, blockingMode bool) (requestID int64, err error) {
	commandBytesDump := r.deepCopy(commandBytes)

	if _, ok := msg.ConstructorNames[constructor]; !ok {
		return 0, domain.ErrInvalidConstructor
	}
	waitGroup := new(sync.WaitGroup)
	requestID = domain.RandomInt63()

	log.LOG.Debug("River::ExecuteCommand()",
		zap.String("Constructor", msg.ConstructorNames[constructor]),
	)

	// if function is in blocking mode set the waitGroup to block until the job is done, otherwise
	// save 'delegate' into delegates list to be fetched later.
	if blockingMode {
		waitGroup.Add(1)
		defer waitGroup.Wait()
	} else if delegate != nil {
		r.delegateMutex.Lock()
		r.delegates[requestID] = delegate
		r.delegateMutex.Unlock()
	}
	timeoutCallback := func() {
		if blockingMode {
			defer waitGroup.Done()
		}
		err = domain.ErrRequestTimeout
		delegate.OnTimeout(err)
		r.releaseDelegate(requestID)
	}
	successCallback := func(envelope *msg.MessageEnvelope) {
		if blockingMode {
			defer waitGroup.Done()
		}
		b, _ := envelope.Marshal()
		delegate.OnComplete(b)
		r.releaseDelegate(requestID)
	}

	_, isRealTimeRequest := r.realTimeRequest[constructor]
	if isRealTimeRequest {
		err := r.queueCtrl.ExecuteRealtimeCommand(
			uint64(requestID),
			constructor,
			commandBytesDump,
			timeoutCallback,
			successCallback,
			blockingMode,
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

	return
}

func (r *River) executeLocalCommand(requestID uint64, constructor int64, commandBytes []byte, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	log.LOG.Debug("River::executeLocalCommand()",
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
	log.LOG.Debug("River::executeRemoteCommand()",
		zap.String("Constructor", msg.ConstructorNames[constructor]),
	)
	r.queueCtrl.ExecuteCommand(requestID, constructor, commandBytes, timeoutCB, successCB)
}

func (r *River) releaseDelegate(requestID int64) {
	log.LOG.Debug("River::releaseDelegate()",
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
	log.LOG.Debug("River::CreateAuthKey()")
	// wait untill network connects
	for r.networkCtrl.Quality() == domain.DISCONNECTED || r.networkCtrl.Quality() == domain.CONNECTING {
		time.Sleep(200)
	}

	var clientNonce, serverNonce, serverPubFP, serverDHFP, serverPQ uint64
	// 1. Send InitConnect to Server
	req1 := new(msg.InitConnect)
	req1.ClientNonce = domain.RandomUint64()
	req1Bytes, _ := req1.Marshal()
	waitGroup := new(sync.WaitGroup)
	waitGroup.Add(1)

	log.LOG.Info("River::CreateAuthKey() 1st Step Started :: InitConnect")

	r.executeRemoteCommand(
		//r.executeRealtimeCommand(
		domain.RandomUint64(),
		msg.C_InitConnect,
		req1Bytes,
		func() {
			defer waitGroup.Done()
			err = domain.ErrRequestTimeout
		},
		func(res *msg.MessageEnvelope) {
			defer waitGroup.Done()
			log.LOG.Debug("River::CreateAuthKey() Success Callback Called")
			switch res.Constructor {
			case msg.C_InitResponse:
				x := new(msg.InitResponse)
				err = x.Unmarshal(res.Message)
				if err != nil {
					log.LOG.Debug("River::CreateAuthKey() Success Callback",
						zap.String("Error", err.Error()),
					)
				}
				clientNonce = x.ClientNonce
				serverNonce = x.ServerNonce
				serverPubFP = x.RSAPubKeyFingerPrint
				serverDHFP = x.DHGroupFingerPrint
				serverPQ = x.PQ
				log.LOG.Debug("River::CreateAuthKey() InitResponse Received",
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
		log.LOG.Debug("River::CreateAuthKey() InitConnect",
			zap.String("Error", err.Error()),
		)
		return
	} else {
		log.LOG.Info("River::CreateAuthKey() 1st Step Finished")
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
	log.LOG.Debug("River::CreateAuthKey() PQ Split",
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
		log.LOG.Debug("River::CreateAuthKey() -> EncryptPKCS1v15()",
			zap.String("Error", err.Error()),
		)
	}
	req2.EncryptedPayload = encrypted
	req2Bytes, _ := req2.Marshal()

	waitGroup.Add(1)
	log.LOG.Info("River::CreateAuthKey() 2nd Step Started :: InitConnect")
	r.executeRemoteCommand(
		//r.executeRealtimeCommand(
		domain.RandomUint64(),
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
						log.LOG.Debug("River::CreateAuthKey() -> ComputeKey()",
							zap.String("Error", err.Error()),
						)
						return
					}
					// r.ConnInfo.AuthKey = serverDhKey.Bytes()
					copy(configs.Get().AuthKey[:], serverDhKey.Bytes())
					authKeyHash, _ := domain.Sha256(configs.Get().AuthKey[:])
					configs.Get().AuthID = int64(binary.LittleEndian.Uint64(authKeyHash[24:32]))

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
				configs.Get().Save()
				r.networkCtrl.SetAuthorization(configs.Get().AuthID, configs.Get().AuthKey[:])
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

	// inform external UI that authKey generated
	if getMainDelegate() != nil {
		if getMainDelegate().OnAuthKeyCreated != nil {
			getMainDelegate().OnAuthKeyCreated(configs.Get().AuthID)
		}
	}

	// call authRecall to receive data from websocket
	r.callAuthRecall()

	return
}

// CancelRequests
func (r *River) CancelRequest(requestID int64) {

	// Remove delegate
	r.delegateMutex.Lock()
	delete(r.delegates, int64(requestID))
	r.delegateMutex.Unlock()

	// Remove callback
	domain.RemoveRequestCallback(uint64(requestID))

	// Remove from goque levelDB
	// the goque pkg does not support this
	r.queueCtrl.CancelRequest(requestID)

}

// DeletePendingMessage
func (r *River) DeletePendingMessage(id int64) (isSuccess bool) {
	err := repo.Ctx().PendingMessages.DeletePendingMessage(id)
	isSuccess = err == nil
	return
}

func (r *River) RetryPendingMessage(id int64) (isSuccess bool) {
	pmsg, err := repo.Ctx().PendingMessages.GetPendingMessageByID(id)
	if err != nil {
		log.LOG.Debug("River::RetryPendingMessage()",
			zap.String("GetPendingMessageByID", err.Error()),
		)
		isSuccess = false
		return
	}
	req := new(msg.MessagesSend)
	pmsg.MapToMessageSend(req)

	buff, _ := req.Marshal()
	r.queueCtrl.ExecuteCommand(uint64(req.RandomID), msg.C_MessagesSend, buff, nil, nil)
	isSuccess = true
	log.LOG.Debug("River::RetryPendingMessage() Request enqueued")

	return
}

func (r *River) GetNetworkStatus() int32 {
	return int32(r.networkCtrl.Quality())
}

func (r *River) GetSyncStatus() int32 {

	log.LOG.Debug("River::GetSyncStatus()",
		zap.String("syncStatus", domain.SyncStatusName[r.syncCtrl.Status()]),
	)
	return int32(r.syncCtrl.Status())
}

func (r *River) Logout() (int64, error) {

	dataDir, err := r.queueCtrl.DropQueue()

	if err != nil {
		log.LOG.Debug("River::Logout() failed to drop queue",
			zap.Error(err),
		)
	}

	// drop and recreate database
	err = repo.Ctx().ReinitiateDatabase()
	if err != nil {
		log.LOG.Debug("River::Logout() failed to re initiate database",
			zap.Error(err),
		)
	}

	// open queue
	err = r.queueCtrl.OpenQueue(dataDir)
	if err != nil {
		log.LOG.Debug("River::Logout() failed to re open queue",
			zap.Error(err),
		)
	}

	// reset connection info
	configs.Clear()

	// TODO : send logout request to server
	requestID := domain.RandomInt63()
	timeoutCallback := func() {
		err = domain.ErrRequestTimeout
		r.releaseDelegate(requestID)
	}
	successCallback := func(envelope *msg.MessageEnvelope) {
		r.releaseDelegate(requestID)
	}

	req := new(msg.AuthLogout)
	buff, _ := req.Marshal()
	err = r.queueCtrl.ExecuteRealtimeCommand(uint64(requestID), msg.C_AuthLogout, buff, timeoutCallback, successCallback, true)
	if err != nil {
		r.releaseDelegate(requestID)
	}

	if getMainDelegate() != nil && getMainDelegate().OnSessionClosed != nil {
		getMainDelegate().OnSessionClosed(0)
	}

	return requestID, err
}

func (r *River) onGeneralError(e *msg.Error) {
	// TODO:: calll external handler
	log.LOG.Info("River::onGeneralError()")
	if getMainDelegate() != nil && getMainDelegate().OnGeneralError != nil {
		buff, _ := e.Marshal()
		getMainDelegate().OnGeneralError(buff)
	}
}

// UISettingGet fetch from key/value storage for UI settings
func (r *River) UISettingGet(key string) string {
	val, err := repo.Ctx().UISettings.Get(key)
	if err != nil {
		log.LOG.Info("River::UISettingsGet()",
			zap.String("Error", err.Error()),
		)
	}
	return val
}

// UISettingPut save to key/value storage for UI settings
func (r *River) UISettingPut(key, value string) bool {
	err := repo.Ctx().UISettings.Put(key, value)
	if err != nil {
		log.LOG.Info("River::UISettingsPut()",
			zap.String("Error", err.Error()),
		)
	}
	return err == nil
}

// UISettingDelete remove from key/value storage for UI settings
func (r *River) UISettingDelete(key string) bool {
	err := repo.Ctx().UISettings.Delete(key)
	if err != nil {
		log.LOG.Info("River::UISettingsDelete()",
			zap.String("Error", err.Error()),
		)
	}
	return err == nil
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
	for _, m := range msgs {
		cb := domain.GetRequestCallback(m.RequestID)
		if cb != nil {
			// if there was any listener maybe request already timedout
			log.LOG.Warn("River::onReceivedMessage() Callback Found")

			select {
			case cb.ResponseChannel <- m:
				log.LOG.Warn("River::onReceivedMessage() passed to callback listener")
			default:
				log.LOG.Warn("River::onReceivedMessage() there is no callback listener")
			}
			domain.RemoveRequestCallback(m.RequestID)
		} else {
			log.LOG.Debug("River::onReceivedMessage() callback does not exists",
				zap.Uint64(domain.LK_REQUEST_ID, m.RequestID),
			)
		}
	}

}

// called when network flushes received updates
func (r *River) onReceivedUpdate(upds []*msg.UpdateContainer) {
	updateContainer := new(msg.UpdateContainer)

	minID := int64(^uint64(0) >> 1)
	maxID := int64(0)

	// remove duplicated users and updates and pass it to sync controller
	userIDs := domain.MInt64B{}
	updateIDs := domain.MInt64B{}
	users := make([]*msg.User, 0)
	updates := make([]*msg.UpdateEnvelope, 0)
	for _, val := range upds {
		if val.MinUpdateID < minID {
			minID = val.MinUpdateID
		}
		if val.MaxUpdateID > maxID {
			maxID = val.MaxUpdateID
		}

		for _, u := range val.Updates {
			if _, ok := updateIDs[u.UpdateID]; !ok {
				updateIDs[u.UpdateID] = true
				updates = append(updates, u)
			}
		}
		for _, u := range val.Users {
			if _, ok := userIDs[u.ID]; !ok {
				userIDs[u.ID] = true
				users = append(users, u)
			}
		}

	}

	// sort updates
	sort.Slice(updates, func(i, j int) bool {
		return updates[i].UpdateID < updates[j].UpdateID
	})

	updateContainer.Updates = updates
	updateContainer.Users = users
	updateContainer.Length = int32(len(updates))
	updateContainer.MinUpdateID = minID
	updateContainer.MaxUpdateID = maxID
	r.syncCtrl.UpdateHandler(updateContainer)
}
