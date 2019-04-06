package network

import (
	"encoding/hex"
	"net/http"
	"sort"
	"sync"
	"time"

	"git.ronaksoftware.com/ronak/riversdk/msg"
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
	//sync.Mutex
	wsWriteLock sync.Mutex

	// Internal Controller Channels
	connectChannel chan bool
	stopChannel    chan bool
	pongChannel    chan bool

	// Authorization Keys
	authID     int64
	authKey    []byte
	messageSeq int64

	// Websocket Settings
	wsDialer          *websocket.Dialer
	websocketEndpoint string
	wsKeepConnection  bool
	wsPingTime        time.Duration
	wsPongTimeout     time.Duration
	wsConn            *websocket.Conn
	//wsOnMessage             domain.MessageHandler
	//wsOnUpdate              domain.UpdateHandler
	wsOnError               domain.ErrorHandler
	wsOnConnect             domain.OnConnectCallback
	wsOnNetworkStatusChange domain.NetworkStatusUpdateCallback

	// Internals
	wsQuality  domain.NetworkStatus
	pingDelays [3]time.Duration
	pingIdx    int

	// queue
	messageQueue       *domain.QueueMessages
	updateQueue        *domain.QueueUpdates
	sendQueue          *domain.QueueMessages
	chMessageDebouncer chan bool
	chUpdateDebouncer  chan bool
	chSendDebouncer    chan bool
	// prevent write to socket concurrently
	// it used to prevent invoke read request while qriting on socket too
	wsSendDebouncerLock sync.Mutex

	OnMessage domain.ReceivedMessageHandler
	OnUpdate  domain.ReceivedUpdateHandler

	// requests that it should sent unencrypted
	unauthorizedRequests map[int64]bool

	// client and server time difference
	clientTimeDifference int64
}

// NewNetworkController
func NewNetworkController(config Config) *Controller {
	m := new(Controller)
	if config.ServerEndpoint == "" {
		m.websocketEndpoint = domain.WebsocketEndpoint
	} else {
		m.websocketEndpoint = config.ServerEndpoint
	}
	if config.PingTime <= 0 {
		m.wsPingTime = domain.WebsocketPingTime
	} else {
		m.wsPingTime = config.PingTime
	}
	if config.PongTimeout <= 0 {
		m.wsPongTimeout = domain.WebsocketPongTime
	} else {
		m.wsPongTimeout = config.PongTimeout
	}

	m.wsDialer = &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 10 * time.Second,
		WriteBufferSize:  64 * 1024, // 32kB
		ReadBufferSize:   64 * 1024, // 32kB
	}
	m.stopChannel = make(chan bool)
	m.connectChannel = make(chan bool)
	m.pongChannel = make(chan bool)

	m.updateQueue = domain.NewQueueUpdates()
	m.messageQueue = domain.NewQueueMessages()
	m.sendQueue = domain.NewQueueMessages()

	m.chMessageDebouncer = make(chan bool)
	m.chUpdateDebouncer = make(chan bool)
	m.chSendDebouncer = make(chan bool)

	m.unauthorizedRequests = map[int64]bool{
		msg.C_SystemGetServerTime: true,
	}

	return m
}

// Start
// Starts the controller background controller and watcher routines
func (ctrl *Controller) Start() error {
	if ctrl.OnUpdate == nil || ctrl.OnMessage == nil {
		return domain.ErrHandlerNotSet
	}
	go ctrl.keepAlive()
	go ctrl.watchDog()
	go ctrl.messageDebouncer()
	go ctrl.updateDebouncer()
	go ctrl.sendDebouncer()
	return nil
}

func (ctrl *Controller) messageDebouncer() {
	counter := 0
	interval := 50 * time.Millisecond
	timer := time.NewTimer(interval)
	for {
		select {
		case <-ctrl.chMessageDebouncer:
			counter++
			logs.Debug("messageDebouncer() Received",
				zap.Int("Counter", counter),
			)
			// on receive any update we wait another interval until we receive 3 update
			if counter < 3 {
				logs.Debug("messageDebouncer() Received Timer Reset",
					zap.Int("Counter", counter),
				)
				timer.Reset(interval)
			}
		case <-timer.C:
			if ctrl.OnMessage != nil && ctrl.messageQueue.Length() > 0 {
				logs.Debug("messageDebouncer() Flushed",
					zap.Int("ItemCount", ctrl.messageQueue.Length()),
				)
				ctrl.OnMessage(ctrl.messageQueue.PopAll())
				counter = 0
			}
			timer.Reset(interval)
		case <-ctrl.stopChannel:
			logs.Debug("messageDebouncer() Stopped")
			return
		}
	}

}

