package networkCtrl

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/internal/logs"
	mon "git.ronaksoft.com/river/sdk/internal/monitoring"
	"git.ronaksoft.com/river/sdk/internal/salt"
	"git.ronaksoft.com/river/sdk/internal/uiexec"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/ronaksoft/rony"
	"github.com/ronaksoft/rony/registry"
	"github.com/ronaksoft/rony/tools"
	"go.uber.org/zap"
	"io/ioutil"
	"net"
	"net/http"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Config network controller config
type Config struct {
	WebsocketEndpoint string
	HttpEndpoint      string
	HttpTimeout       time.Duration
	CountryCode       string
	DumpData          bool
	DumpPath          string
}

// Controller websocket network controller
type Controller struct {
	// Authorization Keys
	authID       int64
	authKey      []byte
	authRecalled int32 // atomic boolean for checking if AuthRecall is sent
	messageSeq   int64

	// Websocket Settings
	wsWriteLock      sync.Mutex
	wsDialer         *ws.Dialer
	wsEndpoint       string
	wsKeepConnection bool
	wsConn           net.Conn
	wsPingsMtx       sync.Mutex
	wsPings          map[uint64]chan struct{}
	wsQuality        int32 // atomic network quality switch

	// Http Settings
	httpEndpoint string
	httpClient   *http.Client

	// External Handlers
	OnMessage             domain.ReceivedMessageHandler
	OnUpdate              domain.ReceivedUpdateHandler
	OnGeneralError        domain.ErrorHandler
	OnWebsocketConnect    domain.OnConnectCallback
	OnNetworkStatusChange domain.NetworkStatusChangeCallback

	// internals
	connectChannel       chan bool
	stopChannel          chan bool
	endpointUpdated      bool
	unauthorizedRequests map[int64]bool // requests that it should sent unencrypted
	countryCode          string         // the country
	sendFlusher          *tools.FlusherPool
}

