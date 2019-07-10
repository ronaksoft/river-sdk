package queueCtrl

import (
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
}

// Controller ...
// This controller will be connected to networkController and messages will be queued here
// before passing to the networkController.
type Controller struct {
	distributorLock sync.Mutex

	rateLimiter *ratelimit.Bucket
	waitingList *goque.Queue
	network     *networkCtrl.Controller

	// Internal Flags
	distributorRunning bool

	// Cancelled request
	cancelLock       sync.Mutex
	cancelledRequest map[int64]bool
}

// New
func New(network *networkCtrl.Controller, dataDir string) (*Controller, error) {
	ctrl := new(Controller)
	ctrl.rateLimiter = ratelimit.NewBucket(time.Second, 20)
	if dataDir == "" {
		return nil, domain.ErrQueuePathIsNotSet
	}
	if q, err := goque.OpenQueue(dataDir); err != nil {
		return nil, err
	} else {
		ctrl.waitingList = q
	}

	ctrl.cancelledRequest = make(map[int64]bool)
	ctrl.network = network
	return ctrl, nil
}

// distributor
// Pulls the next request from the waitingList and pass it to the executor. It uses
// a rate limiter to throttle the throughput
func (ctrl *Controller) distributor() {
	// double check
	if ctrl.isDistributorRunning() {
		return
	}

	ctrl.setDistributorState(true)
	defer ctrl.setDistributorState(false)

	for {
		// Wait until network is available
		ctrl.network.WaitForNetwork()

		logs.Debug("QueueController:: distributor started",
			zap.Uint64("Length", ctrl.waitingList.Length()),
		)

		if ctrl.waitingList.Length() == 0 {
			break
		}

		// Peek item from the queue
		item, err := ctrl.waitingList.Dequeue()
		if err != nil {
			logs.Warn("QueueController:: distributor()->Dequeue()", zap.Error(err))
			return
		}

		// Prepare
		req := request{}
		if err := req.UnmarshalJSON(item.Value); err != nil {
			logs.Error("QueueController:: distributor()->UnmarshalJSON()", zap.Error(err))
			return
		}

		if !ctrl.IsRequestCancelled(int64(req.ID)) {
			logs.Debug("QueueController:: Request peeked from waiting list",
				zap.Uint64("RequestID", req.ID),
				zap.String("RequestName", msg.ConstructorNames[req.MessageEnvelope.Constructor]),
			)
			go ctrl.executor(req)
		} else {
			logs.Debug("QueueController:: Request cancelled",
				zap.Uint64("RequestID", req.ID),
				zap.String("RequestName", msg.ConstructorNames[req.MessageEnvelope.Constructor]),
			)
		}
	}

}

// setDistributorState
func (ctrl *Controller) setDistributorState(b bool) {
	ctrl.distributorLock.Lock()
	ctrl.distributorRunning = b
	ctrl.distributorLock.Unlock()
	return
}

// isDistributorRunning
func (ctrl *Controller) isDistributorRunning() bool {
	ctrl.distributorLock.Lock()
	b := ctrl.distributorRunning
	ctrl.distributorLock.Unlock()
	return b
}

// addToWaitingList
func (ctrl *Controller) addToWaitingList(req *request) {
	jsonRequest, err := req.MarshalJSON()
	if err != nil {
		logs.Error("addToWaitingList()->MarshalJSON()", zap.Error(err))
		return
	}
	if _, err := ctrl.waitingList.Enqueue(jsonRequest); err != nil {
		logs.Error("addToWaitingList()->Enqueue()", zap.Error(err))
		return
	}
	logs.Debug("QueueController:: Request added to waiting list",
		zap.String("Constructor", msg.ConstructorNames[req.MessageEnvelope.Constructor]),
		zap.Uint64("RequestID", req.MessageEnvelope.RequestID),
	)
	if !ctrl.isDistributorRunning() {
		go ctrl.distributor()
	}
}

