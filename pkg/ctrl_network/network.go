package networkCtrl

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"git.ronaksoftware.com/ronak/riversdk/pkg/salt"
	ronak "git.ronaksoftware.com/ronak/toolbox"
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
	// PingTime is the interval between each ping sent to server
	PingTime time.Duration
	// PongTimeout is the duration that will wait for the ping response (PONG) is received from
	// server. If we did not receive the pong message in PongTimeout then we assume connection
	// is not active, then we close it and try to re-connect.
	PongTimeout time.Duration
}

// Controller websocket network controller
type Controller struct {
	wsWriteLock sync.Mutex

	// Internal Controller Channels
	connectChannel chan bool
	stopChannel    chan bool

	// Authorization Keys
	authID     int64
	authKey    []byte
	messageSeq int64

	// Websocket Settings
	wsDialer                *websocket.Dialer
	websocketEndpoint       string
	wsKeepConnection        bool
	wsPingTime              time.Duration
	wsPongTimeout           time.Duration
	wsConn                  *websocket.Conn
	wsOnError               domain.ErrorHandler
	wsOnConnect             domain.OnConnectCallback
	wsOnNetworkStatusChange domain.NetworkStatusUpdateCallback

	// Http Settings
	httpEndpoint string

	// Internals
	wsQuality  domain.NetworkStatus
	pingDelays [3]time.Duration
	pingIdx    int

	// flusher
	updateFlusher  *ronak.Flusher
	messageFlusher *ronak.Flusher
	sendFlusher    *ronak.Flusher

	OnMessage domain.ReceivedMessageHandler
	OnUpdate  domain.ReceivedUpdateHandler

	// requests that it should sent unencrypted
	unauthorizedRequests map[int64]bool

	// internal parameters to detect network switch
	localIP       net.IP
	interfaceName string
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
	if config.PingTime <= 0 {
		ctrl.wsPingTime = domain.WebsocketPingTime
	} else {
		ctrl.wsPingTime = config.PingTime
	}
	if config.PongTimeout <= 0 {
		ctrl.wsPongTimeout = domain.WebsocketPongTime
	} else {
		ctrl.wsPongTimeout = config.PongTimeout
	}

	ctrl.wsKeepConnection = true
	ctrl.wsDialer = &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 10 * time.Second,
		WriteBufferSize:  64 * 1024, // 32kB
		ReadBufferSize:   64 * 1024, // 32kB
	}

	ctrl.stopChannel = make(chan bool)
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
		itemsCount := len(entries)
		updates := make([]*msg.UpdateContainer, 0, itemsCount)
		for idx := range entries {
			uc := entries[idx].Value.(*msg.UpdateContainer)
			sort.Slice(uc.Updates, func(i, j int) bool {
				return uc.Updates[i].UpdateID < uc.Updates[j].UpdateID
			})
			updates = append(updates, uc)
		}

		// Call the update handler in blocking mode
		ctrl.OnUpdate(updates)
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
			logs.Warn("NetworkController::SendWebsocket Flush Error", zap.Error(err))
			return
		}
		startIdx = endIdx
		endIdx += chunkSize
	}

	logs.Debug("NetworkController::SendWebsocket Flushed",
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
			if ctrl.wsConn != nil {
				ctrl.receiver()
			}
			ctrl.updateNetworkStatus(domain.NetworkDisconnected)
			if ctrl.wsKeepConnection {
				go ctrl.Connect(true)
			}
		case <-ctrl.stopChannel:
			logs.Info("Stop Called")
			return
		}
	}
}

