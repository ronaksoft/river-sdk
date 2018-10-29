package network

import (
	"encoding/hex"
	"net/http"
	"sort"
	"sync"
	"time"

	"git.ronaksoftware.com/ronak/riversdk/domain"
	"git.ronaksoftware.com/ronak/riversdk/log"
	"git.ronaksoftware.com/ronak/riversdk/msg"
	"github.com/dustin/go-humanize"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// TODO : implement this interface for NetworkController
// type NetworkController interface {
// }

// NetworkConfig
type NetworkConfig struct {
	ServerEndpoint string
	// PingTime is the interval between each ping sent to server
	PingTime time.Duration
	// PongTimeout is the duration that will wait for the ping response (PONG) is received from
	// server. If we did not receive the pong message in PongTimeout then we assume connection
	// is not active, then we close it and try to re-connect.
	PongTimeout time.Duration
}

// NetworkController
type NetworkController struct {
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
	messageQueue      *domain.QueueMessages
	updateQueue       *domain.QueueUpdates
	sendQueue         *domain.QueueMessages
	chMessageDebuncer chan bool
	chUpdateDebuncer  chan bool
	chSendDebuncer    chan bool

	OnMessage domain.OnMessageHandler
	OnUpdate  domain.OnUpdateHandler
}

// NewNetworkController
func NewNetworkController(config NetworkConfig) *NetworkController {
	m := new(NetworkController)
	if config.ServerEndpoint == "" {
		m.websocketEndpoint = domain.PRODUCTION_SERVER_WEBSOCKET_ENDPOINT
	} else {
		m.websocketEndpoint = config.ServerEndpoint
	}
	if config.PingTime <= 0 {
		m.wsPingTime = domain.DEFAULT_WS_PING_TIME
	} else {
		m.wsPingTime = config.PingTime
	}
	if config.PongTimeout <= 0 {
		m.wsPongTimeout = domain.DEFAULT_WS_PONG_TIMEOUT
	} else {
		m.wsPongTimeout = config.PongTimeout
	}

	m.wsDialer = &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 10 * time.Second,
	}
	m.stopChannel = make(chan bool)
	m.connectChannel = make(chan bool)
	m.pongChannel = make(chan bool)

	m.updateQueue = domain.NewQueueUpdates()
	m.messageQueue = domain.NewQueueMessages()
	m.sendQueue = domain.NewQueueMessages()

	m.chMessageDebuncer = make(chan bool)
	m.chUpdateDebuncer = make(chan bool)
	m.chSendDebuncer = make(chan bool)
	return m
}

// Start
// Starts the controller background controller and watcher routines
func (ctrl *NetworkController) Start() error {
	log.LOG.Debug("NetworkController::Start()")
	if ctrl.OnUpdate == nil || ctrl.OnMessage == nil {
		return domain.ErrHandlerNotSet
	}
	go ctrl.keepAlive()
	go ctrl.watchDog()
	go ctrl.messageDebuncer()
	go ctrl.updateDebuncer()
	// go ctrl.sendDebuncer()
	return nil
}

func (ctrl *NetworkController) messageDebuncer() {
	counter := 0
	timer := time.NewTimer(100 * time.Millisecond)
	for {
		select {
		case <-ctrl.chMessageDebuncer:
			counter++
			log.LOG.Debug("NetworkController::messageDebuncer() Received",
				zap.Int("Counter", counter),
			)
			// on receive any update we wait another 100 ms until we receive 3 update
			if counter < 3 {
				log.LOG.Debug("NetworkController::messageDebuncer() Received Timer Reset",
					zap.Int("Counter", counter),
				)
				timer.Reset(100 * time.Millisecond)
			}
		case <-timer.C:
			if ctrl.OnMessage != nil && counter > 0 {
				log.LOG.Debug("NetworkController::messageDebuncer() Flushed",
					zap.Int("ItemCount", ctrl.messageQueue.Length()),
				)
				ctrl.OnMessage(ctrl.messageQueue.PopAll())
				counter = 0
			}
		case <-ctrl.stopChannel:
			log.LOG.Debug("NetworkController::messageDebuncer() Stopped")
			return
		}
	}

}

