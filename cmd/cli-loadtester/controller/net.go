package controller

import (
	"fmt"
	"hash/crc32"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"

	"git.ronaksoftware.com/ronak/riversdk/loadtester/shared"
	"git.ronaksoftware.com/ronak/riversdk/msg"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	"github.com/gorilla/websocket"
)

// CtrlNetwork network layer
type CtrlNetwork struct {
	isConnected  bool
	Disconnected int64

	connWriteLock       sync.Mutex
	conn                *websocket.Conn
	wsDialer            *websocket.Dialer
	stop                chan bool
	keepConnectionAlive bool

	messageSeq int64

	actor shared.Acter

	onError   domain.ErrorHandler
	onMessage domain.ReceivedMessageHandler
	onUpdate  domain.ReceivedUpdateHandler
}

// NewCtrlNetwork create new instance
func NewCtrlNetwork(act shared.Acter,
	onMessage domain.ReceivedMessageHandler,
	onUpdate domain.ReceivedUpdateHandler,
	onError domain.ErrorHandler,
) shared.Neter {
	n := &CtrlNetwork{
		stop:      make(chan bool),
		wsDialer:  websocket.DefaultDialer,
		actor:     act,
		onError:   onError,
		onMessage: onMessage,
		onUpdate:  onUpdate,
	}
	n.wsDialer.ReadBufferSize = 32 * 1024  // 32KB
	n.wsDialer.WriteBufferSize = 32 * 1024 // 32KB

	return n
}

// Start start websocket
func (ctrl *CtrlNetwork) Start() error {
	err := ctrl.connect()
	if err == nil {
		ctrl.keepConnectionAlive = true
		go ctrl.watchDog()
		ctrl.onConnect()
	}
	return err
}

// Stop disconnect websocket and signal stop chan
func (ctrl *CtrlNetwork) Stop() {
	ctrl.keepConnectionAlive = false
	ctrl.disconnect()
	// signal watchDog
	ctrl.stop <- true
	ctrl.isConnected = false
	ctrl.conn = nil
}

// Send the data payload is binary
func (ctrl *CtrlNetwork) Send(msgEnvelope *msg.MessageEnvelope) error {
	shared.Metrics.CounterVec(shared.CntRequest).WithLabelValues(msg.ConstructorNames[msgEnvelope.Constructor]).Add(1)

	authID, authKey := ctrl.actor.GetAuthInfo()

	protoMessage := new(msg.ProtoMessage)
	protoMessage.AuthID = authID
	protoMessage.MessageKey = make([]byte, 32)
	if authID == 0 {
		payload, err := msgEnvelope.Marshal()
		if err != nil {
			return err
		}
		protoMessage.Payload = payload
	} else {
		ctrl.messageSeq++
		encryptedPayload := msg.ProtoEncryptedPayload{
			ServerSalt: 234242, // TODO:: ServerSalt ?
			Envelope:   msgEnvelope,
		}
		encryptedPayload.MessageID = uint64(time.Now().Unix()<<32 | ctrl.messageSeq)
		unencryptedBytes, err := encryptedPayload.Marshal()
		if err != nil {
			return err
		}
		encryptedPayloadBytes, err := domain.Encrypt(authKey, unencryptedBytes)
		if err != nil {
			return err
		}
		messageKey := domain.GenerateMessageKey(authKey, unencryptedBytes)
		copy(protoMessage.MessageKey, messageKey)
		protoMessage.Payload = encryptedPayloadBytes
	}

	// log
	if msgEnvelope.Constructor != msg.C_AuthRecall {
		msgKey := crc32.ChecksumIEEE(protoMessage.MessageKey)
		logs.Debug("Crc32 of MessageKey & RequestID",
			zap.Uint32("Crc32", msgKey),
			zap.Uint64("ReqID", msgEnvelope.RequestID),
			zap.String("Constructor", msg.ConstructorNames[msgEnvelope.Constructor]),
		)
	}

	b, err := protoMessage.Marshal()
	if err != nil {
		return err
	}
	// // dump request
	// if msgEnvelope.Constructor == msg.C_ContactsImport {
	// 	err := ioutil.WriteFile("ImportContact_Dump.raw", b, os.ModePerm)
	// 	if err != nil {
	// 		log.Error("Packet Dump Failed", zap.Error(err))
	// 	}
	// }

	// metric
	shared.Metrics.Counter(shared.CntSend).Add(float64(len(b)))

	if ctrl.conn == nil {
		return fmt.Errorf("network connection is null")
	}
	ctrl.connWriteLock.Lock()
	ctrl.conn.SetWriteDeadline(time.Now().Add(shared.DefaultSendTimeout))
	err = ctrl.conn.WriteMessage(websocket.BinaryMessage, b)
	ctrl.connWriteLock.Unlock()

	// log sent packet
	logSentPacket(protoMessage)

	return err
}

