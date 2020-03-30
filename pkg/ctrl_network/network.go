package networkCtrl

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	mon "git.ronaksoftware.com/ronak/riversdk/pkg/monitoring"
	"git.ronaksoftware.com/ronak/riversdk/pkg/salt"
	ronak "git.ronaksoftware.com/ronak/toolbox"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"io/ioutil"
	"net"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	msg "git.ronaksoftware.com/river/msg/chat"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	"go.uber.org/zap"
)

// Config network controller config
type Config struct {
	WebsocketEndpoint string
	HttpEndpoint      string
	CountryCode       string
}

// Controller websocket network controller
type Controller struct {
	// Internal Controller Channels
	connectChannel  chan bool
	stopChannel     chan bool
	endpointUpdated bool

	// Authorization Keys
	authID     int64
	authKey    []byte
	messageSeq int64

	// Websocket Settings
	wsWriteLock      sync.Mutex
	wsDialer         *ws.Dialer
	wsEndpoint       string
	wsKeepConnection bool
	wsConn           net.Conn

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
	OnNetworkStatusChange domain.NetworkStatusChangeCallback

	// requests that it should sent unencrypted
	unauthorizedRequests map[int64]bool

	// internal parameters to detect network switch
	countryCode string
}

// New
func New(config Config) *Controller {
	ctrl := new(Controller)
	ctrl.httpEndpoint = config.HttpEndpoint
	ctrl.countryCode = strings.ToUpper(config.CountryCode)
	if config.WebsocketEndpoint == "" {
		ctrl.wsEndpoint = domain.DefaultWebsocketEndpoint
	} else {
		ctrl.wsEndpoint = config.WebsocketEndpoint
	}
	ctrl.wsKeepConnection = true

	ctrl.createDialer(domain.WebsocketDialTimeout)

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
func (ctrl *Controller) createDialer(timeout time.Duration) {
	ctrl.wsDialer = &ws.Dialer{
		ReadBufferSize:  32 * 1024, // 32kB
		WriteBufferSize: 32 * 1024, // 32kB
		Timeout:         timeout,
		NetDial: func(ctx context.Context, network, addr string) (conn net.Conn, err error) {
			host, port, err := net.SplitHostPort(addr)
			if err != nil {
				return nil, err
			}
			ips, err := net.LookupIP(host)
			if err != nil {
				return nil, err
			}
			logs.Info("DNS LookIP",
				zap.String("Addr", addr),
				zap.Any("IPs", ips),
			)
			d := net.Dialer{Timeout: timeout}
			for _, ip := range ips {
				if ip.To4() != nil {
					conn, err = d.DialContext(ctx, "tcp4", net.JoinHostPort(ip.String(), port))
					if err != nil {
						continue
					}
					return
				}
			}
			return nil, domain.ErrNoConnection
		},
		OnStatusError: nil,
		OnHeader:      nil,
		TLSClient:     nil,
		TLSConfig:     nil,
		WrapConn:      nil,
	}
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
	itemsCount := len(entries)
	switch itemsCount {
	case 0:
		return
	case 1:
		m := entries[0].Value.(*msg.MessageEnvelope)
		err := ctrl.sendWebsocket(m)
		if err != nil {
			logs.Warn("NetCtrl got error on flushing outgoing messages",
				zap.Uint64("ReqID", m.RequestID),
				zap.String("Constructor", msg.ConstructorNames[m.Constructor]),
				zap.Error(err),
			)
		}
	default:
		messages := make([]*msg.MessageEnvelope, 0, itemsCount)
		for idx := range entries {
			m := entries[idx].Value.(*msg.MessageEnvelope)
			logs.Debug("Message", zap.Int("Idx", idx), zap.String("Constructor", msg.ConstructorNames[m.Constructor]))
			messages = append(messages, m)
		}

		// Make sure messages are sorted before sending to the wire
		sort.Slice(messages, func(i, j int) bool {
			return messages[i].RequestID < messages[j].RequestID
		})

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
				logs.Warn("NetCtrl got error on flushing outgoing messages",
					zap.Error(err),
					zap.Int("Count", len(chunk)),
				)
				return
			}
			startIdx = endIdx
			endIdx += chunkSize
		}

	}

	logs.Debug("NetCtrl flushed outgoing messages",
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
		_ = ctrl.wsConn.SetReadDeadline(time.Now().Add(time.Minute * 5))
		message, messageType, err := wsutil.ReadServerData(ctrl.wsConn)
		if err != nil {
			return
		}
		switch messageType {
		case ws.OpBinary:
			// If it is a BINARY message
			err := res.Unmarshal(message)
			if err != nil {
				logs.Error("NetCtrl couldn't unmarshal received BinaryMessage", zap.Error(err), zap.String("Dump", string(message)))
				continue
			}

			if res.AuthID == 0 {
				receivedEnvelope := new(msg.MessageEnvelope)
				err = receivedEnvelope.Unmarshal(res.Payload)
				if err != nil {
					logs.Error("NetCtrl couldn't unmarshal plain-text MessageEnvelope", zap.Error(err))
					continue
				}
				logs.Debug("NetCtrl received plain-text message",
					zap.String("Constructor", msg.ConstructorNames[receivedEnvelope.Constructor]),
					zap.Uint64("ReqID", receivedEnvelope.RequestID),
				)
				ctrl.messageHandler(receivedEnvelope)
				continue
			}

			// We received an encrypted message
			decryptedBytes, err := domain.Decrypt(ctrl.authKey, res.MessageKey, res.Payload)
			if err != nil {
				logs.Error("NetCtrl couldn't decrypt the received message",
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
				logs.Error("NetCtrl couldn't unmarshal decrypted message", zap.Error(err))
				continue
			}
			logs.Debug("NetCtrl received encrypted message",
				zap.String("Constructor", msg.ConstructorNames[receivedEncryptedPayload.Envelope.Constructor]),
				zap.Uint64("ReqID", receivedEncryptedPayload.Envelope.RequestID),
			)
			// TODO:: check message id and server salt before handling the message
			ctrl.messageHandler(receivedEncryptedPayload.Envelope)
		default:
			logs.Warn("NetCtrl received unhandled message type",
				zap.Any("MessageType", messageType),
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
	case msg.C_SystemSalts:
		x := new(msg.SystemSalts)
		_ = x.Unmarshal(m.Message)
		salt.Set(x)
		fallthrough
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
	logs.Info("NetCtrl is stopping")
	ctrl.Disconnect()
	select {
	case ctrl.stopChannel <- true: // receiver may or may not be listening
	default:
	}
	logs.Info("NetCtrl stopped")
}

// Connect dial websocket
func (ctrl *Controller) Connect() {
	_, _, _ = domain.SingleFlight.Do("NetworkConnect", func() (i interface{}, e error) {
		logs.Info("NetCtrl is connecting")
		defer ctrl.recoverPanic("NetworkController:: Connect", ronak.M{
			"AuthID": ctrl.authID,
		})

		ctrl.wsKeepConnection = true
		keepGoing := true
		attempts := 0
		for keepGoing {
			ctrl.updateNetworkStatus(domain.NetworkConnecting)
			// If there is a wsConn then close it before creating a new one
			if ctrl.wsConn != nil {
				_ = ctrl.wsConn.Close()
			}

			// Stop connecting if KeepConnection is FALSE
			if !ctrl.wsKeepConnection {
				ctrl.updateNetworkStatus(domain.NetworkDisconnected)
				return
			}
			reqHdr := http.Header{}
			reqHdr.Set("X-Client-Type", fmt.Sprintf("SDK-%s-%s-%s", domain.SDKVersion, domain.ClientPlatform, domain.ClientVersion))

			// Detect the country we are calling from to determine the server Cyrus address
			if !ctrl.endpointUpdated {
				ctrl.UpdateEndpoint("")
				ctrl.endpointUpdated = true
			}

			ctrl.wsDialer.Header = ws.HandshakeHeaderHTTP(reqHdr)
			wsConn, _, _, err := ctrl.wsDialer.Dial(context.Background(), ctrl.wsEndpoint)
			if err != nil {
				logs.Warn("NetCtrl could not dial", zap.Error(err), zap.String("Url", ctrl.wsEndpoint))
				time.Sleep(domain.GetExponentialTime(100*time.Millisecond, 3*time.Second, attempts))
				attempts++
				if attempts > 5 {
					attempts = 0
					ctrl.createDialer(domain.WebsocketDialTimeoutLong)
				}
				continue
			}
			ctrl.ignoreSIGPIPE(wsConn)
			keepGoing = false
			domain.StartTime = time.Now()
			ctrl.wsConn = wsConn

			// it should be started here cuz we need receiver to get AuthRecall answer
			// SendWebsocket Signal to start the 'receiver' and 'keepAlive' routines
			ctrl.connectChannel <- true
			logs.Info("NetCtrl connected")
			ctrl.updateNetworkStatus(domain.NetworkFast)

			// Call the OnConnect handler here b4 changing network status that trigger queue to start working
			// basically we sendWebsocket priority requests b4 queue starts to work
			err = ctrl.OnWebsocketConnect()
			if !ctrl.Connected() || err != nil {
				ctrl.updateNetworkStatus(domain.NetworkConnecting)
				keepGoing = true
				continue
			}
		}
		return nil, nil
	})
}

// IgnoreSIGPIPE prevents SIGPIPE from being raised on TCP sockets when remote hangs up
// See: https://github.com/golang/go/issues/17393
func (ctrl *Controller) UpdateEndpoint(country string) {
	if country != "" {
		ctrl.countryCode = country
	}
	wsEndpointParts := strings.Split(ctrl.wsEndpoint, ".")
	httpEndpointParts := strings.Split(ctrl.httpEndpoint, ".")

	switch ctrl.countryCode {
	case "IR", "":
		wsEndpointParts[0] = strings.TrimSuffix(wsEndpointParts[0], "-cf")
		httpEndpointParts[0] = strings.TrimSuffix(httpEndpointParts[0], "-cf")
	default:
		if !strings.HasSuffix(wsEndpointParts[0], "-cf") {
			wsEndpointParts[0] = fmt.Sprintf("%s-cf", wsEndpointParts[0])
		}
		if !strings.HasSuffix(httpEndpointParts[0], "-cf") {
			httpEndpointParts[0] = fmt.Sprintf("%s-cf", httpEndpointParts[0])
		}
	}

	ctrl.wsEndpoint = strings.Join(wsEndpointParts, ".")
	ctrl.httpEndpoint = strings.Join(httpEndpointParts, ".")
	logs.Info("NetCtrl endpoints updated",
		zap.String("WS", ctrl.wsEndpoint),
		zap.String("Http", ctrl.httpEndpoint),
		zap.String("Country", ctrl.countryCode),
	)
}

// Disconnect close websocket
func (ctrl *Controller) Disconnect() {
	_, _, _ = domain.SingleFlight.Do("NetworkDisconnect", func() (i interface{}, e error) {
		ctrl.wsKeepConnection = false
		if ctrl.wsConn != nil {
			_ = ctrl.wsConn.Close()
			logs.Info("NetCtrl disconnected")
		}
		return nil, nil
	})
}

// SetAuthorization ...
// If authID and authKey are defined then sending messages will be encrypted before
// writing on the wire.
func (ctrl *Controller) SetAuthorization(authID int64, authKey []byte) {
	logs.Info("NetCtrl set authorization info", zap.Int64("AuthID", authID))
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
	startTime := time.Now()
	protoMessage := new(msg.ProtoMessage)
	protoMessage.MessageKey = make([]byte, 32)
	_, unauthorized := ctrl.unauthorizedRequests[msgEnvelope.Constructor]

	logs.Info("NetCtrl call sendWebsocket",
		zap.Uint64("ReqID", msgEnvelope.RequestID),
		zap.String("Constructor", msg.ConstructorNames[msgEnvelope.Constructor]),
		zap.Bool("Unauthorized", unauthorized),
		zap.Bool("NoAuth", ctrl.authID == 0),
	)
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
	err = wsutil.WriteClientMessage(ctrl.wsConn, ws.OpBinary, b)
	ctrl.wsWriteLock.Unlock()
	if err != nil {
		_ = ctrl.wsConn.Close()
		return err
	}
	logs.Debug("NetCtrl sent over websocket",
		zap.String("Constructor", msg.ConstructorNames[msgEnvelope.Constructor]),
		zap.Duration("Duration", time.Now().Sub(startTime)),
	)

	return nil
}

// Send encrypt and send request to server and receive and decrypt its response
func (ctrl *Controller) SendHttp(
	ctx context.Context, msgEnvelope *msg.MessageEnvelope, timeout time.Duration,
) (*msg.MessageEnvelope, error) {
	var totalUploadBytes, totalDownloadBytes int
	startTime := time.Now()

	if ctx == nil {
		ctx = context.Background()
	}
	protoMessage := msg.ProtoMessage{
		AuthID:     ctrl.authID,
		MessageKey: make([]byte, 32),
	}
	if ctrl.authID == 0 {
		protoMessage.Payload, _ = msgEnvelope.Marshal()
	} else {
		ctrl.messageSeq++
		encryptedPayload := msg.ProtoEncryptedPayload{
			ServerSalt: salt.Get(),
			Envelope:   msgEnvelope,
			MessageID:  uint64(domain.Now().Unix()<<32 | ctrl.messageSeq),
		}
		unencryptedBytes, _ := encryptedPayload.Marshal()
		encryptedPayloadBytes, _ := domain.Encrypt(ctrl.authKey, unencryptedBytes)
		messageKey := domain.GenerateMessageKey(ctrl.authKey, unencryptedBytes)
		copy(protoMessage.MessageKey, messageKey)
		protoMessage.Payload = encryptedPayloadBytes
	}

	protoMessageBytes, err := protoMessage.Marshal()
	reqBuff := bytes.NewBuffer(protoMessageBytes)
	if err != nil {
		return nil, err
	}
	totalUploadBytes += len(protoMessageBytes)

	// Set timeout
	ctrl.httpClient.Timeout = timeout

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
	totalDownloadBytes += len(resBuff)
	mon.DataTransfer(totalUploadBytes, totalDownloadBytes, time.Now().Sub(startTime))

	// Decrypt response
	res := &msg.ProtoMessage{}
	err = res.Unmarshal(resBuff)
	if err != nil {
		return nil, err
	}
	if res.AuthID == 0 {
		receivedEnvelope := &msg.MessageEnvelope{}
		err = receivedEnvelope.Unmarshal(res.Payload)
		return receivedEnvelope, err
	}
	decryptedBytes, err := domain.Decrypt(ctrl.authKey, res.MessageKey, res.Payload)

	receivedEncryptedPayload := &msg.ProtoEncryptedPayload{}
	err = receivedEncryptedPayload.Unmarshal(decryptedBytes)
	if err != nil {
		return nil, err
	}

	return receivedEncryptedPayload.Envelope, nil
}

// Reconnect by wsKeepConnection = true the watchdog will connect itself again no need to call ctrl.Connect()
func (ctrl *Controller) Reconnect() {
	_, _, _ = domain.SingleFlight.Do("NetworkReconnect", func() (i interface{}, e error) {
		if ctrl.GetQuality() == domain.NetworkDisconnected {
			ctrl.Connect()
		} else if ctrl.wsConn != nil {
			_ = ctrl.wsConn.Close()
		}
		return nil, nil
	})
}

// WaitForNetwork
func (ctrl *Controller) WaitForNetwork() {
	// Wait While Network is Disconnected or Connecting
	for ctrl.wsQuality == domain.NetworkDisconnected || ctrl.wsQuality == domain.NetworkConnecting {
		logs.Debug("NetCtrl is waiting for Network",
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
