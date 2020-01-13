package riversdk

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/binary"
	"fmt"
	"git.ronaksoftware.com/ronak/riversdk/msg/chat"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo"
	"git.ronaksoftware.com/ronak/riversdk/pkg/uiexec"
	"github.com/monnand/dhkx"
	"go.uber.org/zap"
	"math/big"
	"sync"
	"time"
)

// ExecuteCommand ...
// This is a wrapper function to pass the request to the queueController, to be passed to networkController for final
// delivery to the server.
func (r *River) ExecuteCommand(constructor int64, commandBytes []byte, delegate RequestDelegate) (requestID int64, err error) {
	if _, ok := msg.ConstructorNames[constructor]; !ok {
		return 0, domain.ErrInvalidConstructor
	}

	commandBytesDump := deepCopy(commandBytes)

	waitGroup := new(sync.WaitGroup)
	requestID = domain.SequentialUniqueID()
	logs.Debug("River executes command",
		zap.String("Constructor", msg.ConstructorNames[constructor]),
	)

	blockingMode := delegate.Flags() & RequestBlocking != 0
	serverForce := delegate.Flags() & RequestServerForced != 0

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
		r.releaseDelegate(uint64(requestID))
		if blockingMode {
			waitGroup.Done()
		}
	}

	// Success Callback
	successCallback := func(envelope *msg.MessageEnvelope) {
		b, _ := envelope.Marshal()
		delegate.OnComplete(b)
		r.releaseDelegate(uint64(requestID))
		if blockingMode {
			waitGroup.Done()
		}
	}

	// If this request must be sent to the server then executeRemoteCommand
	if serverForce {
		executeRemoteCommand(r, uint64(requestID), constructor, commandBytesDump, timeoutCallback, successCallback)
		return
	}

	// If the constructor is a local command then
	applier, ok := r.localCommands[constructor]
	if ok {
		if blockingMode {
			executeLocalCommand(applier, uint64(requestID), constructor, commandBytesDump, timeoutCallback, successCallback)
		} else {
			go executeLocalCommand(applier, uint64(requestID), constructor, commandBytesDump, timeoutCallback, successCallback)
		}
		return
	}

	// If we reached here, then execute the remote commands
	executeRemoteCommand(r, uint64(requestID), constructor, commandBytesDump, timeoutCallback, successCallback)

	return
}
func executeLocalCommand(applier domain.LocalMessageHandler, requestID uint64, constructor int64, commandBytes []byte, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	logs.Debug("River executes local command",
		zap.String("Constructor", msg.ConstructorNames[constructor]),
	)

	in := new(msg.MessageEnvelope)
	out := new(msg.MessageEnvelope)
	in.Constructor = constructor
	in.Message = commandBytes
	in.RequestID = requestID
	out.RequestID = in.RequestID
	applier(in, out, timeoutCB, successCB)
}
func executeRemoteCommand(r *River, requestID uint64, constructor int64, commandBytes []byte, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	logs.Debug("River executes remote command",
		zap.String("Constructor", msg.ConstructorNames[constructor]),
	)

	blocking := false
	dontWaitForNetwork := false
	d, ok := r.getDelegate(requestID)
	if ok {
		blocking = d.Flags() & RequestBlocking != 0
		dontWaitForNetwork = d.Flags() & RequestDontWaitForNetwork != 0
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
						uiexec.Ctx().Exec(func() { reqCB.TimeoutCallback() })
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
		r.queueCtrl.RealtimeCommand(requestID, constructor, commandBytes, timeoutCB, successCB, blocking,true)
	} else {
		r.queueCtrl.EnqueueCommand(requestID, constructor, commandBytes, timeoutCB, successCB, true)
	}


}
func deepCopy(commandBytes []byte) []byte {
	// Takes a copy of commandBytes b4 IOS/Android GC/OS collect/alter them
	length := len(commandBytes)
	buff := make([]byte, length)
	copy(buff, commandBytes)
	return buff
}

func (r *River) releaseDelegate(requestID uint64) {
	logs.Debug("River releases delegate",
		zap.Uint64("RequestID", requestID),
	)
	r.delegateMutex.Lock()
	if _, ok := r.delegates[requestID]; ok {
		delete(r.delegates, requestID)
	}
	r.delegateMutex.Unlock()
}

