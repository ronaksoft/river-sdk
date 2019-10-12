package networkCtrl

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"git.ronaksoftware.com/ronak/riversdk/pkg/salt"
	ronak "git.ronaksoftware.com/ronak/toolbox"
	"github.com/felixge/tcpkeepalive"
	"io/ioutil"
	"net"
	"net/http"
	"sort"
	"sync"
	"time"

	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// Config network controller config
type Config struct {
	WebsocketEndpoint string
	HttpEndpoint      string
}

// Controller websocket network controller
type Controller struct {
	// Internal Controller Channels
	mtx            sync.Mutex
	connectChannel chan bool
	stopChannel    chan bool

	// Authorization Keys
	authID     int64
	authKey    []byte
	messageSeq int64

	// Websocket Settings
	wsWriteLock       sync.Mutex
	wsDialer          *websocket.Dialer
	websocketEndpoint string
	wsKeepConnection  bool
	wsConn            *websocket.Conn

	// Http Settings
	httpEndpoint string
	httpClient   http.Client

	// Internals
	wsQuality domain.NetworkStatus

	// flusher
	updateFlusher  *ronak.Flusher
	messageFlusher *ronak.Flusher
	sendFlusher    *ronak.Flusher

	// External Handlers
	OnMessage             domain.ReceivedMessageHandler
	OnUpdate              domain.ReceivedUpdateHandler
	OnGeneralError        domain.ErrorHandler
	OnWebsocketConnect    domain.OnConnectCallback
	OnNetworkStatusChange domain.NetworkStatusUpdateCallback

	// requests that it should sent unencrypted
	unauthorizedRequests map[int64]bool

	// internal parameters to detect network switch
	localIP net.IP
}

// New
func New(config Config) *Controller {
	ctrl := new(Controller)
	ctrl.httpEndpoint = config.HttpEndpoint
	if config.WebsocketEndpoint == "" {
		ctrl.websocketEndpoint = domain.WebsocketEndpoint
	} else {
		ctrl.websocketEndpoint = config.WebsocketEndpoint
	}
	ctrl.wsKeepConnection = true
	ctrl.wsDialer = &websocket.Dialer{
		NetDial: func(network, addr string) (conn net.Conn, e error) {
			return net.DialTimeout(network, addr, 5*time.Second)
		},
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 5 * time.Second,
		WriteBufferSize:  64 * 1024, // 32kB
		ReadBufferSize:   64 * 1024, // 32kB
	}

	ctrl.stopChannel = make(chan bool, 1)
	ctrl.connectChannel = make(chan bool)

	ctrl.updateFlusher = ronak.NewFlusher(1000, 1, 50*time.Millisecond, ctrl.updateFlushFunc)
	ctrl.messageFlusher = ronak.NewFlusher(1000, 1, 50*time.Millisecond, ctrl.messageFlushFunc)
	ctrl.sendFlusher = ronak.NewFlusher(1000, 1, 50*time.Millisecond, ctrl.sendFlushFunc)

	ctrl.unauthorizedRequests = map[int64]bool{
		msg.C_SystemGetServerTime: true,
		msg.C_InitConnect:         true,
		msg.C_InitCompleteAuth:    true,
		msg.C_SystemGetSalts:      true,
	}

	return ctrl
}

func (ctrl *Controller) updateFlushFunc(entries []ronak.FlusherEntry) {
	if ctrl.OnUpdate != nil {
		updateContainer := new(msg.UpdateContainer)
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].Value.(*msg.UpdateContainer).MinUpdateID < entries[j].Value.(*msg.UpdateContainer).MinUpdateID
		})
		users := make(map[int64]*msg.User)
		groups := make(map[int64]*msg.Group)
		for idx := range entries {
			uc := entries[idx].Value.(*msg.UpdateContainer)
			sort.Slice(uc.Updates, func(i, j int) bool {
				return uc.Updates[i].UpdateID < uc.Updates[j].UpdateID
			})

			if uc.MinUpdateID != 0 && (updateContainer.MinUpdateID == 0 || uc.MinUpdateID < updateContainer.MinUpdateID) {
				updateContainer.MinUpdateID = uc.MinUpdateID
			}
			if uc.MaxUpdateID > updateContainer.MaxUpdateID {
				updateContainer.MaxUpdateID = uc.MaxUpdateID
			}
			for _, u := range uc.Updates {
				updateContainer.Updates = append(updateContainer.Updates, u)
			}
			for _, u := range uc.Users {
				users[u.ID] = u
			}
			for _, g := range uc.Groups {
				groups[g.ID] = g
			}

			for _, g := range groups {
				updateContainer.Groups = append(updateContainer.Groups, g)
			}
			for _, u := range users {
				updateContainer.Users = append(updateContainer.Users, u)
			}
		}

		// Call the update handler in blocking mode
		ctrl.OnUpdate(updateContainer)
	}
}