func (ctrl *NetworkController) updateDebuncer() {
	counter := 0
	timer := time.NewTimer(100 * time.Millisecond)
	for {
		select {
		case <-ctrl.chUpdateDebuncer:
			counter++
			log.LOG.Debug("NetworkController::updateDebuncer() Received",
				zap.Int("Counter", counter),
			)
			// on receive any update we wait another 100 ms until we receive 3 update
			if counter < 3 {
				log.LOG.Debug("NetworkController::updateDebuncer() Received Timer Reset",
					zap.Int("Counter", counter),
				)
				timer.Reset(100 * time.Millisecond)
			}
		case <-timer.C:
			if ctrl.OnUpdate != nil && counter > 0 {
				log.LOG.Debug("NetworkController::updateDebuncer() Flushed",
					zap.Int("ItemCount", ctrl.updateQueue.Length()),
				)
				ctrl.OnUpdate(ctrl.updateQueue.PopAll())
				counter = 0
			}
		case <-ctrl.stopChannel:
			log.LOG.Debug("NetworkController::updateDebuncer() Stopped")
			return
		}
	}
}

func (ctrl *NetworkController) sendDebuncer() {
	counter := 0
	timer := time.NewTimer(100 * time.Millisecond)
	for {
		select {
		case <-ctrl.chSendDebuncer:
			counter++
			log.LOG.Debug("NetworkController::sendDebuncer() Received",
				zap.Int("Counter", counter),
			)
			// on receive any update we wait another 100 ms until we receive 3 update
			if counter < 3 {
				log.LOG.Debug("NetworkController::sendDebuncer() Received Timer Reset",
					zap.Int("Counter", counter),
				)
				timer.Reset(100 * time.Millisecond)
			}
		case <-timer.C:
			if counter > 0 {
				log.LOG.Debug("NetworkController::sendDebuncer() Flushed",
					zap.Int("ItemCount", ctrl.updateQueue.Length()),
				)
				ctrl.sendFlush(ctrl.sendQueue.PopAll())
				counter = 0
			}
		case <-ctrl.stopChannel:
			log.LOG.Debug("NetworkController::sendDebuncer() Stopped")
			return
		}
	}
}

// watchDog
// watchDog constantly listens to connectChannel and stopChannel to take the right
// actions in each case. If connect signal received the 'receiver' routine will be run
// to listen and accept web-socket packets. If stop signal received it means we are
// going to shutdown, hence returns from the function.
func (ctrl *NetworkController) watchDog() {
	for {
		select {
		case <-ctrl.connectChannel:
			log.LOG.Debug("NetworkController::watchDog() Received connect signal")
			if ctrl.wsConn != nil {
				ctrl.receiver()
			}

			log.LOG.Debug("NetworkController::watchDog() DISCONNECTED")
			ctrl.updateNetworkStatus(domain.DISCONNECTED)
			if ctrl.wsKeepConnection {

				log.LOG.Debug("NetworkController::watchDog() Retry to Connect")
				go ctrl.Connect()
			}
		case <-ctrl.stopChannel:
			log.LOG.Debug("NetworkController::watchDog() Stopped")
			return
		}

	}
}

// keepAlive
// This function by sending ping messages to server and measuring the server's response time
// calculates the quality of network
func (ctrl *NetworkController) keepAlive() {
	for {
		select {
		case <-time.After(ctrl.wsPingTime):
			if ctrl.wsConn == nil {
				continue
			}
			ctrl.wsWriteLock.Lock()
			ctrl.wsConn.SetWriteDeadline(time.Now().Add(domain.DEFAULT_WS_WRITE_TIMEOUT))
			err := ctrl.wsConn.WriteMessage(websocket.PingMessage, nil)
			ctrl.wsWriteLock.Unlock()
			if err != nil {
				ctrl.wsConn.SetReadDeadline(time.Now())
				continue
			}
			pingTime := time.Now()
			select {
			case <-ctrl.pongChannel:
				pingDelay := time.Now().Sub(pingTime)
				log.LOG.Debug("NetworkController::keepAlive() Ping/Pong",
					zap.Duration("Duration", pingDelay),
					zap.String("wsQuality", domain.NetworkStatusName[ctrl.wsQuality]),
				)
				ctrl.pingIdx++
				ctrl.pingDelays[ctrl.pingIdx%3] = pingDelay
				avgDelay := (ctrl.pingDelays[0] + ctrl.pingDelays[1] + ctrl.pingDelays[2]) / 3
				switch {
				case avgDelay > 1500*time.Millisecond:
					ctrl.updateNetworkStatus(domain.WEAK)
				case avgDelay > 500*time.Millisecond:
					ctrl.updateNetworkStatus(domain.SLOW)
				default:
					ctrl.updateNetworkStatus(domain.FAST)
				}
			case <-time.After(ctrl.wsPongTimeout):
				ctrl.wsConn.SetReadDeadline(time.Now())
			}
		case <-ctrl.stopChannel:
			log.LOG.Debug("NetworkController::keepAlive() Stopped")
			return
		}
	}
}