// reinitializePendingMessages load queue items from storage
// FIXME:: this function must be removed
func (ctrl *Controller) reinitializePendingMessages() {
	logs.Info("reinitializePendingMessages()")
	// Remove all MessageSend requests from queue and add all pending messages back to queue
	items := make([]*goque.Item, 0)
	for {
		item, err := ctrl.waitingList.Dequeue()
		if err != nil || item == nil {
			break
		}
		tmp := new(msg.MessageEnvelope)
		_ = tmp.Unmarshal(item.Value)
		if tmp.Constructor != msg.C_MessagesSend {
			items = append(items, item)
		}
	}

	// get all pendingMessages
	pendingMessages := repo.PendingMessages.GetAllPendingMessages()

	// add pendingMessages to queue
	for _, v := range pendingMessages {
		messageEnvelope := new(msg.MessageEnvelope)
		messageEnvelope.RequestID = uint64(v.RandomID)
		// v.RandomID = domain.SequentialUniqueID()
		messageEnvelope.Constructor = msg.C_MessagesSend
		messageEnvelope.Message, _ = v.Marshal()
		req := &request{
			ID:              messageEnvelope.RequestID,
			Timeout:         domain.WebsocketRequestTime,
			MessageEnvelope: messageEnvelope,
		}

		// add its callback here
		ctrl.addToWaitingList(req)
	}

	// add items to queue
	for _, v := range items {
		_, _ = ctrl.waitingList.Enqueue(v.Value)
	}

	logs.Info("reinitializePendingMessages() Finished",
		zap.Uint64("MessageQueue Length", ctrl.waitingList.Length()),
	)
}