func (ctrl *Controller) updateDebouncer() {
	counter := 0
	interval := 50 * time.Millisecond
	timer := time.NewTimer(interval)
	for {
		select {
		case <-ctrl.chUpdateDebouncer:
			counter++
			logs.Debug("updateDebouncer() Received",
				zap.Int("Counter", counter),
			)
			// on receive any update we wait another interval until we receive 3 update
			if counter < 3 {
				logs.Debug("updateDebouncer() Received Timer Reset",
					zap.Int("Counter", counter),
				)
				timer.Reset(interval)
			}
		case <-timer.C:
			if ctrl.OnUpdate != nil && ctrl.updateQueue.Length() > 0 {
				logs.Debug("updateDebouncer() Flushed",
					zap.Int("ItemCount", ctrl.updateQueue.Length()),
				)
				ctrl.OnUpdate(ctrl.updateQueue.PopAll())
				counter = 0
			}
			timer.Reset(interval)
		case <-ctrl.stopChannel:
			logs.Debug("updateDebouncer() Stopped")
			return
		}
	}
}

func (ctrl *Controller) sendDebouncer() {
	counter := 0
	interval := 50 * time.Millisecond
	timer := time.NewTimer(interval)
	for {
		// wait for network to connect
		for ctrl.wsQuality == domain.NetworkConnecting || ctrl.wsQuality == domain.NetworkDisconnected {
			time.Sleep(interval)
		}

		select {
		case <-ctrl.chSendDebouncer:
			counter++
			logs.Debug("sendDebouncer() Received",
				zap.Int("Counter", counter),
			)
			// on receive any update we wait another interval until we receive 3 update
			if counter < 3 {
				logs.Debug("sendDebouncer() Received Timer Reset",
					zap.Int("Counter", counter),
				)
				timer.Reset(interval)
			}
		case <-timer.C:
			if ctrl.sendQueue.Length() > 0 {
				logs.Debug("sendDebouncer() Flushed",
					zap.Int("ItemCount", ctrl.sendQueue.Length()),
				)
				ctrl.sendFlush(ctrl.sendQueue.PopAll())
				counter = 0
			}
			timer.Reset(interval)
		case <-ctrl.stopChannel:
			logs.Debug("sendDebouncer() Stopped")
			return
		}
	}
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
				go ctrl.Connect()
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
			ctrl.wsConn.SetWriteDeadline(time.Now().Add(domain.WebsocketWriteTime))
			err := ctrl.wsConn.WriteMessage(websocket.PingMessage, nil)
			ctrl.wsWriteLock.Unlock()
			if err != nil {
				ctrl.wsConn.SetReadDeadline(time.Now())
				logs.Error("keepAlive() -> wsConn.WriteMessage()", zap.Error(err))
				continue
			}
			pingTime := time.Now()
			select {
			case <-ctrl.pongChannel:
				pingDelay := time.Now().Sub(pingTime)
				logs.Debug("keepAlive() Ping/Pong",
					zap.Duration("Duration", pingDelay),
					zap.String("wsQuality", domain.NetworkStatusName[ctrl.wsQuality]),
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
				ctrl.wsConn.SetReadDeadline(time.Now())
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
	res := msg.ProtoMessage{}
	for {
		// prevent webSocket.ReadMessage() request while writing message to websocket
		// this will wait till sendFlush() finish's its job
		ctrl.wsSendDebouncerLock.Lock()
		ctrl.wsSendDebouncerLock.Unlock()

		messageType, message, err := ctrl.wsConn.ReadMessage()
		if err != nil {
			logs.Error("receiver()-> ReadMessage()", zap.Error(err))
			return
		}
		logs.Debug("receiver() Message Received",
			zap.Int("messageType", messageType),
			zap.Int("messageSize", len(message)))

		switch messageType {
		case websocket.BinaryMessage:
			// If it is a BINARY message
			err := res.Unmarshal(message)
			if err != nil {
				logs.Error("receiver()", zap.Error(err), zap.String("Dump", string(message)))
				continue
			}
			if res.AuthID == 0 {

				logs.Debug("receiver()",
					zap.String("Warning", "res.AuthID is zero ProtoMessage is unencrypted"),
					zap.Int64("AuthID", res.AuthID),
				)
				receivedEnvelope := new(msg.MessageEnvelope)
				err = receivedEnvelope.Unmarshal(res.Payload)
				if err != nil {
					logs.Error("receiver() Failed to unmarshal", zap.Error(err))
					continue
				}
				ctrl.messageHandler(receivedEnvelope)
			} else {
				decryptedBytes, err := domain.Decrypt(ctrl.authKey, res.MessageKey, res.Payload)
				if err != nil {
					logs.Error("receiver()->Decrypt()",
						zap.String("Error", err.Error()),
						zap.Int64("ctrl.authID", ctrl.authID),
						zap.Int64("resp.AuthID", res.AuthID),
						zap.String("resp.MessageKey", hex.Dump(res.MessageKey)))

					continue
				}
				receivedEncryptedPayload := new(msg.ProtoEncryptedPayload)
				err = receivedEncryptedPayload.Unmarshal(decryptedBytes)
				if err != nil {
					logs.Error("receiver() Failed to unmarshal", zap.Error(err))
					continue
				}
				ctrl.messageHandler(receivedEncryptedPayload.Envelope)
			}
		}

	}
}

// updateNetworkStatus
// The average ping times will be calculated and this function will be called if
// quality of service changed.
func (ctrl *Controller) updateNetworkStatus(newStatus domain.NetworkStatus) {
	if ctrl.wsQuality == newStatus {
		logs.Debug("updateNetworkStatus() wsQuality not changed", zap.String("status", domain.NetworkStatusName[newStatus]))
		return
	}
	switch newStatus {
	case domain.NetworkDisconnected:
		logs.Info("updateNetworkStatus() Disconnected")
	case domain.NetworkWeak:
		logs.Info("updateNetworkStatus() Weak")
	case domain.NetworkSlow:
		logs.Info("updateNetworkStatus() Slow")
	case domain.NetworkFast:
		logs.Info("updateNetworkStatus() Fast")
	}
	ctrl.wsQuality = newStatus
	if ctrl.wsOnNetworkStatusChange != nil {
		ctrl.wsOnNetworkStatusChange(newStatus)
	}
}

// extractMessages recursively ectract update and messages
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

// send signal to debouncer
func (ctrl *Controller) messageReceived() {
	select {
	case ctrl.chMessageDebouncer <- true:
		logs.Debug("messageReceived() Signal")
	default:
		logs.Debug("messageReceived() Skipped")
	}
}

// send signal to debouncer
func (ctrl *Controller) updateReceived() {
	select {
	case ctrl.chUpdateDebouncer <- true:
		logs.Debug("updateReceived() Signal")
	default:
		logs.Debug("updateReceived() Skipped")
	}
}

// send signal to debouncer
func (ctrl *Controller) sendRequestReceived() {
	select {
	case ctrl.chSendDebouncer <- true:
		logs.Debug("sendRequestReceived() Signal")
	default:
		logs.Debug("sendRequestReceived() Skipped")
	}
}

// messageHandler
// MessageEnvelopes will be extracted and separates updates and messages.
func (ctrl *Controller) messageHandler(message *msg.MessageEnvelope) {
	logs.Debug("messageHandler() Begin")
	// extract all updates/ messages
	messages, updates := ctrl.extractMessages(message)
	messageCount := len(messages)
	updateCount := len(updates)
	logs.Debug("messageHandler()->extractMessages()",
		zap.Int("Messages Count", messageCount),
		zap.Int("Updates Count", updateCount),
	)

	// TODO : add to buffer queue and notify next processor to start processing received updates
	ctrl.messageQueue.PushMany(messages)
	ctrl.updateQueue.PushMany(updates)

	if messageCount > 0 {
		ctrl.messageReceived()
	}
	if updateCount > 0 {
		ctrl.updateReceived()
	}
	logs.Debug("messageHandler() End")
}

// Stop sends stop signal to keepAlive and watchDog routines.
func (ctrl *Controller) Stop() {
	ctrl.stopChannel <- true // keepAlive
	ctrl.stopChannel <- true // updateDebouncer
	ctrl.stopChannel <- true // messageDebouncer
	ctrl.stopChannel <- true // sendDebouncer
	select {
	case ctrl.stopChannel <- true: // receiver may or may not be listening
	default:
	}
}

// Connect dial websocket
func (ctrl *Controller) Connect() {
	logs.Info("Connect() Connecting")
	ctrl.updateNetworkStatus(domain.NetworkConnecting)
	keepGoing := true
	for keepGoing {
		if ctrl.wsConn != nil {
			ctrl.wsConn.Close()
		}
		if wsConn, _, err := ctrl.wsDialer.Dial(ctrl.websocketEndpoint, nil); err != nil {
			logs.Error("Connect()-> Dial()", zap.Error(err))
			time.Sleep(3 * time.Second)
		} else {
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

	logs.Info("Connect()  Connected")

}

// Disconnect close websocket
func (ctrl *Controller) Disconnect() {
	if ctrl.wsConn != nil {
		ctrl.wsKeepConnection = false
		ctrl.wsConn.Close()

		logs.Info("Disconnect() Disconnected")
	}
}

// SetAuthorization ...
// If authID and authKey are defined then sending messages will be encrypted before
// writing on the wire.
func (ctrl *Controller) SetAuthorization(authID int64, authKey []byte) {
	logs.Info("SetAuthorization()",
		zap.Int64("AuthID", authID),
		zap.Binary("authKey", authKey),
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

	// send without debouncer
	// return ctrl._send(msgEnvelope)

	_, unauthorized := ctrl.unauthorizedRequests[msgEnvelope.Constructor]
	if direct || unauthorized {
		// this will wait till sendFlush() done its job
		ctrl.wsSendDebouncerLock.Lock()
		ctrl.wsSendDebouncerLock.Unlock()

		return ctrl._send(msgEnvelope)
	} else {
		// this will probably solve queueController unordered message burst
		// add to send buffer
		ctrl.sendQueue.Push(msgEnvelope)
		// signal the send debouncer
		ctrl.sendRequestReceived()
	}
	return nil
}

// sendFlush will be called in sendDebouncer that running in another go routine so its ok to run in sync mode
func (ctrl *Controller) sendFlush(queueMsgs []*msg.MessageEnvelope) {

	logs.Debug("sendFlush()",
		zap.Int("queueMsgs Count", len(queueMsgs)),
	)

	ctrl.wsSendDebouncerLock.Lock()

	if len(queueMsgs) > 1 {
		// Implemented domain.SequentialUniqueID() to make sure requestID are sequential and unique
		sort.Slice(queueMsgs, func(i, j int) bool {
			return queueMsgs[i].RequestID < queueMsgs[j].RequestID
		})

		// chunk data by 50 and send
		limit := 50
		length := len(queueMsgs)
		rangeCount := length / limit
		for i := 0; i <= rangeCount; i++ {
			startIdx := i * limit
			// empty already
			if startIdx >= length {
				break
			}
			endIdx := startIdx + limit
			// check slice bounds
			if endIdx > length {
				endIdx = length
			}
			// fetch chunk
			chunkMsgs := queueMsgs[startIdx:endIdx]

			msgContainer := new(msg.MessageContainer)
			msgContainer.Envelopes = chunkMsgs
			msgContainer.Length = int32(len(chunkMsgs))

			messageEnvelop := new(msg.MessageEnvelope)
			messageEnvelop.Constructor = msg.C_MessageContainer
			messageEnvelop.Message, _ = msgContainer.Marshal()
			messageEnvelop.RequestID = 0 //uint64(domain.SequentialUniqueID())

			err := ctrl._send(messageEnvelop)
			if err != nil {
				logs.Error("sendFlush() -> ctrl._send() many", zap.Error(err))

				// add requests again to sendQueue and try again later
				logs.Debug("sendFlush() -> ctrl._send() many : pushed requests back to sendQueue")
				ctrl.sendQueue.PushMany(queueMsgs[startIdx:])
				break
			}
		}

	} else {
		err := ctrl._send(queueMsgs[0])
		if err != nil {
			logs.Error("sendFlush() -> ctrl._send() one", zap.Error(err))

			// add requests again to sendQueue and try again later
			logs.Warn("sendFlush() -> ctrl._send() one : pushed request back to sendQueue")
			ctrl.sendQueue.Push(queueMsgs[0])
		}
	}

	ctrl.wsSendDebouncerLock.Unlock()
}

// _send Writes the message on the wire. It will encrypts the message if authorization has been set.
func (ctrl *Controller) _send(msgEnvelope *msg.MessageEnvelope) error {
	protoMessage := new(msg.ProtoMessage)
	protoMessage.MessageKey = make([]byte, 32)
	_, unauthorized := ctrl.unauthorizedRequests[msgEnvelope.Constructor]
	if ctrl.authID == 0 || unauthorized {
		protoMessage.AuthID = 0
		logs.Debug("_send()",
			zap.String("Warning", "AuthID is zero ProtoMessage is unencrypted"),
			zap.Int64("ctrl.AuthID", ctrl.authID),
			zap.Int64("protoMessage.AuthID", protoMessage.AuthID),
		)
		protoMessage.Payload, _ = msgEnvelope.Marshal()
	} else {
		protoMessage.AuthID = ctrl.authID
		ctrl.messageSeq++
		encryptedPayload := msg.ProtoEncryptedPayload{
			ServerSalt: 234242, // TODO:: ServerSalt ?
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

	// // dump the message
	// err = ioutil.WriteFile(fmt.Sprintf("_dump/%v_%v.bin", time.Now().UnixNano(), msg.ConstructorNames[msgEnvelope.Constructor]), b, 777)
	// if err != nil {
	// 	log.Debug("failed to write dump",
	// 		zap.String("error", err.Error()),
	// 	)
	// }

	if err != nil {
		logs.Debug("_send()-> Failed to marshal", zap.Error(err))
	}
	if ctrl.wsConn == nil {
		logs.Debug("_send()->Marshal()",
			zap.String("Error", domain.ErrNoConnection.Error()),
		)
		return domain.ErrNoConnection
	}

	ctrl.wsWriteLock.Lock()
	ctrl.wsConn.SetWriteDeadline(time.Now().Add(domain.WebsocketWriteTime))
	err = ctrl.wsConn.WriteMessage(websocket.BinaryMessage, b)
	ctrl.wsWriteLock.Unlock()

	if err != nil {
		logs.Error("_send() -> wsConn.WriteMessage()", zap.Error(domain.ErrNoConnection))
		ctrl.updateNetworkStatus(domain.NetworkDisconnected)
		ctrl.wsConn.SetReadDeadline(time.Now())
		return err
	}
	logs.Debug("_send() Message sent to the wire",
		zap.String("Constructor", msg.ConstructorNames[msgEnvelope.Constructor]),
		zap.String("Size", humanize.Bytes(uint64(protoMessage.Size()))),
	)
	return nil
}

// Quality returns NetworkStatus
func (ctrl *Controller) Quality() domain.NetworkStatus {
	return ctrl.wsQuality
}

// PrintDebouncerStatus displays debuncer queues status
func (ctrl *Controller) PrintDebouncerStatus() {
	logs.Debug("Messages queue",
		zap.Int("Count", ctrl.messageQueue.Length()),
		zap.Int("Items Count", len(ctrl.messageQueue.GetRawItems())),
	)
}

// Reconnect by wsKeepConnection = true the watchdog will connect itself again no need to call ctrl.Connect()
func (ctrl *Controller) Reconnect() {
	if ctrl.wsConn != nil {
		ctrl.wsKeepConnection = true
		ctrl.wsConn.Close()
		// watchDog() will take care of this
		//go ctrl.Connect()

		logs.Info("Disconnect() Reconnected")
	}
}

// SetClientTimeDifference set client and server time difference
func (ctrl *Controller) SetClientTimeDifference(delta int64) {
	ctrl.clientTimeDifference = delta
}
