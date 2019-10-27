package queueCtrl

import (
	mon "git.ronaksoftware.com/ronak/riversdk/pkg/monitoring"
	"sync"
	"time"

	"git.ronaksoftware.com/ronak/riversdk/pkg/uiexec"

	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
	"git.ronaksoftware.com/ronak/riversdk/pkg/ctrl_network"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo"
	"github.com/beeker1121/goque"
	"github.com/juju/ratelimit"
	"go.uber.org/zap"
)

// easyjson:json
// request
type request struct {
	ID              uint64               `json:"id"`
	Timeout         time.Duration        `json:"timeout"`
	MessageEnvelope *msg.MessageEnvelope `json:"message_envelope"`
	InsertTime      time.Time            `json:"insert_time"`
}

// Controller ...
// This controller will be connected to networkController and messages will be queued here
// before passing to the networkController.
type Controller struct {
	dataDir     string
	rateLimiter *ratelimit.Bucket
	waitingList *goque.Queue
	network     *networkCtrl.Controller

	// Internal Flags
	distributorLock    sync.Mutex
	distributorRunning bool

	// Cancelled request
	cancelLock       sync.Mutex
	cancelledRequest map[int64]bool
}

// New
func New(network *networkCtrl.Controller, dataDir string) (*Controller, error) {
	ctrl := new(Controller)
	ctrl.dataDir = dataDir
	ctrl.rateLimiter = ratelimit.NewBucket(time.Second, 20)
	if dataDir == "" {
		return nil, domain.ErrQueuePathIsNotSet
	}

	ctrl.cancelledRequest = make(map[int64]bool)
	ctrl.network = network
	return ctrl, nil
}

// distributor
// Pulls the next request from the waitingList and pass it to the executor. It uses
// a rate limiter to throttle the throughput
func (ctrl *Controller) distributor() {
	for {
		// Wait until network is available
		ctrl.network.WaitForNetwork()

		ctrl.distributorLock.Lock()
		if ctrl.waitingList.Length() == 0 {
			ctrl.distributorRunning = false
			ctrl.distributorLock.Unlock()
			break
		}
		ctrl.distributorLock.Unlock()

		// Peek item from the queue
		item, err := ctrl.waitingList.Dequeue()
		if err != nil {
			continue
		}

		// Prepare
		req := request{}
		if err := req.UnmarshalJSON(item.Value); err != nil {
			logs.Error("QueueController could not unmarshal popped request", zap.Error(err))
			continue
		}

		mon.QueueTime(req.MessageEnvelope.Constructor, time.Now().Sub(req.InsertTime))
		if !ctrl.IsRequestCancelled(int64(req.ID)) {
			go ctrl.executor(req)
		} else {
			logs.Info("QueueController discarded a canceled request",
				zap.Uint64("RequestID", req.ID),
				zap.String("RequestName", msg.ConstructorNames[req.MessageEnvelope.Constructor]),
			)
		}
	}
}

// addToWaitingList
func (ctrl *Controller) addToWaitingList(req *request) {
	req.InsertTime = time.Now()
	jsonRequest, err := req.MarshalJSON()
	if err != nil {
		logs.Warn("QueueController couldn't marshal the request", zap.Error(err))
		return
	}
	if _, err := ctrl.waitingList.Enqueue(jsonRequest); err != nil {
		logs.Warn("QueueController couldn't enqueue the request", zap.Error(err))
		return
	}
	ctrl.distributorLock.Lock()
	if !ctrl.distributorRunning {
		ctrl.distributorRunning = true
		go ctrl.distributor()
	}
	ctrl.distributorLock.Unlock()
}