// executor
// Sends the message to the networkController and waits for the response. If time is up then it call the
// TimeoutCallback otherwise if response arrived in time, SuccessCallback will be called.
func (ctrl *Controller) executor(req request) {
	reqCallbacks := domain.GetRequestCallback(req.ID)
	if reqCallbacks == nil {
		logs.Debug("QueueController:: Callback not found",
			zap.Uint64("RequestID", req.ID),
		)

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
	logs.Debug("QueueController:: Request handover to network controller",
		zap.Uint64("RequestID", req.ID),
	)

	// Try to send it over wire, if error happened put it back into the queue
	if err := ctrl.network.Send(req.MessageEnvelope, false); err != nil {
		logs.Error("executor() -> network.Send()", zap.Error(err))
		ctrl.addToWaitingList(&req)
		return
	}

	select {
	case <-time.After(req.Timeout):
		domain.RemoveRequestCallback(req.ID)
		if reqCallbacks.TimeoutCallback != nil {
			if reqCallbacks.IsUICallback {
				uiexec.Ctx().Exec(func() { reqCallbacks.TimeoutCallback() })
			} else {
				reqCallbacks.TimeoutCallback()
			}
		}

		// hotfix check pendingMessage &&  messagesReadHistory on timeout
		if req.MessageEnvelope.Constructor == msg.C_MessagesSend {
			pmsg, err := repo.PendingMessages.GetPendingMessageByRequestID(int64(req.ID))
			if err == nil && pmsg != nil {
				logs.Warn("executor() :: NOT SENT and request added to queue again !!",
					zap.String("ConstructorName", msg.ConstructorNames[req.MessageEnvelope.Constructor]),
					zap.Uint64("RequestID", req.ID),
				)
				ctrl.addToWaitingList(&req)
			}
		} else if req.MessageEnvelope.Constructor == msg.C_MessagesReadHistory {

			logs.Warn("executor() :: NOT SENT and request added to queue again !!",
				zap.String("ConstructorName", msg.ConstructorNames[req.MessageEnvelope.Constructor]),
				zap.Uint64("RequestID", req.ID),
			)
			ctrl.addToWaitingList(&req)
		}

	case res := <-reqCallbacks.ResponseChannel:
		logs.Debug("QueueController:: ResponseChannel received signal",
			zap.String("ConstructorName", msg.ConstructorNames[res.Constructor]),
			zap.Uint64("RequestID", res.RequestID),
		)
		if reqCallbacks.SuccessCallback != nil {
			if reqCallbacks.IsUICallback {
				uiexec.Ctx().Exec(func() { reqCallbacks.SuccessCallback(res) })
			} else {
				reqCallbacks.SuccessCallback(res)
			}
		} else {
			logs.Warn("QueueController:: ResponseChannel received signal SuccessCallback is null",
				zap.String("ConstructorName", msg.ConstructorNames[res.Constructor]),
				zap.Uint64("RequestID", res.RequestID),
			)
		}
	}
	return
}

// ExecuteRealtimeCommand run request immediately and do not save it in queue
func (ctrl *Controller) ExecuteRealtimeCommand(requestID uint64, constructor int64, commandBytes []byte, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler, blockingMode, isUICallback bool) (err error) {
	messageEnvelope := new(msg.MessageEnvelope)
	messageEnvelope.Constructor = constructor
	messageEnvelope.RequestID = requestID
	messageEnvelope.Message = commandBytes

	// Add the callback functions
	domain.AddRequestCallback(requestID, successCB, domain.WebsocketDirectTime, timeoutCB, isUICallback)

	execBlock := func(reqID uint64, req *msg.MessageEnvelope) error {
		err := ctrl.network.Send(req, blockingMode)
		if err != nil {
			logs.Error("ExecuteRealtimeCommand()->network.Send()",
				zap.String("Error", err.Error()),
				zap.String("ConstructorName", msg.ConstructorNames[req.Constructor]),
				zap.Uint64("RequestID", requestID),
			)
			return err
		}

		reqCB := domain.GetRequestCallback(reqID)
		if reqCB != nil {
			select {
			case <-time.After(domain.WebsocketDirectTime):
				logs.Debug("QueueController::ExecuteRealtimeCommand()->execBlock() : Timeout",
					zap.String("ConstructorName", msg.ConstructorNames[req.Constructor]),
					zap.Uint64("RequestID", requestID),
				)
				domain.RemoveRequestCallback(reqID)
				if reqCB.TimeoutCallback != nil {
					if reqCB.IsUICallback {
						uiexec.Ctx().Exec(func() { reqCB.TimeoutCallback() })
					} else {
						reqCB.TimeoutCallback()
					}
				}
				return domain.ErrRequestTimeout
			case res := <-reqCB.ResponseChannel:
				logs.Debug("QueueController::ExecuteRealtimeCommand()->execBlock()  : Success",
					zap.String("ConstructorName", msg.ConstructorNames[req.Constructor]),
					zap.Uint64("RequestID", requestID),
				)
				if reqCB.SuccessCallback != nil {
					if reqCB.IsUICallback {
						uiexec.Ctx().Exec(func() { reqCB.SuccessCallback(res) })
					} else {
						reqCB.SuccessCallback(res)
					}
				}
			}
		} else {
			logs.Debug("QueueController::ExecuteRealtimeCommand() RequestCallback not found",
				zap.String("Constructor", msg.ConstructorNames[req.Constructor]),
				zap.Uint64("RequestID", requestID),
			)
		}
		return nil
	}

	if blockingMode {
		err = execBlock(requestID, messageEnvelope)
	} else {
		go execBlock(requestID, messageEnvelope)
	}

	return err
}

// ExecuteCommand put request in queue and distributor will execute it later
func (ctrl *Controller) ExecuteCommand(requestID uint64, constructor int64, requestBytes []byte, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler, isUICallback bool) {
	logs.Debug("QueueController::",
		zap.String("Constructor", msg.ConstructorNames[constructor]),
		zap.Uint64("RequestID", requestID),
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
	logs.Info("QueueController:: Start")

	// ctrl.reinitializePendingMessages()

	go ctrl.distributor()
}

// Stop queue
func (ctrl *Controller) Stop() {
	logs.Debug("StopServices-SyncController::Stop() called")
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
func (ctrl *Controller) DropQueue() (dataDir string, err error) {
	dataDir = ctrl.waitingList.DataDir
	ctrl.waitingList.Drop()
	return
}

// OpenQueue init queue files in storage
func (ctrl *Controller) OpenQueue(dataDir string) (err error) {
	ctrl.waitingList, err = goque.OpenQueue(dataDir)
	return
}