func (ctrl *Controller) messageFlushFunc(entries []ronak.FlusherEntry) {
	if ctrl.OnMessage != nil {
		itemsCount := len(entries)
		messages := make([]*msg.MessageEnvelope, 0, itemsCount)
		for idx := range entries {
			messages = append(messages, entries[idx].Value.(*msg.MessageEnvelope))
		}

		// Call the message handler in blocking mode
		ctrl.OnMessage(messages)
	}
}

func (ctrl *Controller) sendFlushFunc(entries []ronak.FlusherEntry) {
	ctrl.WaitForNetwork()

	itemsCount := len(entries)
	messages := make([]*msg.MessageEnvelope, 0, itemsCount)
	for idx := range entries {
		messages = append(messages, entries[idx].Value.(*msg.MessageEnvelope))
	}

	// Make sure messages are sorted before sending to the wire
	sort.Slice(
		messages,
		func(i, j int) bool {
			return messages[i].RequestID < messages[j].RequestID
		},
	)

	chunkSize := 50
	startIdx := 0
	endIdx := chunkSize
	keepGoing := true
	for keepGoing {
		if endIdx > len(messages) {
			endIdx = len(messages)
			keepGoing = false
		}
		chunk := messages[startIdx:endIdx]

		msgContainer := new(msg.MessageContainer)
		msgContainer.Envelopes = chunk
		msgContainer.Length = int32(len(chunk))
		messageEnvelope := new(msg.MessageEnvelope)
		messageEnvelope.Constructor = msg.C_MessageContainer
		messageEnvelope.Message, _ = msgContainer.Marshal()
		messageEnvelope.RequestID = 0
		err := ctrl.sendWebsocket(messageEnvelope)
		if err != nil {
			logs.Warn("NetworkController got error on flushing outgoing messages",
				zap.Error(err),
				zap.Int("Count", len(chunk)),
			)
			return
		}
		startIdx = endIdx
		endIdx += chunkSize
	}

	logs.Debug("NetworkController flushed outgoing messages",
		zap.Int("Count", itemsCount),
	)
}

// watchDog
// watchDog constantly listens to connectChannel and stopChannel to take the right
// actions in each case. If connect signal received the 'receiver' routine will be run
// to listen and accept web-socket packets. If stop signal received it means we are
// going to shutdown, hence returns from the function.
func (ctrl *Controller) watchDog() {
	for {
		select {
		case <-ctrl.connectChannel:
			ctrl.receiver()
			ctrl.updateNetworkStatus(domain.NetworkDisconnected)
			if ctrl.wsKeepConnection {
				go ctrl.Connect()
			}
		case <-ctrl.stopChannel:
			logs.Info("NetworkController's watchdog stopped!")
			return
		}
	}
}

// receiver
// This is the background routine listen for incoming websocket packets and _Decrypt
// the received message, if necessary, and  pass the extracted envelopes to messageHandler.
func (ctrl *Controller) receiver() {
	defer ctrl.recoverPanic("NetworkController:: receiver", ronak.M{
		"AuthID": ctrl.authID,
	})
	res := msg.ProtoMessage{}
	for {
		messageType, message, err := ctrl.wsConn.ReadMessage()
		if err != nil {
			return
		}
		switch messageType {
		case websocket.BinaryMessage:
			// If it is a BINARY message
			err := res.Unmarshal(message)
			if err != nil {
				logs.Error("NetworkController couldn't unmarshal received BinaryMessage", zap.Error(err), zap.String("Dump", string(message)))
				continue
			}

			if res.AuthID == 0 {
				receivedEnvelope := new(msg.MessageEnvelope)
				err = receivedEnvelope.Unmarshal(res.Payload)
				if err != nil {
					logs.Error("NetworkController couldn't unmarshal plain-text MessageEnvelope", zap.Error(err))
					continue
				}
				ctrl.messageHandler(receivedEnvelope)
				continue
			}

			// We received an encrypted message
			decryptedBytes, err := domain.Decrypt(ctrl.authKey, res.MessageKey, res.Payload)
			if err != nil {
				logs.Error("NetworkController couldn't decrypt the received message",
					zap.String("Error", err.Error()),
					zap.Int64("ctrl.authID", ctrl.authID),
					zap.Int64("resp.AuthID", res.AuthID),
					zap.String("resp.MessageKey", hex.Dump(res.MessageKey)),
				)
				continue
			}
			receivedEncryptedPayload := new(msg.ProtoEncryptedPayload)
			err = receivedEncryptedPayload.Unmarshal(decryptedBytes)
			if err != nil {
				logs.Error("NetworkController couldn't unmarshal decrypted message", zap.Error(err))
				continue
			}
			// TODO:: check message id and server salt before handling the message
			ctrl.messageHandler(receivedEncryptedPayload.Envelope)
		default:
			logs.Warn("NetworkController received unhandled message type",
				zap.Int("MessageType", messageType),
			)
		}
	}
}

