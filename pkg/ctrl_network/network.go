package network

import (
	"encoding/hex"
	ronak "git.ronaksoftware.com/ronak/toolbox"
	"net/http"
	"sort"
	"sync"
	"time"

	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	"github.com/dustin/go-humanize"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// Config network controller config
type Config struct {
	ServerEndpoint string
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
	pongChannel    chan bool

	// Authorization Keys
	authID        int64
	authKey       []byte
	messageSeq    int64
	salt          int64
	saltExpiry    int64

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

	// client and server time difference
	clientTimeDifference int64
}

// NewController
func NewController(config Config) *Controller {
	ctrl := new(Controller)
	if config.ServerEndpoint == "" {
		ctrl.websocketEndpoint = domain.WebsocketEndpoint
	} else {
		ctrl.websocketEndpoint = config.ServerEndpoint
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
	ctrl.pongChannel = make(chan bool)

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
			updates = append(updates, entries[idx].Value.(*msg.UpdateContainer))
		}

		// Call the update handler in blocking mode
		ctrl.OnUpdate(updates)
		logs.Debug("Updates Flushed",
			zap.Int("Count", itemsCount),
		)
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
		logs.Debug("Messages Flushed",
			zap.Int("Count", itemsCount),
		)
	}
}

func (ctrl *Controller) sendFlushFunc(entries []ronak.FlusherEntry) {
	// wait for network to connect
	for ctrl.wsQuality == domain.NetworkConnecting || ctrl.wsQuality == domain.NetworkDisconnected {
		time.Sleep(time.Second)
	}

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
		err := ctrl.send(messageEnvelope)
		if err != nil {
			logs.Error("NetworkController::Send Flush Error",
				zap.Error(err),
			)
			return
		}
		startIdx = endIdx
		endIdx += chunkSize
	}

	logs.Info("NetworkController::Send Flushed",
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
			logs.Debug("watchDog() Received connect signal")
			if ctrl.wsConn != nil {
				ctrl.receiver()
			}

			logs.Debug("watchDog() NetworkDisconnected")
			ctrl.updateNetworkStatus(domain.NetworkDisconnected)
			if ctrl.wsKeepConnection {

				logs.Debug("watchDog() Retry to Connect")
				go ctrl.Connect(false)
			}
		case <-ctrl.stopChannel:
			logs.Debug("watchDog() Stopped")
			return
		}
	}
}