// executor
// Sends the message to the networkController and waits for the response. If time is up then it call the
// TimeoutCallback otherwise if response arrived in time, SuccessCallback will be called.
func (ctrl *Controller) executor(req request) {
	reqCallbacks := domain.GetRequestCallback(req.ID)
	if reqCallbacks == nil {
		reqCallbacks = domain.AddRequestCallback(
			req.ID,
			nil,
			req.Timeout,
			nil,
			true,
		)
	}
	if req.Timeout == 0 {
		req.Timeout = domain.WebsocketRequestTime
	}

	// Try to send it over wire, if error happened put it back into the queue
	if err := ctrl.network.SendWebsocket(req.MessageEnvelope, false); err != nil {
		logs.Error("QueueCtrl got error from NetCtrl", zap.Error(err))
		logs.Info("QueueCtrl re-push the request into the queue")
		ctrl.addToWaitingList(&req)
		return
	}

	select {
	case <-time.After(req.Timeout):
		switch req.MessageEnvelope.Constructor {
		case msg.C_MessagesSend, msg.C_MessagesSendMedia:
			pmsg, err := repo.PendingMessages.GetByRandomID(int64(req.ID))
			if err == nil && pmsg != nil {
				ctrl.addToWaitingList(&req)
				return
			}
		case msg.C_MessagesReadHistory, msg.C_MessagesGetHistory, msg.C_ContactsImport, msg.C_ContactsGet,
			msg.C_AuthSendCode, msg.C_AuthRegister, msg.C_AuthLogin:
			ctrl.addToWaitingList(&req)
			return
		default:
			if reqCallbacks.TimeoutCallback != nil {
				if reqCallbacks.IsUICallback {
					uiexec.Ctx().Exec(func() { reqCallbacks.TimeoutCallback() })
				} else {
					reqCallbacks.TimeoutCallback()
				}
			}
		}
	case res := <-reqCallbacks.ResponseChannel:
		switch req.MessageEnvelope.Constructor {
		case msg.C_MessagesSend, msg.C_MessagesSendMedia:
			switch res.Constructor {
			case msg.C_Error:
				errMsg := new(msg.Error)
				_ = errMsg.Unmarshal(res.Message)
				if errMsg.Code == msg.ErrCodeAlreadyExists && errMsg.Items == msg.ErrItemRandomID {
					pm, _ := repo.PendingMessages.GetByRandomID(int64(req.ID))
					if pm != nil {
						_ = repo.PendingMessages.Delete(pm.ID)
					}
				}
			}
		default:
			switch res.Constructor {
			case msg.C_Error:
				errMsg := new(msg.Error)
				_ = errMsg.Unmarshal(res.Message)
				if errMsg.Code == msg.ErrCodeInvalid && errMsg.Items == msg.ErrItemSalt {
					ctrl.addToWaitingList(&req)
					return
				}
			}
		}
		if reqCallbacks.SuccessCallback != nil {
			if reqCallbacks.IsUICallback {
				uiexec.Ctx().Exec(func() { reqCallbacks.SuccessCallback(res) })
			} else {
				reqCallbacks.SuccessCallback(res)
			}
		} else {
			logs.Warn("QueueCtrl received response but no callback exists!!!",
				zap.String("Constructor", msg.ConstructorNames[res.Constructor]),
				zap.Uint64("RequestID", res.RequestID),
			)
		}
	}
	domain.RemoveRequestCallback(req.ID)
	return
}

// RealtimeCommand run request immediately and do not save it in queue
func (ctrl *Controller) RealtimeCommand(requestID uint64, constructor int64, commandBytes []byte, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler, blockingMode, isUICallback bool) {
	logs.Debug("QueueCtrl fires realtime command",
		zap.Uint64("ReqID", requestID), zap.String("Constructor", msg.ConstructorNames[constructor]),
	)
	messageEnvelope := new(msg.MessageEnvelope)
	messageEnvelope.Constructor = constructor
	messageEnvelope.RequestID = requestID
	messageEnvelope.Message = commandBytes

	// Add the callback functions
	reqCB := domain.AddRequestCallback(requestID, successCB, domain.WebsocketRequestTime, timeoutCB, isUICallback)

	execBlock := func(reqID uint64, req *msg.MessageEnvelope) {
		err := ctrl.network.SendWebsocket(req, blockingMode)
		if err != nil {
			logs.Warn("QueueCtrl got error from NetCtrl",
				zap.String("Error", err.Error()),
				zap.String("Constructor", msg.ConstructorNames[req.Constructor]),
				zap.Uint64("ReqID", requestID),
			)
			if timeoutCB != nil {
				timeoutCB()
			}
			return
		}

		select {
		case <-time.After(reqCB.Timeout):
			logs.Debug("QueueCtrl got timeout on realtime command",
				zap.String("Constructor", msg.ConstructorNames[req.Constructor]),
				zap.Uint64("ReqID", requestID),
			)
			domain.RemoveRequestCallback(reqID)
			if reqCB.TimeoutCallback != nil {
				if reqCB.IsUICallback {
					uiexec.Ctx().Exec(func() { reqCB.TimeoutCallback() })
				} else {
					reqCB.TimeoutCallback()
				}
			}
			return
		case res := <-reqCB.ResponseChannel:
			logs.Debug("QueueCtrl got response on realtime command",
				zap.Uint64("ReqID", requestID),
				zap.String("Req", msg.ConstructorNames[req.Constructor]),
				zap.String("Res", msg.ConstructorNames[res.Constructor]),

			)
			if reqCB.SuccessCallback != nil {
				if reqCB.IsUICallback {
					uiexec.Ctx().Exec(func() { reqCB.SuccessCallback(res) })
				} else {
					reqCB.SuccessCallback(res)
				}
			}
		}
		return
	}

	if blockingMode {
		execBlock(requestID, messageEnvelope)
	} else {
		go execBlock(requestID, messageEnvelope)
	}

	return
}