// receiver
// This is the background routine listen for incoming websocket packets and _Decrypt
// the received message, if necessary, and  pass the extracted envelopes to messageHandler.
func (ctrl *NetworkController) receiver() {
	res := msg.ProtoMessage{}
	for {
		messageType, message, err := ctrl.wsConn.ReadMessage()
		if err != nil {
			log.LOG.Debug("NetworkController::receiver()->ReadMessage()",
				zap.String("Error", err.Error()))
			return
		}
		log.LOG.Debug("NetworkController::receiver() Message Received",
			zap.Int("messageType : ", messageType),
			zap.Int("messageSize : ", len(message)))

		switch messageType {
		case websocket.BinaryMessage:
			// If it is a BINARY message
			res.Unmarshal(message)
			if res.AuthID == 0 {

				log.LOG.Debug("NetworkController::receiver()",
					zap.String("Warning", "res.AuthID is zero ProtoMessage is unencrypted"),
					zap.Int64("AuthID", res.AuthID),
				)
				receivedEnvelope := new(msg.MessageEnvelope)
				err = receivedEnvelope.Unmarshal(res.Payload)
				if err != nil {
					log.LOG.Debug(err.Error())
					continue
				}
				ctrl.messageHandler(receivedEnvelope)
			} else {
				decryptedBytes, err := domain.Decrypt(ctrl.authKey, res.MessageKey, res.Payload)
				if err != nil {
					log.LOG.Debug("NetworkController::receiver()->Decrypt()",
						zap.String("Error", err.Error()),
						zap.Int64(domain.LK_CLIENT_AUTH_ID, ctrl.authID),
						zap.Int64(domain.LK_SERVER_AUTH_ID, res.AuthID),
						zap.String(domain.LK_MSG_KEY, hex.Dump(res.MessageKey)))

					continue
				}
				receivedEncryptedPayload := new(msg.ProtoEncryptedPayload)
				err = receivedEncryptedPayload.Unmarshal(decryptedBytes)
				if err != nil {
					log.LOG.Debug("NetworkController::receiver()->Unmarshal()",
						zap.String("Error", err.Error()))
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
func (ctrl *NetworkController) updateNetworkStatus(newStatus domain.NetworkStatus) {
	if ctrl.wsQuality == newStatus {
		log.LOG.Info("NetworkController::updateNetworkStatus() wsQuality not changed")
		return
	}
	switch newStatus {
	case domain.DISCONNECTED:
		log.LOG.Info("NetworkController::updateNetworkStatus() Disconnected")
	case domain.WEAK:
		log.LOG.Info("NetworkController::updateNetworkStatus() Weak")
	case domain.SLOW:
		log.LOG.Info("NetworkController::updateNetworkStatus() Slow")
	case domain.FAST:
		log.LOG.Info("NetworkController::updateNetworkStatus() Fast")
	}
	ctrl.wsQuality = newStatus
	if ctrl.wsOnNetworkStatusChange != nil {
		ctrl.wsOnNetworkStatusChange(newStatus)
	}
}

func (ctrl *NetworkController) extractMessages(m *msg.MessageEnvelope) ([]*msg.MessageEnvelope, []*msg.UpdateContainer) {
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

// send signal to debuncer
func (ctrl *NetworkController) messageReceived() {
	select {
	case ctrl.chMessageDebuncer <- true:
		log.LOG.Debug("NetworkController::messageReceived() Signal")
	default:
		log.LOG.Debug("NetworkController::messageReceived() Skipped")
	}
}

// send signal to debuncer
func (ctrl *NetworkController) updateReceived() {
	select {
	case ctrl.chUpdateDebuncer <- true:
		log.LOG.Debug("NetworkController::updateReceived() Signal")
	default:
		log.LOG.Debug("NetworkController::updateReceived() Skipped")
	}
}

// send signal to debuncer
func (ctrl *NetworkController) sendRequestReceived() {
	select {
	case ctrl.chSendDebuncer <- true:
		log.LOG.Debug("NetworkController::sendRequestReceived() Signal")
	default:
		log.LOG.Debug("NetworkController::sendRequestReceived() Skipped")
	}
}

// messageHandler
// MessageEnvelopes will be extracted and separates updates and messages.
func (ctrl *NetworkController) messageHandler(message *msg.MessageEnvelope) {
	log.LOG.Debug("NetworkController::messageHandler() Begin")
	// extract all updates/ messages
	messages, updates := ctrl.extractMessages(message)
	messageCount := len(messages)
	updateCount := len(updates)
	log.LOG.Debug("NetworkController::messageHandler()->extractMessages()",
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
	log.LOG.Debug("NetworkController::messageHandler() End")
}

// Stop
// Sends stop signal to keepAlive and watchDog routines.
func (ctrl *NetworkController) Stop() {
	ctrl.stopChannel <- true // keepAlive
	ctrl.stopChannel <- true // receiver
	ctrl.stopChannel <- true // updateDebuncer
	ctrl.stopChannel <- true // messageDebuncer
	// ctrl.stopChannel <- true // sendDebuncer
}

// Connect
func (ctrl *NetworkController) Connect() {
	log.LOG.Info("NetworkController::Connect() Connecting")
	ctrl.updateNetworkStatus(domain.CONNECTING)
	keepGoing := true
	for keepGoing {
		if ctrl.wsConn != nil {
			ctrl.wsConn.Close()
		}
		if wsConn, _, err := ctrl.wsDialer.Dial(ctrl.websocketEndpoint, nil); err != nil {
			log.LOG.Info("NetworkController::Connect()->Dial()",
				zap.String("Error", err.Error()),
			)
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

			ctrl.updateNetworkStatus(domain.FAST)
		}
	}

	log.LOG.Info("NetworkController::Connect()  Connected")

}

// Disconnect
func (ctrl *NetworkController) Disconnect() {
	if ctrl.wsConn != nil {
		ctrl.wsKeepConnection = false
		ctrl.wsConn.Close()

		log.LOG.Info("NetworkController::Disconnect() Disconnected")
	}
}

// setAuthorization
// If authID and authKey are defined then sending messages will be encrypted before
// writing on the wire.
func (ctrl *NetworkController) SetAuthorization(authID int64, authKey []byte) {
	ctrl.authKey = make([]byte, len(authKey))
	ctrl.authID = authID
	copy(ctrl.authKey, authKey)
}

// SetErrorHandler
func (ctrl *NetworkController) SetErrorHandler(h domain.ErrorHandler) {
	ctrl.wsOnError = h
}

// SetMessageHandler
func (ctrl *NetworkController) SetMessageHandler(h domain.OnMessageHandler) {
	ctrl.OnMessage = h
}

// SetUpdateHandler
func (ctrl *NetworkController) SetUpdateHandler(h domain.OnUpdateHandler) {
	ctrl.OnUpdate = h
}

// SetOnConnectCallback
func (ctrl *NetworkController) SetOnConnectCallback(h domain.OnConnectCallback) {
	ctrl.wsOnConnect = h
}

// SetNetworkStatusChangedCallback
func (ctrl *NetworkController) SetNetworkStatusChangedCallback(h domain.NetworkStatusUpdateCallback) {
	ctrl.wsOnNetworkStatusChange = h
}

// Send
// TODO: implement queue for send requests but first check if server do accept MessageContainer or no
func (ctrl *NetworkController) Send(msgEnvelope *msg.MessageEnvelope) error {

	// // this will probably solve queueController unordered message burst
	// // add to send buffer
	// ctrl.sendQueue.Push(msgEnvelope)
	// // signal the send debuncer
	// ctrl.sendRequestReceived()
	// return nil

	// remove this line when sendDebuncer is Enabled
	return ctrl._send(msgEnvelope)
}

// sendFlush will be called in sendDebuncer that running in another go routine so its ok to run in sync mode
func (ctrl *NetworkController) sendFlush(queueMsgs []*msg.MessageEnvelope) {

	// TODO: need sort factor to make sure requests will be sent in desired order
	sort.Slice(queueMsgs, func(i, j int) bool {
		return queueMsgs[i].RequestID < queueMsgs[j].RequestID
	})

	msgContainer := new(msg.MessageContainer)
	msgContainer.Envelopes = queueMsgs
	msgContainer.Length = int32(len(queueMsgs))

	messageEnvelop := new(msg.MessageEnvelope)
	messageEnvelop.Constructor = msg.C_MessageContainer
	messageEnvelop.Message, _ = msgContainer.Marshal()
	messageEnvelop.RequestID = 0

	ctrl._send(messageEnvelop)
}

// sendWebsocket
// Writes the message on the wire. It will encrypts the message if authorization has been set.
func (ctrl *NetworkController) _send(msgEnvelope *msg.MessageEnvelope) error {
	protoMessage := new(msg.ProtoMessage)
	protoMessage.AuthID = ctrl.authID
	protoMessage.MessageKey = make([]byte, 32)
	if ctrl.authID == 0 {
		log.LOG.Debug("NetworkController::_send()",
			zap.String("Warning", "AuthID is zero ProtoMessage is unencrypted"),
			zap.Int64("AuthID", ctrl.authID),
		)
		protoMessage.Payload, _ = msgEnvelope.Marshal()
	} else {
		ctrl.messageSeq++
		encryptedPayload := msg.ProtoEncryptedPayload{
			ServerSalt: 234242, // TODO:: ServerSalt ?
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

	// // dump the message
	// err = ioutil.WriteFile(fmt.Sprintf("_dump/%v_%v.bin", time.Now().UnixNano(), msg.ConstructorNames[msgEnvelope.Constructor]), b, 777)
	// if err != nil {
	// 	log.LOG.Debug("failed to write dump",
	// 		zap.String("error", err.Error()),
	// 	)
	// }

	if err != nil {
		log.LOG.Debug("NetworkController::_send()->Marshal()",
			zap.String("Error", err.Error()),
		)
	}
	if ctrl.wsConn == nil {
		log.LOG.Debug("NetworkController::_send()->Marshal()",
			zap.String("Error", domain.ErrNoConnection.Error()),
		)
		return domain.ErrNoConnection
	}

	ctrl.wsWriteLock.Lock()
	ctrl.wsConn.SetWriteDeadline(time.Now().Add(domain.DEFAULT_WS_WRITE_TIMEOUT))
	err = ctrl.wsConn.WriteMessage(websocket.BinaryMessage, b)
	ctrl.wsWriteLock.Unlock()

	if err != nil {
		log.LOG.Debug("NetworkController::_send()->wsConn.WriteMessage()",
			zap.String("Error", domain.ErrNoConnection.Error()),
		)
		log.LOG.Debug(err.Error())
		ctrl.updateNetworkStatus(domain.DISCONNECTED)
		ctrl.wsConn.SetReadDeadline(time.Now())
		return err
	}
	log.LOG.Debug("NetworkController::_send() Message sent to the wire",
		zap.String(domain.LK_FUNC_NAME, "NetworkController::send"),
		zap.String("Constructor", msg.ConstructorNames[msgEnvelope.Constructor]),
		zap.String(domain.LK_MSG_SIZE, humanize.Bytes(uint64(protoMessage.Size()))),
	)
	return nil
}

// Quality returns NetworkStatus
func (ctrl *NetworkController) Quality() domain.NetworkStatus {
	return ctrl.wsQuality
}