func (r *River) getDelegate(requestID uint64) (RequestDelegate, bool){
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
	r.networkCtrl.WaitForNetwork()

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
		r,
		// r.executeRealtimeCommand(
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
	req := new(msg.MessagesSend)
	req.Body = pmsg.Body
	req.Peer = new(msg.InputPeer)
	req.Peer.AccessHash = pmsg.AccessHash
	req.Peer.ID = pmsg.PeerID
	req.Peer.Type = msg.PeerType(pmsg.PeerType)
	req.RandomID = pmsg.RequestID
	req.ReplyTo = pmsg.ReplyTo
	req.ClearDraft = pmsg.ClearDraft
	req.Entities = pmsg.Entities
	buff, _ := req.Marshal()
	r.queueCtrl.EnqueueCommand(uint64(req.RandomID), msg.C_MessagesSend, buff, nil, nil, true)

	logs.Debug("River::RetryPendingMessage() Request enqueued")
	return true
}

// GetSyncStatus returns SyncController status
func (r *River) GetSyncStatus() int32 {
	return int32(r.syncCtrl.GetSyncStatus())
}

// Logout drop queue & database , etc ...
func (r *River) Logout(notifyServer bool, reason int) error {
	// unregister device if token exist
	if notifyServer {
		if r.DeviceToken != nil {
			req := new(msg.AccountUnregisterDevice)
			req.Token = r.DeviceToken.Token
			req.TokenType = int32(r.DeviceToken.TokenType)
			reqBytes, _ := req.Marshal()
			r.queueCtrl.RealtimeCommand(
				uint64(domain.SequentialUniqueID()),
				msg.C_AccountUnregisterDevice,
				reqBytes, nil, nil, true, false,
			)
		}
		// send logout request to server
		requestID := domain.SequentialUniqueID()
		req := new(msg.AuthLogout)
		buff, _ := req.Marshal()
		r.queueCtrl.RealtimeCommand(uint64(requestID), msg.C_AuthLogout, buff, nil, nil, true, false)
	}

	if r.mainDelegate != nil {
		r.mainDelegate.OnSessionClosed(reason)
	}

	// Stop Controllers
	r.networkCtrl.Stop()
	r.syncCtrl.Stop()
	r.queueCtrl.Stop()

	// Close UI Executor
	uiexec.Ctx().Stop()

	repo.DropAll()
	r.ResetAuthKey()
	r.syncCtrl.ResetIDs()
	r.ConnInfo.FirstName = ""
	r.ConnInfo.LastName = ""
	r.ConnInfo.Phone = ""
	r.ConnInfo.UserID = 0
	r.ConnInfo.Username = ""
	r.ConnInfo.Bio = ""
	r.ConnInfo.Save()
	r.DeviceToken = new(msg.AccountRegisterDevice)
	r.saveDeviceToken()

	err := r.Start()
	if err != nil {
		return err
	}

	r.StartNetwork()

	return err
}

// UpdateContactInfo update contact name
func (r *River) UpdateContactInfo(userID int64, firstName, lastName string) error {
	repo.Users.UpdateContactInfo(userID, firstName, lastName)
	return nil
}