// updateNetworkStatus
// The average ping times will be calculated and this function will be called if
// quality of service changed.
func (ctrl *Controller) updateNetworkStatus(newStatus domain.NetworkStatus) {
	if ctrl.wsQuality == newStatus {
		return
	}
	ctrl.wsQuality = newStatus
	if ctrl.OnNetworkStatusChange != nil {
		ctrl.OnNetworkStatusChange(newStatus)
	}
}

// extractMessages recursively extract update and messages
func (ctrl *Controller) extractMessages(m *msg.MessageEnvelope) ([]*msg.MessageEnvelope, []*msg.UpdateContainer) {
	messages := make([]*msg.MessageEnvelope, 0)
	updates := make([]*msg.UpdateContainer, 0)
	switch m.Constructor {
	case msg.C_MessageContainer:
		x := new(msg.MessageContainer)
		err := x.Unmarshal(m.Message)
		if err == nil {
			for _, env := range x.Envelopes {
				msgs, upds := ctrl.extractMessages(env)
				messages = append(messages, msgs...)
				updates = append(updates, upds...)
			}
		}
	case msg.C_UpdateContainer:
		x := new(msg.UpdateContainer)
		err := x.Unmarshal(m.Message)
		if err == nil {
			updates = append(updates, x)
		}
	case msg.C_Error:
		e := new(msg.Error)
		_ = e.Unmarshal(m.Message)
		if ctrl.OnGeneralError != nil {
			ctrl.OnGeneralError(m.RequestID, e)
		}
		if m.RequestID != 0 {
			messages = append(messages, m)
		}
	default:
		messages = append(messages, m)
	}
	return messages, updates
}

// messageHandler
// MessageEnvelopes will be extracted and separates updates and messages.
func (ctrl *Controller) messageHandler(message *msg.MessageEnvelope) {
	// extract all updates/ messages
	messages, updates := ctrl.extractMessages(message)
	for idx := range messages {
		ctrl.messageFlusher.Enter(nil, messages[idx])
	}
	for idx := range updates {
		ctrl.updateFlusher.Enter(nil, updates[idx])
	}
}

// Start
// Starts the controller background controller and watcher routines
func (ctrl *Controller) Start() {
	// Run the keepAlive and watchDog in background
	go ctrl.watchDog()
	return
}

// Stop sends stop signal to keepAlive and watchDog routines.
func (ctrl *Controller) Stop() {
	logs.Info("NetworkController is stopping")
	ctrl.Disconnect()
	select {
	case ctrl.stopChannel <- true: // receiver may or may not be listening
	default:
	}
	logs.Info("NetworkController stopped")
}

