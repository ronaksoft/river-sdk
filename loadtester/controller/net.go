package controller

import (
	"sync"
	"time"

	"git.ronaksoftware.com/ronak/riversdk/domain"
	"git.ronaksoftware.com/ronak/riversdk/loadtester/shared"
	"git.ronaksoftware.com/ronak/riversdk/msg"
	"github.com/gorilla/websocket"
)

// CtrlNetwork network layer
type CtrlNetwork struct {
	IsConnected  bool
	Disconnected int

	connWriteLock       sync.Mutex
	conn                *websocket.Conn
	wsDialer            *websocket.Dialer
	stop                chan bool
	keepConnectionAlive bool

	messageSeq int64

	// Actor Info :
	authID    int64
	authKey   []byte //[256]byte
	onError   domain.ErrorHandler
	onMessage domain.OnMessageHandler
	onUpdate  domain.OnUpdateHandler
}

// NewCtrlNetwork create new instance
func NewCtrlNetwork(authID int64, authKey []byte, onMessage domain.OnMessageHandler, onUpdate domain.OnUpdateHandler, onError domain.ErrorHandler) *CtrlNetwork {
	n := &CtrlNetwork{
		stop:      make(chan bool),
		wsDialer:  websocket.DefaultDialer,
		authID:    authID,
		authKey:   authKey,
		onError:   onError,
		onMessage: onMessage,
		onUpdate:  onUpdate,
	}
	// n.wsDialer.ReadBufferSize = 32 * 1024  // 32KB
	// n.wsDialer.WriteBufferSize = 32 * 1024 // 32KB

	return n
}

// Start start websocket
func (ctrl *CtrlNetwork) Start() error {
	err := ctrl.connect()
	if err == nil {
		ctrl.keepConnectionAlive = true
		go ctrl.watchDog()
		go ctrl.onConnect()
	}
	return err
}

// Stop disconnect websocket and signal stop chan
func (ctrl *CtrlNetwork) Stop() {
	ctrl.keepConnectionAlive = false
	ctrl.disconnect()
	// signal watchDog
	ctrl.stop <- true
	ctrl.IsConnected = false
}

// Send the data payload is binary
func (ctrl *CtrlNetwork) Send(msgEnvelope *msg.MessageEnvelope) error {
	protoMessage := new(msg.ProtoMessage)
	protoMessage.AuthID = ctrl.authID
	protoMessage.MessageKey = make([]byte, 32)
	if ctrl.authID == 0 {
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

	ctrl.connWriteLock.Lock()
	ctrl.conn.SetWriteDeadline(time.Now().Add(shared.DefaultSendTimeout))
	err = ctrl.conn.WriteMessage(websocket.BinaryMessage, b)
	ctrl.connWriteLock.Unlock()

	return err
}

// connect to websocket and start receiving data from websocket
func (ctrl *CtrlNetwork) connect() error {
	conn, _, err := ctrl.wsDialer.Dial(shared.DefaultServerURL, nil)
	if err == nil {
		ctrl.conn = conn
	}
	return err
}

// disconnect close websocket
func (ctrl *CtrlNetwork) disconnect() error {
	return ctrl.conn.Close()
}

// watchDog keeps websocket connected
func (ctrl *CtrlNetwork) watchDog() {
	for {
		select {
		case <-ctrl.stop:
			return
		default:
			if ctrl.conn != nil {
				ctrl.receiver()
			}
			ctrl.IsConnected = false
			if ctrl.keepConnectionAlive {
				err := ctrl.connect()
				if err == nil {
					go ctrl.onConnect()
				}
			}
		}
	}
}

// onConnect send AuthRecall request to server
func (ctrl *CtrlNetwork) onConnect() {
	req := msg.AuthRecall{}
	reqID := uint64(domain.SequentialUniqueID())
	for {
		envelop := new(msg.MessageEnvelope)
		envelop.Constructor = msg.C_AuthRecall
		envelop.Message, _ = req.Marshal()
		envelop.RequestID = reqID
		err := ctrl.Send(envelop)
		if err == nil {
			break
		} else {
			time.Sleep(1 * time.Second)
		}
	}
	ctrl.IsConnected = true
}

// receiver read messages from websocket and pass them to proper handler
func (ctrl *CtrlNetwork) receiver() {
	for {
		messageType, message, err := ctrl.conn.ReadMessage()
		if err != nil {
			ctrl.Disconnected++
			return
		}
		if messageType != websocket.BinaryMessage {
			continue
		}

		res := msg.ProtoMessage{}
		res.Unmarshal(message)
		if res.AuthID == 0 {
			receivedEnvelope := new(msg.MessageEnvelope)
			err = receivedEnvelope.Unmarshal(res.Payload)
			if err != nil {
				continue
			}
			ctrl.messageHandler(receivedEnvelope)
		} else {
			decryptedBytes, err := domain.Decrypt(ctrl.authKey, res.MessageKey, res.Payload)
			if err != nil {
				continue
			}
			receivedEncryptedPayload := new(msg.ProtoEncryptedPayload)
			err = receivedEncryptedPayload.Unmarshal(decryptedBytes)
			if err != nil {
				continue
			}
			ctrl.messageHandler(receivedEncryptedPayload.Envelope)
		}
	}
}

func (ctrl *CtrlNetwork) extractMessages(m *msg.MessageEnvelope) ([]*msg.MessageEnvelope, []*msg.UpdateContainer) {
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
		if ctrl.onError != nil && m.RequestID == 0 {
			ctrl.onError(e)
		} else {
			// ui callback delegate will handle it
			messages = append(messages, m)
		}

	default:
		messages = append(messages, m)
	}
	return messages, updates
}

func (ctrl *CtrlNetwork) messageHandler(message *msg.MessageEnvelope) {
	// extract all updates/ messages
	messages, updates := ctrl.extractMessages(message)
	if ctrl.onMessage != nil {
		ctrl.onMessage(messages)
	}
	if ctrl.onMessage != nil {
		ctrl.onUpdate(updates)
	}
}