// GetGroupInputUser get group participant user
func (r *River) GetGroupInputUser(requestID int64, groupID int64, userID int64, delegate RequestDelegate) {
	res := new(msg.MessageEnvelope)
	res.Constructor = msg.C_InputUser
	res.RequestID = uint64(requestID)

	user := new(msg.InputUser)
	user.UserID = userID

	accessHash, err := repo.Users.GetAccessHash(userID)
	if err != nil || accessHash == 0 {
		participant, err := repo.Groups.GetParticipants(groupID)
		if err == nil {
			for _, p := range participant {
				if p.UserID == userID {
					accessHash = p.AccessHash
					break
				}
			}
		} else {
			logs.Error("GetGroupInputUser() -> GetParticipants()", zap.Error(err))
		}
	} else {
		logs.Error("GetGroupInputUser() -> GetAccessHash()", zap.Error(err))
	}

	if accessHash == 0 {
		// get group full and get its access hash from its participants
		req := new(msg.GroupsGetFull)
		req.GroupID = groupID
		reqBytes, _ := req.Marshal()

		out := new(msg.MessageEnvelope)
		// Timeout Callback
		timeoutCB := func() {
			if delegate != nil {
				delegate.OnTimeout(domain.ErrRequestTimeout)
			}
		}

		// Success Callback
		successCB := func(response *msg.MessageEnvelope) {
			if response.Constructor != msg.C_GroupFull {
				msg.ResultError(out, &msg.Error{Code: "00", Items: "response type is not GroupFull"})
				if delegate != nil {
					outBytes, _ := out.Marshal()
					delegate.OnComplete(outBytes)
				}
				return
			}

			groupFull := new(msg.GroupFull)
			err := groupFull.Unmarshal(response.Message)
			if err != nil {
				msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
				if delegate != nil {
					outBytes, _ := out.Marshal()
					delegate.OnComplete(outBytes)
				}
				return
			}

			for _, p := range groupFull.Participants {
				if p.UserID == userID {
					user.AccessHash = p.AccessHash
					break
				}
			}

			res.Message, _ = user.Marshal()
			resBytes, _ := res.Marshal()
			if delegate != nil {
				delegate.OnComplete(resBytes)
			}
		}
		// SendWebsocket GroupsGetFull request to get user AccessHash
		r.queueCtrl.RealtimeCommand(uint64(requestID), msg.C_GroupsGetFull, reqBytes, timeoutCB, successCB, true, false)

	} else {
		user.AccessHash = accessHash
		res.Message, _ = user.Marshal()

		buff, _ := res.Marshal()
		if delegate != nil {
			delegate.OnComplete(buff)
		}
	}
}

func (r *River) GetScrollStatus(peerID int64, peerType int32) int64 {
	return repo.MessagesExtra.GetScrollID(peerID, peerType)
}

func (r *River) SetScrollStatus(peerID, msgID int64, peerType int32) {
	repo.MessagesExtra.SaveScrollID(peerID, peerType, msgID)

}

func (r *River) GetServerTimeUnix() int64 {
	return domain.Now().Unix()
}

// GenSrpHash generates a hash to be used in AuthCheckPassword and other related apis
func (r *River) GenSrpHash(password, prime, generator, salt1, salt2 []byte) []byte {
	g := big.NewInt(0).SetBytes(generator)
	p := big.NewInt(0).SetBytes(prime)

	x := big.NewInt(0).SetBytes(domain.PH2(password, salt1, salt2))
	v := big.NewInt(0).Exp(g, x, p)
	return v.Bytes()
}

// GenInputPassword  accepts AccountPassword marshaled as argument and return InputPassword marshaled
func (r *River) GenInputPassword(password []byte, accountPasswordBytes []byte) []byte {
	ap := &msg.AccountPassword{}
	err := ap.Unmarshal(accountPasswordBytes)

	algo := &msg.PasswordAlgorithmVer6A{}
	err = algo.Unmarshal(ap.AlgorithmData)
	if err != nil {
		logs.Warn("Error On GenInputPassword", zap.Error(err))
		return nil
	}

	p := big.NewInt(0).SetBytes(algo.P)
	g := big.NewInt(0).SetInt64(int64(algo.G))
	k := big.NewInt(0).SetBytes(domain.K(p, g))

	x := big.NewInt(0).SetBytes(domain.PH2(password, algo.Salt1, algo.Salt2))
	v := big.NewInt(0).Exp(g, x, p)
	a := big.NewInt(0).SetBytes(ap.RandomData)
	ga := big.NewInt(0).Exp(g, a, p)
	gb := big.NewInt(0).SetBytes(ap.SrpB)
	u := big.NewInt(0).SetBytes(domain.U(ga, gb))
	kv := big.NewInt(0).Mod(big.NewInt(0).Mul(k, v), p)
	t := big.NewInt(0).Mod(big.NewInt(0).Sub(gb, kv), p)
	if t.Sign() < 0 {
		t.Add(t, p)
	}
	sa := big.NewInt(0).Exp(t, big.NewInt(0).Add(a, big.NewInt(0).Mul(u, x)), p)
	m1 := domain.M(p, g, algo.Salt1, algo.Salt2, ga, gb, sa)

	inputPassword := &msg.InputPassword{
		SrpID: r.ConnInfo.UserID,
		A: domain.Pad(ga),
		M1: m1,
	}
	res, _ := inputPassword.Marshal()
	return res
}