// Connect dial websocket
func (ctrl *Controller) Connect() {
	_, _, _ = domain.SingleFlight.Do("NetworkConnect", func() (i interface{}, e error) {
		logs.Info("NetworkController is connecting")
		defer ctrl.recoverPanic("NetworkController:: Connect", ronak.M{
			"AuthID": ctrl.authID,
		})
		ctrl.mtx.Lock()
		ctrl.wsKeepConnection = true
		ctrl.updateNetworkStatus(domain.NetworkConnecting)
		keepGoing := true
		for keepGoing {
			// If there is a wsConn then close it before creating a new one
			if ctrl.wsConn != nil {
				_ = ctrl.wsConn.Close()
			}

			// Stop connecting if KeepConnection is FALSE
			if !ctrl.wsKeepConnection {
				ctrl.updateNetworkStatus(domain.NetworkDisconnected)
				ctrl.mtx.Unlock()
				return
			}
			reqHdr := http.Header{}
			reqHdr.Set("X-Client-Type", fmt.Sprintf("SDK-%s", domain.SDKVersion))
			wsConn, _, err := ctrl.wsDialer.Dial(ctrl.websocketEndpoint, reqHdr)
			if err != nil {
				logs.Warn("NetworkController could not dial", zap.Error(err), zap.String("Url", ctrl.websocketEndpoint))
				time.Sleep(1 * time.Second)
				continue
			}
			ctrl.mtx.Unlock()
			underlyingConn := wsConn.UnderlyingConn()
			_ = tcpkeepalive.SetKeepAlive(underlyingConn, 30*time.Second, 2, 5*time.Second)
			localIP := underlyingConn.LocalAddr()
			switch x := localIP.(type) {
			case *net.IPNet:
				ctrl.localIP = x.IP
			case *net.TCPAddr:
				ctrl.localIP = x.IP
			default:
				ctrl.localIP = nil
			}

			keepGoing = false
			ctrl.wsConn = wsConn

			// it should be started here cuz we need receiver to get AuthRecall answer
			// SendWebsocket Signal to start the 'receiver' and 'keepAlive' routines
			ctrl.connectChannel <- true
			logs.Info("NetworkController connected")

			// Call the OnConnect handler here b4 changing network status that trigger queue to start working
			// basically we sendWebsocket priority requests b4 queue starts to work
			ctrl.OnWebsocketConnect()

			if ctrl.GetQuality() == domain.NetworkDisconnected {
				continue
			}

			ctrl.updateNetworkStatus(domain.NetworkFast)
		}
		return nil, nil
	})
}

// Disconnect close websocket
func (ctrl *Controller) Disconnect() {
	_, _, _ = domain.SingleFlight.Do("NetworkDisconnect", func() (i interface{}, e error) {
		ctrl.mtx.Lock()
		ctrl.wsKeepConnection = false
		if ctrl.wsConn != nil {
			_ = ctrl.wsConn.SetReadDeadline(time.Now())
			_ = ctrl.wsConn.Close()
			logs.Info("NetworkController disconnected")
		}
		ctrl.mtx.Unlock()
		return nil, nil
	})
}

// SetAuthorization ...
// If authID and authKey are defined then sending messages will be encrypted before
// writing on the wire.
func (ctrl *Controller) SetAuthorization(authID int64, authKey []byte) {
	logs.Info("NetworkController set authorization info",
		zap.Int64("AuthID", authID),
	)
	ctrl.authKey = make([]byte, len(authKey))
	ctrl.authID = authID
	copy(ctrl.authKey, authKey)
}

// SendWebsocket direct sends immediately else it put it in flusher
func (ctrl *Controller) SendWebsocket(msgEnvelope *msg.MessageEnvelope, direct bool) error {
	_, unauthorized := ctrl.unauthorizedRequests[msgEnvelope.Constructor]
	if direct || unauthorized {
		return ctrl.sendWebsocket(msgEnvelope)
	}
	ctrl.sendFlusher.Enter(nil, msgEnvelope)
	return nil
}
func (ctrl *Controller) sendWebsocket(msgEnvelope *msg.MessageEnvelope) error {
	protoMessage := new(msg.ProtoMessage)
	protoMessage.MessageKey = make([]byte, 32)
	_, unauthorized := ctrl.unauthorizedRequests[msgEnvelope.Constructor]
	if ctrl.authID == 0 || unauthorized {
		protoMessage.AuthID = 0
		protoMessage.Payload, _ = msgEnvelope.Marshal()
	} else {
		protoMessage.AuthID = ctrl.authID
		ctrl.messageSeq++
		encryptedPayload := msg.ProtoEncryptedPayload{
			ServerSalt: salt.Get(),
			Envelope:   msgEnvelope,
		}
		encryptedPayload.MessageID = uint64(domain.Now().Unix()<<32 | ctrl.messageSeq)
		unencryptedBytes, _ := encryptedPayload.Marshal()
		encryptedPayloadBytes, _ := domain.Encrypt(ctrl.authKey, unencryptedBytes)
		messageKey := domain.GenerateMessageKey(ctrl.authKey, unencryptedBytes)
		copy(protoMessage.MessageKey, messageKey)
		protoMessage.Payload = encryptedPayloadBytes
	}

	b, err := protoMessage.Marshal()
	if err != nil {
		return err
	}
	if ctrl.wsConn == nil {
		return domain.ErrNoConnection
	}

	ctrl.wsWriteLock.Lock()
	_ = ctrl.wsConn.SetWriteDeadline(time.Now().Add(domain.WebsocketWriteTime))
	err = ctrl.wsConn.WriteMessage(websocket.BinaryMessage, b)
	ctrl.wsWriteLock.Unlock()

	if err != nil {
		_ = ctrl.wsConn.SetReadDeadline(time.Now())
		return err
	}

	return nil
}

