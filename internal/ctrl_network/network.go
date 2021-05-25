package networkCtrl

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/internal/logs"
	mon "git.ronaksoft.com/river/sdk/internal/monitoring"
	"git.ronaksoft.com/river/sdk/internal/request"
	"git.ronaksoft.com/river/sdk/internal/salt"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/ronaksoft/rony"
	"github.com/ronaksoft/rony/errors"
	"github.com/ronaksoft/rony/pools"
	"github.com/ronaksoft/rony/registry"
	"github.com/ronaksoft/rony/tools"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"io/ioutil"
	"net"
	"net/http"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var (
	logger *logs.Logger
)

func init() {
	logger = logs.With("NetCtrl")
}

// Config network controller config
type Config struct {
	SeedHosts   []string
	HttpTimeout time.Duration
	CountryCode string
	DumpData    bool
	DumpPath    string
}

// Controller websocket network controller
type Controller struct {
	// Authorization Keys
	authID        int64
	authKey       []byte
	authRecalled  int32 // atomic boolean for checking if AuthRecall is sent
	messageSeq    int64
	endPoints     []string
	curEndpoint   string
	curEndpointIP string

	// Websocket Settings
	wsWriteLock      sync.Mutex
	wsDialer         *ws.Dialer
	wsKeepConnection bool
	wsConn           net.Conn
	wsPingsMtx       sync.Mutex
	wsPings          map[uint64]chan struct{}
	wsQuality        int32 // atomic network quality switch

	// Http Settings
	httpClient *http.Client

	// External Handlers
	OnGeneralError        domain.ErrorHandler
	OnWebsocketConnect    domain.OnConnectCallback
	OnNetworkStatusChange domain.NetworkStatusChangeCallback

	// External Processors
	MessageChan chan []*rony.MessageEnvelope
	UpdateChan  chan *msg.UpdateContainer

	// internals
	connectChannel       chan bool
	stopChannel          chan bool
	unauthorizedRequests map[int64]bool // requests that it should sent unencrypted
	countryCode          string         // the country
	sendRoutines         map[string]*tools.FlusherPool
}