// EnqueueCommand put request in queue and distributor will execute it later
func (ctrl *Controller) EnqueueCommand(requestID uint64, constructor int64, requestBytes []byte, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler, isUICallback bool) {
	logs.Debug("QueueCtrl enqueues command",
		zap.Uint64("ReqID", requestID), zap.String("Constructor", msg.ConstructorNames[constructor]),
	)
	messageEnvelope := new(msg.MessageEnvelope)
	messageEnvelope.RequestID = requestID
	messageEnvelope.Constructor = constructor
	messageEnvelope.Message = requestBytes
	req := request{
		ID:              requestID,
		Timeout:         domain.WebsocketRequestTime,
		MessageEnvelope: messageEnvelope,
	}

	// Add the callback functions
	domain.AddRequestCallback(requestID, successCB, req.Timeout, timeoutCB, isUICallback)

	// Add the request to the queue
	ctrl.addToWaitingList(&req)
}

// Start queue
func (ctrl *Controller) Start() {
	logs.Info("QueueCtrl started")
	if q, err := goque.OpenQueue(ctrl.dataDir); err != nil {
		logs.Fatal("We couldn't initialize the queue", zap.Error(err))
	} else {
		ctrl.waitingList = q
	}

	// Try to resend unsent messages
	for _, pmsg := range repo.PendingMessages.GetAll() {
		switch pmsg.MediaType {
		case msg.InputMediaTypeEmpty:
			// it will be MessagesSend
			req := repo.PendingMessages.ToMessagesSend(pmsg)
			reqBytes, _ := req.Marshal()
			ctrl.EnqueueCommand(uint64(req.RandomID), msg.C_MessagesSend, reqBytes, nil, nil, false)
		default:
			// it will be MessagesSendMedia
			req := repo.PendingMessages.ToMessagesSendMedia(pmsg)
			if req == nil {
				continue
			}
			reqBytes, _ := req.Marshal()
			ctrl.EnqueueCommand(uint64(req.RandomID), msg.C_MessagesSendMedia, reqBytes, nil, nil, false)
		}
	}

	go ctrl.distributor()
}

// Stop queue
func (ctrl *Controller) Stop() {
	logs.Info("QueueCtrl stopped")
	ctrl.waitingList.Close()
}

// IsRequestCancelled handle canceled request to do not process its response
func (ctrl *Controller) IsRequestCancelled(reqID int64) bool {
	_, ok := ctrl.cancelledRequest[reqID]
	if ok {
		ctrl.cancelLock.Lock()
		delete(ctrl.cancelledRequest, reqID)
		ctrl.cancelLock.Unlock()
	}
	return ok
}

// CancelRequest cancel request
func (ctrl *Controller) CancelRequest(reqID int64) {
	ctrl.cancelLock.Lock()
	ctrl.cancelledRequest[reqID] = true
	ctrl.cancelLock.Unlock()
}

// DropQueue remove queue from storage
func (ctrl *Controller) DropQueue() {
	ctrl.waitingList.Drop()
}

// OpenQueue init queue files in storage
func (ctrl *Controller) OpenQueue(dataDir string) (err error) {
	ctrl.waitingList, err = goque.OpenQueue(dataDir)
	return
}