// Send encrypt and send request to server and receive and decrypt its response
func (ctrl *Controller) SendHttp(ctx context.Context, msgEnvelope *msg.MessageEnvelope) (*msg.MessageEnvelope, error) {
	ctrl.WaitForNetwork()

	if ctx == nil {
		ctx = context.Background()
	}
	protoMessage := new(msg.ProtoMessage)
	protoMessage.AuthID = ctrl.authID
	protoMessage.MessageKey = make([]byte, 32)
	if ctrl.authID == 0 {
		protoMessage.Payload, _ = msgEnvelope.Marshal()
	} else {
		ctrl.messageSeq++
		encryptedPayload := msg.ProtoEncryptedPayload{
			ServerSalt: salt.Get(),
			Envelope:   msgEnvelope,
		}
		encryptedPayload.MessageID = uint64(domain.Now().Unix()<<32 | ctrl.messageSeq)
		unencryptedBytes, _ := encryptedPayload.Marshal()
		encryptedPayloadBytes, _ := domain.Encrypt(ctrl.authKey, unencryptedBytes)
		messageKey := domain.GenerateMessageKey(ctrl.authKey, unencryptedBytes)
		copy(protoMessage.MessageKey, messageKey)
		protoMessage.Payload = encryptedPayloadBytes
	}

	b, err := protoMessage.Marshal()
	reqBuff := bytes.NewBuffer(b)
	if err != nil {
		return nil, err
	}

	// Set timeout
	ctrl.httpClient.Timeout = domain.HttpRequestTime


	// Send Data
	httpReq, err := http.NewRequest(http.MethodPost, ctrl.httpEndpoint, reqBuff)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/protobuf")
	httpResp, err := ctrl.httpClient.Do(httpReq.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	// Read response
	resBuff, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		return nil, err
	}

	// Decrypt response
	res := new(msg.ProtoMessage)
	err = res.Unmarshal(resBuff)
	if err != nil {
		return nil, err
	}
	if res.AuthID == 0 {
		receivedEnvelope := new(msg.MessageEnvelope)
		err = receivedEnvelope.Unmarshal(res.Payload)
		return receivedEnvelope, err
	}
	decryptedBytes, err := domain.Decrypt(ctrl.authKey, res.MessageKey, res.Payload)

	receivedEncryptedPayload := new(msg.ProtoEncryptedPayload)
	err = receivedEncryptedPayload.Unmarshal(decryptedBytes)
	if err != nil {
		return nil, err
	}

	return receivedEncryptedPayload.Envelope, nil
}

// Reconnect by wsKeepConnection = true the watchdog will connect itself again no need to call ctrl.Connect()
func (ctrl *Controller) Reconnect() {
	_, _, _ = domain.SingleFlight.Do("NetworkReconnect", func() (i interface{}, e error)  {
		if ctrl.wsConn != nil {
			_ = ctrl.wsConn.SetReadDeadline(time.Now())

		}
		ctrl.Connect()
		return nil, nil
	})
}

// WaitForNetwork
func (ctrl *Controller) WaitForNetwork() {
	// Wait While Network is Disconnected or Connecting
	for ctrl.wsQuality == domain.NetworkDisconnected || ctrl.wsQuality == domain.NetworkConnecting {
		logs.Debug("NetworkController is waiting for Network",
			zap.String("Quality", ctrl.wsQuality.ToString()),
		)
		time.Sleep(time.Second)
	}
}

// Connected
func (ctrl *Controller) Connected() bool {
	if ctrl.wsQuality == domain.NetworkDisconnected || ctrl.wsQuality == domain.NetworkConnecting {
		return false
	}
	return true
}

// GetQuality
func (ctrl *Controller) GetQuality() domain.NetworkStatus {
	return ctrl.wsQuality
}

func (ctrl *Controller) recoverPanic(funcName string, extraInfo interface{}) {
	if r := recover(); r != nil {
		logs.Error("Panic Recovered",
			zap.String("Func", funcName),
			zap.Any("Info", extraInfo),
			zap.Any("Recover", r),
		)
		go ctrl.Reconnect()
	}
}