// keepAlive
// This function by sending ping messages to server and measuring the server's response time
// calculates the quality of network
func (ctrl *Controller) keepAlive() {
	for {
		select {
		case <-time.After(ctrl.wsPingTime):
			if ctrl.wsConn == nil {
				continue
			}
			ctrl.wsWriteLock.Lock()
			_ = ctrl.wsConn.SetWriteDeadline(time.Now().Add(domain.WebsocketWriteTime))
			err := ctrl.wsConn.WriteMessage(websocket.PingMessage, nil)
			ctrl.wsWriteLock.Unlock()
			if err != nil {
				_ = ctrl.wsConn.SetReadDeadline(time.Now())
				logs.Warn("NetworkController::keepAlive() -> wsConn.WriteMessage()", zap.Error(err))
				continue
			}
			pingTime := time.Now()
			select {
			case <-ctrl.pongChannel:
				pingDelay := time.Now().Sub(pingTime)
				logs.Debug("keepAlive() Ping/Pong",
					zap.Duration("Duration", pingDelay),
					zap.String("wsQuality", ctrl.wsQuality.ToString()),
				)
				ctrl.pingIdx++
				ctrl.pingDelays[ctrl.pingIdx%3] = pingDelay
				avgDelay := (ctrl.pingDelays[0] + ctrl.pingDelays[1] + ctrl.pingDelays[2]) / 3
				switch {
				case avgDelay > 1500*time.Millisecond:
					ctrl.updateNetworkStatus(domain.NetworkWeak)
				case avgDelay > 500*time.Millisecond:
					ctrl.updateNetworkStatus(domain.NetworkSlow)
				default:
					ctrl.updateNetworkStatus(domain.NetworkFast)
				}
			case <-time.After(ctrl.wsPongTimeout):
				_ = ctrl.wsConn.SetReadDeadline(time.Now())
			}
		case <-ctrl.stopChannel:
			logs.Debug("keepAlive() Stopped")
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
			logs.Warn("NetworkController::receiver()-> ReadMessage()", zap.Error(err))
			return
		}
		logs.Debug("NetworkController:: Message Received",
			zap.Int("messageType", messageType),
			zap.Int("messageSize", len(message)),
		)

		switch messageType {
		case websocket.BinaryMessage:
			// If it is a BINARY message
			err := res.Unmarshal(message)
			if err != nil {
				logs.Error("NetworkController::receiver()", zap.Error(err), zap.String("Dump", string(message)))
				continue
			}

			if res.AuthID == 0 {
				logs.Debug("NetworkController::",
					zap.String("Warning", "res.AuthID is zero ProtoMessage is unencrypted"),
					zap.Int64("AuthID", res.AuthID),
				)
				receivedEnvelope := new(msg.MessageEnvelope)
				err = receivedEnvelope.Unmarshal(res.Payload)
				if err != nil {
					logs.Error("NetworkController::receiver() Failed to unmarshal", zap.Error(err))
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
					zap.String("resp.MessageKey", hex.Dump(res.MessageKey)))

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
			logs.Debug("NetworkController::",
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
	logs.Info("Network Status Updated",
		zap.String("Status", newStatus.ToString()),
	)
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
	messageCount := len(messages)
	updateCount := len(updates)
	logs.Debug("NetworkController:: Message Handler Received",
		zap.Int("Messages Count", messageCount),
		zap.Int("Updates Count", updateCount),
	)

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
	go ctrl.keepAlive()
	go ctrl.watchDog()
	return nil
}

// Stop sends stop signal to keepAlive and watchDog routines.
func (ctrl *Controller) Stop() {
	ctrl.stopChannel <- true // keepAlive
	select {
	case ctrl.stopChannel <- true: // receiver may or may not be listening
	default:
	}
}

// Connect dial websocket
func (ctrl *Controller) Connect(force bool) {
	defer func() {
		if recoverPanic("NetworkController:: Connect", ronak.M{
			"AuthID": ctrl.authID,
		}) {
			ctrl.Connect(force)
		}
	}()
	ctrl.updateNetworkStatus(domain.NetworkConnecting)
	keepGoing := true
	for keepGoing {
		if ctrl.wsConn != nil {
			_ = ctrl.wsConn.Close()
		}

		// Return if Disconnect() has been called
		if !force && !ctrl.wsKeepConnection {
			return
		}
		wsConn, _, err := ctrl.wsDialer.Dial(ctrl.websocketEndpoint, nil)
		if err != nil {
			logs.Warn("Connect()-> Dial()", zap.Error(err))
			time.Sleep(1 * time.Second)
			continue
		}
		keepGoing = false
		ctrl.wsConn = wsConn
		ctrl.wsKeepConnection = true
		ctrl.wsConn.SetPongHandler(func(appData string) error {
			ctrl.pongChannel <- true
			return nil
		})

		// it should be started here cuz we need receiver to get AuthRecall answer
		// Send Signal to start the 'receiver' and 'keepAlive' routines
		ctrl.connectChannel <- true

		// Call the OnConnect handler here b4 changing network status that trigger queue to start working
		// basically we send priority requests b4 queue starts to work
		ctrl.wsOnConnect()

		ctrl.updateNetworkStatus(domain.NetworkFast)
	}
}

// Disconnect close websocket
func (ctrl *Controller) Disconnect() {
	ctrl.wsKeepConnection = false
	if ctrl.wsConn != nil {
		_ = ctrl.wsConn.Close()

		logs.Info("NetworkController::Disconnect() Disconnected")
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

// Send direct sends immediately else it put it in debouncer
func (ctrl *Controller) Send(msgEnvelope *msg.MessageEnvelope, direct bool) error {
	logs.Debug("NetworkController:: Send",
		zap.String("Constructor", msg.ConstructorNames[msgEnvelope.Constructor]),
		zap.Bool("Direct", direct),
		zap.Int64("AuthID", ctrl.authID),
	)
	_, unauthorized := ctrl.unauthorizedRequests[msgEnvelope.Constructor]
	if direct || unauthorized {
		return ctrl.send(msgEnvelope)
	}
	ctrl.sendFlusher.Enter(nil, msgEnvelope)
	return nil
}

// send Writes the message on the wire. It will encrypts the message if authorization has been set.
func (ctrl *Controller) send(msgEnvelope *msg.MessageEnvelope) error {
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
			ServerSalt: ctrl.salt,
			Envelope:   msgEnvelope,
		}
		encryptedPayload.MessageID = uint64((time.Now().Unix()+ctrl.clientTimeDifference)<<32 | ctrl.messageSeq)
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
	logs.Debug("NetworkController:: Message sent to the wire",
		zap.String("Constructor", msg.ConstructorNames[msgEnvelope.Constructor]),
		zap.String("Size", humanize.Bytes(uint64(protoMessage.Size()))),
	)
	return nil
}

// Reconnect by wsKeepConnection = true the watchdog will connect itself again no need to call ctrl.Connect()
func (ctrl *Controller) Reconnect() {
	if ctrl.wsConn != nil {
		ctrl.wsKeepConnection = true
		ctrl.wsConn.Close()
		logs.Info("NetworkController::Reconnect() Reconnected")
	}
}

// SetClientTimeDifference set client and server time difference
func (ctrl *Controller) SetClientTimeDifference(delta int64) {
	ctrl.clientTimeDifference = delta
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

func (ctrl *Controller) ClientTimeDifference() int64 {
	return ctrl.clientTimeDifference
}

func (ctrl *Controller) SetServerSalt(salt int64) {
	ctrl.salt = salt
}

func (ctrl *Controller) GetServerSalt() int64 {
	return ctrl.salt
}

func (ctrl *Controller) SetSaltExpiry(timestamp int64) {
	ctrl.saltExpiry = timestamp
}

func (ctrl *Controller) GetSaltExpiry() int64 {
	return ctrl.saltExpiry
}

func recoverPanic(funcName string, extraInfo interface{}) bool {
	if r := recover(); r != nil {
		logs.Error("Panic Recovered",
			zap.String("Func", funcName),
			zap.Any("Info", extraInfo),
		)
		return true
	}
	return false
}