// IsConnected connection status
func (ctrl *CtrlNetwork) IsConnected() bool {
	return ctrl.isConnected
}

// DisconnectCount network failed count
func (ctrl *CtrlNetwork) DisconnectCount() int64 {
	return ctrl.Disconnected
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
			ctrl.isConnected = false
			if ctrl.keepConnectionAlive {
				err := ctrl.connect()
				if err == nil {
					ctrl.onConnect()
				}
			}
		}
	}
}

// onConnect send AuthRecall request to server
func (ctrl *CtrlNetwork) onConnect() {

	// if authkey not created skip
	if ctrl.actor.GetAuthID() == 0 {
		return
	}
	req := msg.AuthRecall{}
	reqID := uint64(shared.GetSeqID())

	envelop := new(msg.MessageEnvelope)
	envelop.Constructor = msg.C_AuthRecall
	envelop.Message, _ = req.Marshal()
	envelop.RequestID = reqID

	if ctrl.conn == nil {
		return
	}
	err := ctrl.Send(envelop)
	if err != nil {
		logs.Error("onConnect() AuthRecall", zap.Error(err))
	}

	// // 3 is max retry to send authRecal
	// for i := 0; i < 3; i++ {
	// 	envelop := new(msg.MessageEnvelope)
	// 	envelop.Constructor = msg.C_AuthRecall
	// 	envelop.Message, _ = req.Marshal()
	// 	envelop.RequestID = reqID
	// 	if ctrl.conn == nil {
	// 		return
	// 	}
	// 	err := ctrl.Send(envelop)
	// 	if err == nil {
	// 		break
	// 	} else {
	// 		log.Error("onConnect() AuthRecall", zap.Error(err))
	// 		// time.Sleep(time.Microsecond * 50)
	// 	}
	// }

	ctrl.isConnected = true
}

// receiver read messages from websocket and pass them to proper handler
func (ctrl *CtrlNetwork) receiver() {
	if !ctrl.keepConnectionAlive {
		return
	}
	for {
		messageType, message, err := ctrl.conn.ReadMessage()
		if err != nil {
			// on stop request we set keepConnectionAlive to false
			if ctrl.keepConnectionAlive {
				//metric
				shared.Metrics.Counter(shared.CntDisconnect).Add(1)

				atomic.AddInt64(&ctrl.Disconnected, 1)
			}
			return
		}

		// metric
		shared.Metrics.Counter(shared.CntReceive).Add(float64(len(message)))

		if messageType != websocket.BinaryMessage {
			continue
		}

		res := new(msg.ProtoMessage)
		err = res.Unmarshal(message)
		if err != nil {
			logs.Error("CtrlNetwork::receiver() failed to unmarshal to ProtoMessage", zap.Error(err))
			continue
		}

		if res.AuthID != ctrl.actor.GetAuthID() {
			logs.Error("Received unmatched AuthID ", zap.Int64("Server.AutID", res.AuthID), zap.Int64("Actor.AuthID", ctrl.actor.GetAuthID()))
		}

		// log received packet
		logReceivedPacket(res)

		if res.AuthID == 0 {
			receivedEnvelope := new(msg.MessageEnvelope)
			err = receivedEnvelope.Unmarshal(res.Payload)
			if err != nil {
				logs.Error("CtrlNetwork::receiver() failed to unmarshal to MessageEnvelope", zap.Error(err))
				continue
			}

			ctrl.messageHandler(receivedEnvelope)
		} else {
			decryptedBytes, err := domain.Decrypt(ctrl.actor.GetAuthKey(), res.MessageKey, res.Payload)
			if err != nil {
				logs.Error("CtrlNetwork::receiver() failed to domain.Decrypt()", zap.Error(err))
				continue
			}
			receivedEncryptedPayload := new(msg.ProtoEncryptedPayload)
			err = receivedEncryptedPayload.Unmarshal(decryptedBytes)
			if err != nil {
				logs.Error("CtrlNetwork::receiver() failed to unmarshal to ProtoEncryptedPayload", zap.Error(err))
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
	if ctrl.onUpdate != nil {
		ctrl.onUpdate(updates)
	}
}