// New constructs the network controller
func New(config Config) *Controller {
	ctrl := new(Controller)
	ctrl.wsPings = make(map[uint64]chan struct{})
	ctrl.httpEndpoint = config.HttpEndpoint
	ctrl.countryCode = strings.ToUpper(config.CountryCode)
	if config.WebsocketEndpoint == "" {
		ctrl.wsEndpoint = domain.DefaultWebsocketEndpoint
	} else {
		ctrl.wsEndpoint = config.WebsocketEndpoint
	}
	ctrl.wsKeepConnection = true

	ctrl.createDialer(domain.WebsocketDialTimeout)
	if config.HttpTimeout == 0 {
		config.HttpTimeout = domain.HttpRequestTimeout
	}
	ctrl.createHttpClient(config.HttpTimeout)

	ctrl.stopChannel = make(chan bool, 1)
	ctrl.connectChannel = make(chan bool)
	ctrl.sendFlusher = tools.NewFlusherPool(1, 1000, ctrl.sendFlushFunc)

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
			logs.Info("NetCtrl look up for DNS",
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
func (ctrl *Controller) sendFlushFunc(targetID string, entries []tools.FlushEntry) {
	itemsCount := len(entries)
	switch itemsCount {
	case 0:
		return
	case 1:
		m := entries[0].Value().(*rony.MessageEnvelope)
		err := ctrl.writeToWebsocket(m)
		if err != nil {
			logs.Warn("NetCtrl got error on flushing outgoing messages",
				zap.Uint64("ReqID", m.RequestID),
				zap.String("C", registry.ConstructorName(m.Constructor)),
				zap.Error(err),
			)
		}
	default:
		messages := make([]*rony.MessageEnvelope, 0, itemsCount)
		for idx := range entries {
			m := entries[idx].Value().(*rony.MessageEnvelope)
			logs.Debug("Message",
				zap.Int("Idx", idx),
				zap.String("TeamID", m.Get("TeamID", "0")),
				zap.String("TeamAccess", m.Get("TeamAccess", "0")),
				zap.String("C", registry.ConstructorName(m.Constructor)),
			)
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

			msgContainer := new(rony.MessageContainer)
			msgContainer.Envelopes = chunk
			msgContainer.Length = int32(len(chunk))
			messageEnvelope := new(rony.MessageEnvelope)
			messageEnvelope.Constructor = rony.C_MessageContainer
			messageEnvelope.Message, _ = msgContainer.Marshal()
			messageEnvelope.RequestID = 0
			err := ctrl.writeToWebsocket(messageEnvelope)
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
func (ctrl *Controller) createHttpClient(timeout time.Duration) {
	ctrl.httpClient = http.DefaultClient
	ctrl.httpClient.Timeout = timeout
	ctrl.httpClient.Transport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     false,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   3 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
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
			ctrl.receiver()
			ctrl.updateNetworkStatus(domain.NetworkDisconnected)
			if ctrl.wsKeepConnection {
				go ctrl.Connect()
			}
		case <-ctrl.stopChannel:
			logs.Info("NetCtrl's watchdog stopped!")
			return
		}
	}
}

// Ping the server to check connectivity
func (ctrl *Controller) Ping(id uint64, timeout time.Duration) error {
	logs.Debug("NetCtrl pings server", zap.Uint64("ID", id))
	// Create a channel for pong
	ch := make(chan struct{}, 1)
	ctrl.wsPingsMtx.Lock()
	ctrl.wsPings[id] = ch
	ctrl.wsPingsMtx.Unlock()

	// Create a byte slice for ping payload
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, id)

	// Send the Ping to the wire
	ctrl.wsWriteLock.Lock()
	_ = ctrl.wsConn.SetWriteDeadline(time.Now().Add(domain.WebsocketWriteTime))
	err := wsutil.WriteClientMessage(ctrl.wsConn, ws.OpPing, b)
	ctrl.wsWriteLock.Unlock()
	if err != nil {
		ctrl.wsPingsMtx.Lock()
		delete(ctrl.wsPings, id)
		ctrl.wsPingsMtx.Unlock()
		return err
	}

	select {
	case <-time.After(timeout):
	case <-ch:
		// Pong Received
		return nil
	}

	return domain.ErrRequestTimeout
}

// receiver
// This is the background routine listen for incoming websocket packets and _Decrypt
// the received message, if necessary, and  pass the extracted envelopes to messageHandler.
func (ctrl *Controller) receiver() {
	defer logs.RecoverPanic(
		"NetworkController:: receiver",
		domain.M{
			"AuthID": ctrl.authID,
			"OS":     domain.ClientOS,
			"Ver":    domain.ClientVersion,
		},
		nil,
	)

	res := msg.ProtoMessage{}
	var (
		messages []wsutil.Message
		err      error
	)
	for {
		messages = messages[:0]
		_ = ctrl.wsConn.SetReadDeadline(time.Now().Add(domain.WebsocketIdleTimeout))
		messages, err = wsutil.ReadServerMessage(ctrl.wsConn, messages)
		if err != nil {
			logs.Warn("NetCtrl got error on reading server message", zap.Error(err))
			// If we return then watchdog will re-connect
			return
		}
		logs.Debug("NetCtrl received message", zap.Int("L", len(messages)))
		for _, message := range messages {
			switch message.OpCode {
			case ws.OpPong:
				logs.Info("Pong Received", zap.Int("L", len(message.Payload)))
				if len(message.Payload) == 8 {
					pingID := binary.BigEndian.Uint64(message.Payload)
					ctrl.wsPingsMtx.Lock()
					ch := ctrl.wsPings[pingID]
					delete(ctrl.wsPings, pingID)
					ctrl.wsPingsMtx.Unlock()
					if ch != nil {
						ch <- struct{}{}
					}
				}
			case ws.OpBinary:
				// If it is a BINARY message
				err := res.Unmarshal(message.Payload)
				if err != nil {
					logs.Error("NetCtrl couldn't unmarshal received BinaryMessage",
						zap.Error(err),
						zap.String("Dump", string(message.Payload)),
					)
					continue
				}

				if res.AuthID == 0 {
					receivedEnvelope := new(rony.MessageEnvelope)
					err = receivedEnvelope.Unmarshal(res.Payload)
					if err != nil {
						logs.Error("NetCtrl couldn't unmarshal plain-text MessageEnvelope", zap.Error(err))
						continue
					}
					logs.Debug("NetCtrl received plain-text message",
						zap.String("C", registry.ConstructorName(receivedEnvelope.Constructor)),
						zap.Uint64("ReqID", receivedEnvelope.RequestID),
					)
					messageHandler(ctrl, receivedEnvelope)
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
					zap.String("C", registry.ConstructorName(receivedEncryptedPayload.Envelope.Constructor)),
					zap.Uint64("ReqID", receivedEncryptedPayload.Envelope.RequestID),
				)
				// TODO:: check message id and server salt before handling the message
				messageHandler(ctrl, receivedEncryptedPayload.Envelope)
			}
		}
	}
}
func messageHandler(ctrl *Controller, message *rony.MessageEnvelope) {
	// extract all updates/ messages
	messages, updates := extractMessages(ctrl, message)
	if ctrl.OnMessage != nil {
		// sort messages by requestID
		sort.Slice(messages, func(i, j int) bool {
			return messages[i].RequestID < messages[j].RequestID
		})

		ctrl.OnMessage(messages)
	}
	if ctrl.OnUpdate != nil {
		sort.Slice(updates, func(i, j int) bool {
			return updates[i].MinUpdateID < updates[j].MinUpdateID
		})
		for _, updateContainer := range updates {
			sort.Slice(updateContainer.Updates, func(i, j int) bool {
				return updateContainer.Updates[i].UpdateID < updateContainer.Updates[j].UpdateID
			})
			ctrl.OnUpdate(updateContainer)
		}
	}
}
func extractMessages(ctrl *Controller, m *rony.MessageEnvelope) ([]*rony.MessageEnvelope, []*msg.UpdateContainer) {
	messages := make([]*rony.MessageEnvelope, 0)
	updates := make([]*msg.UpdateContainer, 0)
	switch m.Constructor {
	case rony.C_MessageContainer:
		x := new(rony.MessageContainer)
		err := x.Unmarshal(m.Message)
		if err == nil {
			for _, env := range x.Envelopes {
				msgs, upds := extractMessages(ctrl, env)
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
	case rony.C_Error:
		e := new(rony.Error)
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

// updateNetworkStatus
// The average ping times will be calculated and this function will be called if
// quality of service changed.
func (ctrl *Controller) updateNetworkStatus(newStatus domain.NetworkStatus) {
	if ctrl.GetQuality() == newStatus {
		return
	}
	ctrl.SetQuality(newStatus)
	if ctrl.OnNetworkStatusChange != nil {
		ctrl.OnNetworkStatusChange(newStatus)
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
		defer logs.RecoverPanic(
			"NetworkController::Connect",
			domain.M{
				"AuthID": ctrl.authID,
				"OS":     domain.ClientOS,
				"Ver":    domain.ClientVersion,
			},
			func() {
				ctrl.Reconnect()
			},
		)

		ctrl.wsKeepConnection = true
		keepGoing := true
		attempts := 0
		for keepGoing {
			ctrl.SetAuthRecalled(false)
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
				time.Sleep(domain.GetExponentialTime(100*time.Millisecond, 3*time.Second, attempts))
				attempts++
				if attempts > 2 {
					attempts = 0
					logs.Info("NetCtrl got error on Dial", zap.Error(err), zap.String("Endpoint", ctrl.wsEndpoint))
					ctrl.createDialer(domain.WebsocketDialTimeoutLong)
				}
				continue
			}
			ctrl.ignoreSIGPIPE(wsConn)
			keepGoing = false
			domain.StartTime = time.Now()
			ctrl.wsConn = wsConn

			// it should be started here cuz we need receiver to get AuthRecall answer
			// WebsocketSend Signal to start the 'receiver' and 'keepAlive' routines
			ctrl.connectChannel <- true
			logs.Info("NetCtrl connected")
			ctrl.updateNetworkStatus(domain.NetworkConnected)

			// Call the OnConnect handler here b4 changing network status that trigger queue to start working
			// basically we writeToWebsocket priority requests b4 queue starts to work
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

// UpdateEndpoint updates the url of the server based on the country.
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

func (ctrl *Controller) incMessageSeq() int64 {
	return atomic.AddInt64(&ctrl.messageSeq, 1)
}

// WebsocketSend if 'direct' sends immediately otherwise it put it in flusher
func (ctrl *Controller) WebsocketSend(msgEnvelope *rony.MessageEnvelope, direct bool) error {
	defer logs.RecoverPanic(
		"NetworkController::WebsocketSend",
		domain.M{
			"AuthID": ctrl.authID,
			"OS":     domain.ClientOS,
			"Ver":    domain.ClientVersion,
			"C":      msgEnvelope.Constructor,
		},
		func() {
			ctrl.Reconnect()
		},
	)

	_, unauthorized := ctrl.unauthorizedRequests[msgEnvelope.Constructor]
	if direct || unauthorized {
		return ctrl.writeToWebsocket(msgEnvelope)
	}
	ctrl.sendFlusher.Enter("", tools.NewEntry(msgEnvelope))
	return nil
}
func (ctrl *Controller) writeToWebsocket(msgEnvelope *rony.MessageEnvelope) error {
	defer logs.RecoverPanic(
		"NetworkController::writeToWebsocket",
		domain.M{
			"AuthID": ctrl.authID,
			"OS":     domain.ClientOS,
			"Ver":    domain.ClientVersion,
			"C":      msgEnvelope.Constructor,
		},
		func() {
			ctrl.Reconnect()
		},
	)

	startTime := time.Now()
	protoMessage := &msg.ProtoMessage{
		MessageKey: make([]byte, 32),
	}
	_, unauthorized := ctrl.unauthorizedRequests[msgEnvelope.Constructor]

	logs.Debug("NetCtrl call writeToWebsocket",
		zap.Uint64("ReqID", msgEnvelope.RequestID),
		zap.String("C", registry.ConstructorName(msgEnvelope.Constructor)),
		zap.String("TeamID", msgEnvelope.Get("TeamID", "0")),
		zap.String("TeamAccess", msgEnvelope.Get("TeamAccess", "0")),
		zap.Bool("Plain", unauthorized),
		zap.Bool("NoAuth", ctrl.authID == 0),
	)
	if ctrl.authID == 0 || unauthorized {
		protoMessage.AuthID = 0
		protoMessage.Payload, _ = msgEnvelope.Marshal()
	} else {
		protoMessage.AuthID = ctrl.authID
		encryptedPayload := msg.ProtoEncryptedPayload{
			ServerSalt: salt.Get(),
			Envelope:   msgEnvelope,
		}
		encryptedPayload.MessageID = uint64(domain.Now().Unix()<<32 | ctrl.incMessageSeq())
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
		zap.String("C", registry.ConstructorName(msgEnvelope.Constructor)),
		zap.Duration("Duration", time.Now().Sub(startTime)),
	)

	return nil
}

// WebsocketCommandWithTimeout run request immediately in blocking or non-blocking mode
func (ctrl *Controller) WebsocketCommandWithTimeout(
	messageEnvelope *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler,
	blockingMode, isUICallback bool, timeout time.Duration,
) {
	defer logs.RecoverPanic(
		"NetCtrl::WebsocketCommandWithTimeout",
		domain.M{
			"OS":  domain.ClientOS,
			"Ver": domain.ClientVersion,
			"C":   messageEnvelope.Constructor,
		},
		nil,
	)

	logs.Debug("NetCtrl fires websocket command",
		zap.Uint64("ReqID", messageEnvelope.RequestID),
		zap.String("C", registry.ConstructorName(messageEnvelope.Constructor)),
	)

	// Add the callback functions
	reqCB := domain.AddRequestCallback(
		messageEnvelope.RequestID, messageEnvelope.Constructor, successCB, timeout, timeoutCB, isUICallback,
	)
	execBlock := func(reqID uint64, req *rony.MessageEnvelope) {
		err := ctrl.WebsocketSend(req, blockingMode)
		if err != nil {
			logs.Warn("NetCtrl got error from NetCtrl",
				zap.String("Error", err.Error()),
				zap.String("C", registry.ConstructorName(req.Constructor)),
				zap.Uint64("ReqID", req.RequestID),
			)
			if timeoutCB != nil {
				timeoutCB()
			}
			return
		}

		select {
		case <-time.After(reqCB.Timeout):
			logs.Debug("NetCtrl got timeout on websocket command",
				zap.String("C", registry.ConstructorName(req.Constructor)),
				zap.Uint64("ReqID", req.RequestID),
			)
			domain.RemoveRequestCallback(reqID)
			if reqCB.TimeoutCallback != nil {
				if reqCB.IsUICallback {
					uiexec.ExecTimeoutCB(reqCB.TimeoutCallback)
				} else {
					reqCB.TimeoutCallback()
				}
			}
			return
		case res := <-reqCB.ResponseChannel:
			logs.Debug("NetCtrl got response for websocket command",
				zap.Uint64("ReqID", req.RequestID),
				zap.String("ReqC", registry.ConstructorName(req.Constructor)),
				zap.String("ResC", registry.ConstructorName(res.Constructor)),
			)
			if reqCB.SuccessCallback != nil {
				if reqCB.IsUICallback {
					uiexec.ExecSuccessCB(reqCB.SuccessCallback, res)
				} else {
					reqCB.SuccessCallback(res)
				}
			}
		}
		return
	}

	if blockingMode {
		execBlock(messageEnvelope.RequestID, messageEnvelope)
	} else {
		go execBlock(messageEnvelope.RequestID, messageEnvelope)
	}

	return
}

func (ctrl *Controller) WebsocketCommand(
	messageEnvelope *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler,
	blockingMode, isUICallback bool,
) {
	ctrl.WebsocketCommandWithTimeout(messageEnvelope, timeoutCB, successCB, blockingMode, isUICallback, domain.WebsocketRequestTimeout)
}

// SendHttp encrypt and send request to server and receive and decrypt its response
func (ctrl *Controller) SendHttp(ctx context.Context, msgEnvelope *rony.MessageEnvelope) (*rony.MessageEnvelope, error) {
	st := domain.Now()
	defer func() {
		logs.Info("SendHttp",
			zap.String("C", registry.ConstructorName(msgEnvelope.Constructor)),
			zap.Duration("D", domain.Now().Sub(st)),
		)
	}()

	var totalUploadBytes, totalDownloadBytes int
	startTime := tools.NanoTime()

	if ctx == nil {
		var cf context.CancelFunc
		ctx, cf = context.WithTimeout(context.Background(), domain.HttpRequestTimeout)
		defer cf()
	}
	protoMessage := msg.ProtoMessage{
		AuthID:     ctrl.authID,
		MessageKey: make([]byte, 32),
	}
	if ctrl.authID == 0 {
		protoMessage.Payload, _ = msgEnvelope.Marshal()
	} else {
		encryptedPayload := msg.ProtoEncryptedPayload{
			ServerSalt: salt.Get(),
			Envelope:   msgEnvelope,
			MessageID:  uint64(domain.Now().Unix()<<32 | ctrl.incMessageSeq()),
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
	mon.DataTransfer(totalUploadBytes, totalDownloadBytes, time.Duration(tools.NanoTime()-startTime))

	// Decrypt response
	res := &msg.ProtoMessage{}
	err = res.Unmarshal(resBuff)
	if err != nil {
		return nil, err
	}
	if res.AuthID == 0 {
		receivedEnvelope := &rony.MessageEnvelope{}
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

// HttpCommandWithTimeout run request immediately
func (ctrl *Controller) HttpCommandWithTimeout(
	messageEnvelope *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler,
	timeout time.Duration,
) {
	ctx, cf := context.WithTimeout(context.Background(), timeout)
	defer cf()

	res, err := ctrl.SendHttp(ctx, messageEnvelope)
	switch err {
	case nil:
		successCB(res)
	case context.DeadlineExceeded:
		timeoutCB()
	case context.Canceled:
		rony.ErrorMessage(res, messageEnvelope.RequestID, "E100", "Canceled")
		successCB(res)
	default:
		rony.ErrorMessage(res, messageEnvelope.RequestID, "E100", err.Error())
		successCB(res)
	}

}

// HttpCommand run request immediately
func (ctrl *Controller) HttpCommand(
	messageEnvelope *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler,
) {
	ctrl.HttpCommandWithTimeout(messageEnvelope, timeoutCB, successCB, domain.HttpRequestTimeout)
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

// WaitForNetwork this function sleeps until the websocket connection is established. If you set waitForRecall then
// it also waits the initial AuthRecall request sent to the server.
func (ctrl *Controller) WaitForNetwork(waitForRecall bool) {
	// Wait While Network is Disconnected or Connecting
	for ctrl.GetQuality() != domain.NetworkConnected {
		logs.Debug("NetCtrl is waiting for Network",
			zap.String("Quality", ctrl.GetQuality().ToString()),
		)
		time.Sleep(time.Second)
	}
	if waitForRecall {
		for !ctrl.GetAuthRecalled() {
			logs.Debug("NetCtrl is waiting for AuthRecall")
			time.Sleep(time.Second)
		}
	}
}

// Connected return the connection status for the websocket connection
func (ctrl *Controller) Connected() bool {
	return ctrl.GetQuality() == domain.NetworkConnected
}

// GetQuality returns the network status
func (ctrl *Controller) GetQuality() domain.NetworkStatus {
	return domain.NetworkStatus(atomic.LoadInt32(&ctrl.wsQuality))
}

// SetQuality sets the network status
func (ctrl *Controller) SetQuality(s domain.NetworkStatus) {
	atomic.StoreInt32(&ctrl.wsQuality, int32(s))
}

// SetAuthRecalled update the internal flag to identify AuthRecall api has been successfully called
func (ctrl *Controller) SetAuthRecalled(b bool) {
	if b {
		atomic.StoreInt32(&ctrl.authRecalled, 1)
	} else {
		atomic.StoreInt32(&ctrl.authRecalled, 0)
	}
}

// GetAuthRecalled read the internal flag to check if AuthRecall has been called
func (ctrl *Controller) GetAuthRecalled() bool {
	return atomic.LoadInt32(&ctrl.authRecalled) > 0
}