// receiver
// This is the background routine listen for incoming websocket packets and _Decrypt
// the received message, if necessary, and  pass the extracted envelopes to messageHandler.
func (ctrl *Controller) receiver() {
	defer recoverPanic("NetworkController:: receiver", ronak.M{
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
				logs.Error("NetworkController::receiver()", zap.Error(err), zap.String("Dump", string(message)))
				continue
			}

			if res.AuthID == 0 {
				receivedEnvelope := new(msg.MessageEnvelope)
				err = receivedEnvelope.Unmarshal(res.Payload)
				if err != nil {
					logs.Error("NetworkController::receiver()-> Unmarshal()", zap.Error(err))
					continue
				}
				ctrl.messageHandler(receivedEnvelope)
				continue
			}

			// We received an encrypted message
			decryptedBytes, err := domain.Decrypt(ctrl.authKey, res.MessageKey, res.Payload)
			if err != nil {
				logs.Error("NetworkController::receiver()->Decrypt()",
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
				logs.Error("NetworkController::receiver() Failed to unmarshal", zap.Error(err))
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
	if ctrl.wsOnNetworkStatusChange != nil {
		ctrl.wsOnNetworkStatusChange(newStatus)
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
		e.Unmarshal(m.Message)
		// its general error
		if ctrl.wsOnError != nil && m.RequestID == 0 {
			ctrl.wsOnError(e)
		} else {
			// ui callback delegate will handle it
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
func (ctrl *Controller) Start() error {
	if ctrl.OnUpdate == nil || ctrl.OnMessage == nil {
		return domain.ErrHandlerNotSet
	}

	// Run the keepAlive and watchDog in background
	go ctrl.watchDog()
	return nil
}

// Stop sends stop signal to keepAlive and watchDog routines.
func (ctrl *Controller) Stop() {
	logs.Info("NetworkController is stopping")
	select {
	case ctrl.stopChannel <- true: // receiver may or may not be listening
	default:
	}
	logs.Info("NetworkController stopped")
}

// Connect dial websocket
func (ctrl *Controller) Connect(force bool) {
	logs.Info("NetworkController is connecting",
		zap.Bool("force", force),
	)
	defer func() {
		if recoverPanic("NetworkController:: Connect", ronak.M{
			"AuthID": ctrl.authID,
		}) {
			ctrl.Connect(force)
		}
	}()
	ctrl.wsKeepConnection = true
	ctrl.updateNetworkStatus(domain.NetworkConnecting)
	keepGoing := true
	for keepGoing {
		if ctrl.wsConn != nil {
			_ = ctrl.wsConn.Close()
		}
		if !force && !ctrl.wsKeepConnection {
			return
		}
		reqHdr := http.Header{}
		reqHdr.Set("X-Client-Type", fmt.Sprintf("SDK-%s", domain.SDKVersion))
		wsConn, _, err := ctrl.wsDialer.Dial(ctrl.websocketEndpoint, reqHdr)
		if err != nil {
			logs.Warn("Connect()-> Dial()", zap.Error(err))
			time.Sleep(1 * time.Second)
			continue
		}
		localIP := wsConn.UnderlyingConn().LocalAddr()
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
		ctrl.wsOnConnect()

		ctrl.updateNetworkStatus(domain.NetworkFast)
	}
}

// Disconnect close websocket
func (ctrl *Controller) Disconnect() {
	ctrl.wsKeepConnection = false
	if ctrl.wsConn != nil {
		_ = ctrl.wsConn.Close()
		logs.Info("NetworkController disconnected")
	}
}

// SetAuthorization ...
// If authID and authKey are defined then sending messages will be encrypted before
// writing on the wire.
func (ctrl *Controller) SetAuthorization(authID int64, authKey []byte) {
	logs.Info("NetworkController::SetAuthorization()",
		zap.Int64("AuthID", authID),
	)
	ctrl.authKey = make([]byte, len(authKey))
	ctrl.authID = authID
	copy(ctrl.authKey, authKey)
}

// SetErrorHandler set delegate handler
func (ctrl *Controller) SetErrorHandler(h domain.ErrorHandler) {
	ctrl.wsOnError = h
}

// SetMessageHandler set delegate handler
func (ctrl *Controller) SetMessageHandler(h domain.ReceivedMessageHandler) {
	ctrl.OnMessage = h
}

// SetUpdateHandler set delegate handler
func (ctrl *Controller) SetUpdateHandler(h domain.ReceivedUpdateHandler) {
	ctrl.OnUpdate = h
}

// SetOnConnectCallback set delegate handler
func (ctrl *Controller) SetOnConnectCallback(h domain.OnConnectCallback) {
	ctrl.wsOnConnect = h
}

// SetNetworkStatusChangedCallback set delegate handler
func (ctrl *Controller) SetNetworkStatusChangedCallback(h domain.NetworkStatusUpdateCallback) {
	ctrl.wsOnNetworkStatusChange = h
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
		ctrl.updateNetworkStatus(domain.NetworkDisconnected)
		_ = ctrl.wsConn.SetReadDeadline(time.Now())
		return err
	}
	return nil
}

// Send encrypt and send request to server and receive and decrypt its response
func (ctrl *Controller) SendHttp(msgEnvelope *msg.MessageEnvelope) (*msg.MessageEnvelope, error) {
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
		encryptedPayload.MessageID = uint64(time.Now().Unix()<<32 | ctrl.messageSeq)
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
	client := &http.Client{}
	client.Timeout = domain.HttpRequestTime

	// Send Data
	httpResp, err := client.Post(ctrl.httpEndpoint, "application/protobuf", reqBuff)
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
	if ctrl.wsConn != nil {
		ctrl.wsKeepConnection = true
		_ = ctrl.wsConn.Close()
	}
}

// WaitForNetwork
func (ctrl *Controller) WaitForNetwork() {
	// Wait While Network is Disconnected or Connecting
	for ctrl.wsQuality == domain.NetworkDisconnected || ctrl.wsQuality == domain.NetworkConnecting {
		logs.Debug("NetworkController:: Wait for Network",
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

func recoverPanic(funcName string, extraInfo interface{}) bool {
	if r := recover(); r != nil {
		logs.Error("Panic Recovered",
			zap.String("Func", funcName),
			zap.Any("Info", extraInfo),
			zap.Any("Recover", r),
		)
		return true
	}
	return false
}