// New constructs the network controller
func New(config Config) *Controller {
	ctrl := &Controller{
		wsPings:          make(map[uint64]chan struct{}),
		endPoints:        config.SeedHosts,
		countryCode:      strings.ToUpper(config.CountryCode),
		wsKeepConnection: true,
		stopChannel:      make(chan bool, 1),
		connectChannel:   make(chan bool),
	}

	if config.HttpTimeout == 0 {
		config.HttpTimeout = domain.HttpRequestTimeout
	}

	ctrl.createWebsocketDialer(domain.WebsocketDialTimeout)
	ctrl.createHttpClient(config.HttpTimeout)
	ctrl.sendRoutines = map[string]*tools.FlusherPool{
		"Messages": tools.NewFlusherPool(1, 250, ctrl.sendFlushFunc),
		"General":  tools.NewFlusherPool(3, 250, ctrl.sendFlushFunc),
		"Batch":    tools.NewFlusherPoolWithWaitTime(1, 250, 250*time.Millisecond, ctrl.sendFlushFunc),
	}
	ctrl.unauthorizedRequests = map[int64]bool{
		msg.C_SystemGetServerTime: true,
		msg.C_InitConnect:         true,
		msg.C_InitCompleteAuth:    true,
		msg.C_SystemGetSalts:      true,
	}

	return ctrl
}
func (ctrl *Controller) createWebsocketDialer(timeout time.Duration) {
	ctrl.wsDialer = &ws.Dialer{
		ReadBufferSize:  32 * 1024, // 32kB
		WriteBufferSize: 32 * 1024, // 32kB
		Timeout:         timeout,
		NetDial: func(ctx context.Context, network, addr string) (conn net.Conn, err error) {
			host, port, err := net.SplitHostPort(addr)
			if err != nil {
				return nil, err
			}
			d := net.Dialer{Timeout: timeout}
			if ctrl.curEndpointIP != "" {
				conn, err = d.DialContext(ctx, "tcp4", net.JoinHostPort(ctrl.curEndpointIP, port))
				if err == nil {
					return
				}
			}

			ips, err := net.LookupIP(host)
			if err != nil {
				return nil, err
			}
			logger.Info("look up for DNS", zap.String("Addr", addr), zap.Any("IPs", ips))

			for _, ip := range ips {
				if ip.To4() != nil {
					conn, err = d.DialContext(ctx, "tcp4", net.JoinHostPort(ip.String(), port))
					if err != nil {
						continue
					}
					ctrl.curEndpointIP = ip.String()
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
func (ctrl *Controller) sendFlushFunc(targetID string, entries []tools.FlushEntry) {
	itemsCount := len(entries)
	switch itemsCount {
	case 0:
		return
	case 1:
		m := entries[0].Value().(*rony.MessageEnvelope)
		err := ctrl.writeToWebsocket(m)
		if err != nil {
			logger.Warn("got error on writing to websocket conn",
				zap.Uint64("ReqID", m.RequestID),
				zap.String("C", registry.ConstructorName(m.Constructor)),
				zap.Error(err),
			)
		}
	default:
		messages := make([]*rony.MessageEnvelope, 0, itemsCount)
		for idx := range entries {
			m := entries[idx].Value().(*rony.MessageEnvelope)
			logger.Debug("Message",
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

			msgContainer := &rony.MessageContainer{
				Envelopes: chunk,
				Length:    int32(len(chunk)),
			}
			messageEnvelope := &rony.MessageEnvelope{}
			messageEnvelope.Fill(0, rony.C_MessageContainer, msgContainer)

			err := ctrl.writeToWebsocket(messageEnvelope)
			if err != nil {
				logger.Warn("got error on writing to websocket conn",
					zap.String("C", registry.ConstructorName(messageEnvelope.Constructor)),
					zap.Int("Count", len(chunk)),
					zap.Error(err),
				)
				return
			}
			startIdx = endIdx
			endIdx += chunkSize
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
			ctrl.receiver()
			ctrl.updateNetworkStatus(domain.NetworkDisconnected)
			if ctrl.wsKeepConnection {
				go ctrl.Connect()
			}
		case <-ctrl.stopChannel:
			logger.Info("NetCtrl's watchdog stopped!")
			return
		}
	}
}

// Ping the server to check connectivity
func (ctrl *Controller) Ping(id uint64, timeout time.Duration) error {
	logger.Debug("pings server", zap.Uint64("ID", id))
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
	defer logger.RecoverPanic(
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
			logger.Warn("got error on reading server message", zap.Error(err))
			// If we return then watchdog will re-connect
			return
		}
		logger.Debug("received message", zap.Int("L", len(messages)))
		for _, message := range messages {
			switch message.OpCode {
			case ws.OpPong:
				logger.Info("Pong Received", zap.Int("L", len(message.Payload)))
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
					logger.Error("couldn't unmarshal received BinaryMessage",
						zap.Error(err),
						zap.String("Dump", string(message.Payload)),
					)
					continue
				}

				if res.AuthID == 0 {
					receivedEnvelope := new(rony.MessageEnvelope)
					err = receivedEnvelope.Unmarshal(res.Payload)
					if err != nil {
						logger.Error("couldn't unmarshal plain-text MessageEnvelope", zap.Error(err))
						continue
					}
					logger.Debug("received plain-text message",
						zap.String("C", registry.ConstructorName(receivedEnvelope.Constructor)),
						zap.Uint64("ReqID", receivedEnvelope.RequestID),
					)
					messageHandler(ctrl, receivedEnvelope)
					continue
				}

				// We received an encrypted message
				decryptedBytes, err := domain.Decrypt(ctrl.authKey, res.MessageKey, res.Payload)
				if err != nil {
					logger.Error("couldn't decrypt the received message",
						zap.String("Error", err.Error()),
						zap.Int64("ctrl.authID", ctrl.authID),
						zap.Int64("resp.AuthID", res.AuthID),
						zap.String("resp.MessageKey", hex.Dump(res.MessageKey)),
					)
					continue
				}
				receivedEncryptedPayload := &msg.ProtoEncryptedPayload{}
				err = receivedEncryptedPayload.Unmarshal(decryptedBytes)
				pools.Bytes.Put(decryptedBytes)
				if err != nil {
					logger.Error("couldn't unmarshal decrypted message", zap.Error(err))
					continue
				}
				logger.Debug("received encrypted message",
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
	// sort messages by requestID
	sort.Slice(messages, func(i, j int) bool {
		return messages[i].RequestID < messages[j].RequestID
	})
	ctrl.MessageChan <- messages

	sort.Slice(updates, func(i, j int) bool {
		return updates[i].MinUpdateID < updates[j].MinUpdateID
	})
	for _, updateContainer := range updates {
		sort.Slice(updateContainer.Updates, func(i, j int) bool {
			return updateContainer.Updates[i].UpdateID < updateContainer.Updates[j].UpdateID
		})
		ctrl.UpdateChan <- updateContainer
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
	if ctrl.GetStatus() == newStatus {
		return
	}
	ctrl.setStatus(newStatus)
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
	logger.Info("is stopping")
	ctrl.Disconnect()
	select {
	case ctrl.stopChannel <- true: // receiver may or may not be listening
	default:
	}
	logger.Info("stopped")
}

// Connect dial websocket
func (ctrl *Controller) Connect() {
	_, _, _ = domain.SingleFlight.Do("NetworkConnect", func() (i interface{}, e error) {
		logger.Info("is connecting")
		defer logger.RecoverPanic(
			"NetCtrl::Connect",
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

			ctrl.UpdateEndpoint("")
			ctrl.wsDialer.Header = ws.HandshakeHeaderHTTP(reqHdr)
			wsConn, _, _, err := ctrl.wsDialer.Dial(context.Background(), fmt.Sprintf("ws://%s", ctrl.curEndpoint))
			if err != nil {
				time.Sleep(domain.GetExponentialTime(100*time.Millisecond, 3*time.Second, attempts))
				attempts++
				if attempts > 5 {
					attempts = 0
					logger.Info("got error on Dial", zap.Error(err), zap.String("Endpoint", ctrl.curEndpoint))
					ctrl.createWebsocketDialer(domain.WebsocketDialTimeoutLong)
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
			logger.Info("connected")
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

	ctrl.curEndpoint = ctrl.endPoints[tools.RandomInt(len(ctrl.endPoints))]

	endpointParts := strings.Split(ctrl.curEndpoint, ".")
	switch ctrl.countryCode {
	case "IR", "":
		endpointParts[0] = strings.TrimSuffix(endpointParts[0], "-cf")
	default:
		if !strings.HasSuffix(endpointParts[0], "-cf") {
			endpointParts[0] = fmt.Sprintf("%s-cf", endpointParts[0])
		}
	}

	ctrl.curEndpoint = strings.Join(endpointParts, ".")
	logger.Info("endpoints updated",
		zap.String("WS", ctrl.curEndpoint),
		zap.String("Http", ctrl.curEndpoint),
		zap.String("Country", ctrl.countryCode),
	)
}

// Disconnect close websocket
func (ctrl *Controller) Disconnect() {
	_, _, _ = domain.SingleFlight.Do("NetworkDisconnect", func() (i interface{}, e error) {
		ctrl.wsKeepConnection = false
		if ctrl.wsConn != nil {
			err := ctrl.wsConn.Close()
			logger.Info("websocket conn closed", zap.Error(err))
		}
		return nil, nil
	})
}

// SetAuthorization ...
// If authID and authKey are defined then sending messages will be encrypted before
// writing on the wire.
func (ctrl *Controller) SetAuthorization(authID int64, authKey []byte) {
	logger.Info("set authorization info", zap.Int64("AuthID", authID))
	ctrl.authKey = make([]byte, len(authKey))
	ctrl.authID = authID
	copy(ctrl.authKey, authKey)
}

func (ctrl *Controller) incMessageSeq() int64 {
	return atomic.AddInt64(&ctrl.messageSeq, 1)
}

// WebsocketSend if 'direct' sends immediately otherwise it put it in flusher
func (ctrl *Controller) WebsocketSend(msgEnvelope *rony.MessageEnvelope, flag request.DelegateFlag) error {
	defer logger.RecoverPanic(
		"NetCtrl::WebsocketSend",
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
	if flag&request.SkipFlusher != 0 || unauthorized {
		return ctrl.writeToWebsocket(msgEnvelope)
	}
	switch msgEnvelope.Constructor {
	case msg.C_MessagesSend, msg.C_MessagesSendMedia, msg.C_MessagesForward:
		ctrl.sendRoutines["Messages"].Enter("", tools.NewEntry(msgEnvelope))
	default:
		if flag&request.Batch != 0 {
			ctrl.sendRoutines["Batch"].Enter("", tools.NewEntry(msgEnvelope))
		} else {
			ctrl.sendRoutines["General"].Enter("", tools.NewEntry(msgEnvelope))
		}

	}

	return nil
}
func (ctrl *Controller) writeToWebsocket(msgEnvelope *rony.MessageEnvelope) error {
	defer logger.RecoverPanic(
		"NetCtrl::writeToWebsocket",
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

	startTime := tools.NanoTime()
	protoMessage := msg.PoolProtoMessage.Get()
	defer msg.PoolProtoMessage.Put(protoMessage)
	if cap(protoMessage.MessageKey) < 32 {
		protoMessage.MessageKey = make([]byte, 32)
	} else {
		protoMessage.MessageKey = protoMessage.MessageKey[:32]
	}

	logger.Debug("writing to the websocket conn",
		zap.Uint64("ReqID", msgEnvelope.RequestID),
		zap.String("C", registry.ConstructorName(msgEnvelope.Constructor)),
		zap.String("TeamID", msgEnvelope.Get("TeamID", "0")),
		zap.Bool("SendPlain", ctrl.unauthorizedRequests[msgEnvelope.Constructor]),
		zap.Bool("Authorized", ctrl.authID != 0),
	)
	if ctrl.authID == 0 || ctrl.unauthorizedRequests[msgEnvelope.Constructor] {
		protoMessage.AuthID = 0
		buf := pools.Buffer.FromProto(msgEnvelope)
		protoMessage.Payload = buf.AppendTo(protoMessage.Payload)
		pools.Buffer.Put(buf)
	} else {
		protoMessage.AuthID = ctrl.authID
		encryptedPayload := &msg.ProtoEncryptedPayload{
			ServerSalt: salt.Get(),
			Envelope:   msgEnvelope,
			MessageID:  uint64(domain.Now().Unix()<<32 | ctrl.incMessageSeq()),
		}

		mo := proto.MarshalOptions{UseCachedSize: true}
		unencryptedBytes := pools.Bytes.GetCap(mo.Size(encryptedPayload))
		unencryptedBytes, _ = mo.MarshalAppend(unencryptedBytes, encryptedPayload)
		encryptedPayloadBytes, _ := domain.Encrypt(ctrl.authKey, unencryptedBytes)
		messageKey := domain.GenerateMessageKey(ctrl.authKey, unencryptedBytes)
		copy(protoMessage.MessageKey, messageKey)
		protoMessage.Payload = append(protoMessage.Payload[:0], encryptedPayloadBytes...)
		pools.Bytes.Put(encryptedPayloadBytes)
	}

	reqBuff := pools.Buffer.FromProto(protoMessage)
	defer pools.Buffer.Put(reqBuff)

	if ctrl.wsConn == nil {
		return domain.ErrNoConnection
	}

	ctrl.wsWriteLock.Lock()
	_ = ctrl.wsConn.SetWriteDeadline(time.Now().Add(domain.WebsocketWriteTime))
	err := wsutil.WriteClientMessage(ctrl.wsConn, ws.OpBinary, *reqBuff.Bytes())
	ctrl.wsWriteLock.Unlock()
	if err != nil {
		_ = ctrl.wsConn.Close()
		return err
	}

	logger.Debug("write to websocket completed.",
		zap.Uint64("ReqID", msgEnvelope.RequestID),
		zap.String("C", registry.ConstructorName(msgEnvelope.Constructor)),
		zap.Duration("Duration", time.Duration(tools.NanoTime()-startTime)),
	)

	return nil
}

// WebsocketCommandWithTimeout run request immediately in blocking or non-blocking mode
func (ctrl *Controller) WebsocketCommandWithTimeout(
	messageEnvelope *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler,
	isUICallback bool, flag request.DelegateFlag, timeout time.Duration,
) {
	defer logger.RecoverPanic(
		"NetCtrl::WebsocketCommandWithTimeout",
		domain.M{
			"OS":  domain.ClientOS,
			"Ver": domain.ClientVersion,
			"C":   messageEnvelope.Constructor,
		},
		nil,
	)

	logger.Debug("execute command over websocket",
		zap.Uint64("ReqID", messageEnvelope.RequestID),
		zap.String("C", registry.ConstructorName(messageEnvelope.Constructor)),
	)

	// Add the callback functions
	reqCB := request.AddRequestCallback(
		messageEnvelope.RequestID, messageEnvelope.Constructor, successCB, timeout, timeoutCB, isUICallback,
	)

	execBlock := func(reqID uint64, req *rony.MessageEnvelope) {
		err := ctrl.WebsocketSend(req, flag)
		if err != nil {
			logger.Warn("got error from NetCtrl",
				zap.Uint64("ReqID", req.RequestID),
				zap.String("C", registry.ConstructorName(req.Constructor)),
				zap.Error(err),
			)
			if timeoutCB != nil {
				timeoutCB()
			}
			return
		}

		select {
		case <-time.After(reqCB.Timeout):
			logger.Debug("got timeout on websocket command",
				zap.String("C", registry.ConstructorName(req.Constructor)),
				zap.Uint64("ReqID", req.RequestID),
			)
			request.RemoveRequestCallback(reqID)
			reqCB.OnTimeout()
			return
		case res := <-reqCB.ResponseChannel:
			logger.Debug("got response for websocket command",
				zap.Uint64("ReqID", req.RequestID),
				zap.String("ReqC", registry.ConstructorName(req.Constructor)),
				zap.String("ResC", registry.ConstructorName(res.Constructor)),
			)
			reqCB.OnSuccess(res)
		}
		return
	}
	execBlock(messageEnvelope.RequestID, messageEnvelope)
	return
}

func (ctrl *Controller) WebsocketCommand(
	messageEnvelope *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler,
	isUICallback bool, flag request.DelegateFlag,
) {
	ctrl.WebsocketCommandWithTimeout(messageEnvelope, timeoutCB, successCB, isUICallback, flag, domain.WebsocketRequestTimeout)
}

// SendHttp encrypt and send request to server and receive and decrypt its response
func (ctrl *Controller) SendHttp(ctx context.Context, msgEnvelope *rony.MessageEnvelope) (*rony.MessageEnvelope, error) {
	var (
		totalUploadBytes, totalDownloadBytes int
		startTime                            = tools.NanoTime()
	)

	if ctx == nil {
		var cf context.CancelFunc
		ctx, cf = context.WithTimeout(context.Background(), domain.HttpRequestTimeout)
		defer cf()
	}

	protoMessage := msg.PoolProtoMessage.Get()
	defer msg.PoolProtoMessage.Put(protoMessage)
	if cap(protoMessage.MessageKey) < 32 {
		protoMessage.MessageKey = make([]byte, 32)
	} else {
		protoMessage.MessageKey = protoMessage.MessageKey[:32]
	}
	if ctrl.unauthorizedRequests[msgEnvelope.Constructor] {
		protoMessage.AuthID = 0
	} else {
		protoMessage.AuthID = ctrl.authID
	}

	if protoMessage.AuthID == 0 {
		buf := pools.Buffer.FromProto(msgEnvelope)
		protoMessage.Payload = buf.AppendTo(protoMessage.Payload)
		pools.Buffer.Put(buf)
	} else {
		encryptedPayload := &msg.ProtoEncryptedPayload{
			ServerSalt: salt.Get(),
			Envelope:   msgEnvelope,
			MessageID:  uint64(domain.Now().Unix()<<32 | ctrl.incMessageSeq()),
		}
		mo := proto.MarshalOptions{UseCachedSize: true}
		unencryptedBytes := pools.Bytes.GetCap(mo.Size(encryptedPayload))
		unencryptedBytes, _ = mo.MarshalAppend(unencryptedBytes, encryptedPayload)
		encryptedPayloadBytes, _ := domain.Encrypt(ctrl.authKey, unencryptedBytes)
		messageKey := domain.GenerateMessageKey(ctrl.authKey, unencryptedBytes)
		pools.Bytes.Put(unencryptedBytes)
		copy(protoMessage.MessageKey, messageKey)
		protoMessage.Payload = append(protoMessage.Payload[:0], encryptedPayloadBytes...)
		pools.Bytes.Put(encryptedPayloadBytes)

	}

	reqBuff := pools.Buffer.FromProto(protoMessage)
	totalUploadBytes += reqBuff.Len() // len(reqBuff.Len()protoMessageBytes)

	// Send Data
	httpReq, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s", ctrl.curEndpoint), reqBuff)
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
	pools.Bytes.Put(decryptedBytes)
	if err != nil {
		return nil, err
	}

	logger.Info("SendHttp",
		zap.String("URL", ctrl.curEndpoint),
		zap.String("ReqC", registry.ConstructorName(msgEnvelope.Constructor)),
		zap.String("ResC", registry.ConstructorName(receivedEncryptedPayload.Envelope.Constructor)),
		zap.Duration("D", time.Duration(tools.NanoTime()-startTime)),
	)

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
		res := &rony.MessageEnvelope{}
		errors.New("E100", "Canceled").ToEnvelope(res)
		successCB(res)
	default:
		res := &rony.MessageEnvelope{}
		errors.New("E100", err.Error()).ToEnvelope(res)
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
		if ctrl.GetStatus() == domain.NetworkDisconnected {
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
	for ctrl.GetStatus() != domain.NetworkConnected {
		logger.Debug("is waiting for Network",
			zap.String("Quality", ctrl.GetStatus().ToString()),
		)
		time.Sleep(time.Second)
	}
	if waitForRecall {
		for !ctrl.GetAuthRecalled() {
			logger.Debug("is waiting for AuthRecall")
			time.Sleep(time.Second)
		}
	}
}

// Connected return the connection status for the websocket connection
func (ctrl *Controller) Connected() bool {
	return ctrl.GetStatus() == domain.NetworkConnected
}

func (ctrl *Controller) Disconnected() bool {
	return ctrl.GetStatus() == domain.NetworkDisconnected
}

// GetStatus returns the network status
func (ctrl *Controller) GetStatus() domain.NetworkStatus {
	return domain.NetworkStatus(atomic.LoadInt32(&ctrl.wsQuality))
}

func (ctrl *Controller) setStatus(s domain.NetworkStatus) {
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
