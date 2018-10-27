package queue

import (
	"sync"
	"time"

	"git.ronaksoftware.com/ronak/riversdk/domain"
	"git.ronaksoftware.com/ronak/riversdk/log"
	"git.ronaksoftware.com/ronak/riversdk/msg"
	"git.ronaksoftware.com/ronak/riversdk/network"
	"git.ronaksoftware.com/ronak/riversdk/repo"
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

// // TODO : implement interface for QueueController
// type QueueController interface {

// }

// QueueController
// This controller will be connected to networkController and messages will be queued here
// before passing to the networkController.
type QueueController struct {
	//sync.Mutex
	distributorLock sync.Mutex

	rateLimiter            *ratelimit.Bucket
	waitingList            *goque.Queue
	network                *network.NetworkController
	deferredRequestHandler func(requestID int64, b []byte)

	// Internal Flags
	distributorRunning bool

	//Cancelled request
	cancellLock      sync.Mutex
	cancelledRequest map[int64]bool
}

// NewQueueController
func NewQueueController(network *network.NetworkController, dataDir string, deferredRequestHandler domain.DeferredRequestHandler) (*QueueController, error) {
	ctrl := new(QueueController)
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
	ctrl.deferredRequestHandler = deferredRequestHandler
	return ctrl, nil
}

// distributor
// Pulls the next request from the waitingList and pass it to the executor. It uses
// a rate limiter to throttle the throughput
func (ctrl *QueueController) distributor() {

	// double check
	if ctrl.isDistributorRunning() {
		return
	}

	ctrl.setDistributorState(true)
	defer ctrl.setDistributorState(false)

	for {

		// Wait While Network is Disconnected or Connecting
		for ctrl.network.Quality() == domain.DISCONNECTED || ctrl.network.Quality() == domain.CONNECTING {
			log.LOG.Debug("QueueController::distributor() Network is not connected ...",
				zap.String("Quality", domain.NetworkStatusName[ctrl.network.Quality()]),
			)
			time.Sleep(time.Second)
		}

		log.LOG.Debug("QueueController::distributor()",
			zap.Uint64("Queue Length", ctrl.waitingList.Length()),
		)

		if ctrl.waitingList.Length() == 0 {
			break
		}
		// Peek item from the queue
		item, err := ctrl.waitingList.Dequeue()
		if err != nil {
			log.LOG.Debug("QueueController::distributor()->Dequeue()",
				zap.String("Error", err.Error()),
			)
			return
		}

		// Disabled ratelimiter
		// // Check the rate limit
		// ctrl.rateLimiter.Wait(1)

		// Prepare
		req := request{}
		if err := req.UnmarshalJSON(item.Value); err != nil {
			log.LOG.Debug("QueueController::distributor()->UnmarshalJSON()",
				zap.String("Error", err.Error()),
			)
			return
		}

		if !ctrl.IsRequestCancelled(int64(req.ID)) {
			log.LOG.Debug("QueueController::distributor() Request peeked from waiting list",
				zap.Uint64("RequestID", req.ID),
				zap.String("RequestName", msg.ConstructorNames[req.MessageEnvelope.Constructor]),
			)
			go ctrl.executor(req)
		} else {
			log.LOG.Debug("QueueController::distributor() Request cancelled",
				zap.Uint64("RequestID", req.ID),
				zap.String("RequestName", msg.ConstructorNames[req.MessageEnvelope.Constructor]),
			)
		}
	}

}

// setDistributorState
func (ctrl *QueueController) setDistributorState(b bool) bool {
	changed := false
	ctrl.distributorLock.Lock()
	changed = ctrl.distributorRunning != b
	ctrl.distributorRunning = b
	ctrl.distributorLock.Unlock()

	return changed
}

// isDistributorRunning
func (ctrl *QueueController) isDistributorRunning() bool {
	ctrl.distributorLock.Lock()
	b := ctrl.distributorRunning
	ctrl.distributorLock.Unlock()
	return b
}

// executor
// Sends the message to the networkController and waits for the response. If time is up then it call the
// TimeoutCallback otherwise if response arrived in time, SuccessCallback will be called.
func (ctrl *QueueController) executor(req request) {
	reqCallbacks := domain.GetRequestCallback(req.ID)
	if reqCallbacks == nil {
		log.LOG.Debug("QueueController::executor() Callback not found",
			zap.Uint64("RequestID", req.ID),
		)

		reqCallbacks = domain.AddRequestCallback(
			req.ID,
			func(m *msg.MessageEnvelope) {
				b, _ := m.Marshal()
				if ctrl.deferredRequestHandler != nil {
					ctrl.deferredRequestHandler(int64(req.ID), b)
				}
			},
			req.Timeout,
			nil,
		)
	}
	if req.Timeout == 0 {
		req.Timeout = domain.DEFAULT_REQUEST_TIMEOUT
	}
	log.LOG.Debug("QueueController::executor() Request handover to network controller",
		zap.Uint64("RequestID", req.ID),
	)

	// Try to send it over wire, if error happened put it back into the queue
	if err := ctrl.network.Send(req.MessageEnvelope); err != nil {
		ctrl.addToWaitingList(&req)
		return
	}

	select {
	case <-time.After(req.Timeout):
		domain.RemoveRequestCallback(req.ID)
		if reqCallbacks.TimeoutCallback != nil {
			reqCallbacks.TimeoutCallback()
		}

		// hotfix check pendingMessage &&  messagesReadHistory on timeout
		if req.MessageEnvelope.Constructor == msg.C_MessagesSend {
			pmsg, err := repo.Ctx().PendingMessages.GetPendingMessageByRequestID(int64(req.ID))
			if err == nil && pmsg != nil {
				log.LOG.Warn("QueueController::executor() :: NOT SENT and request added to queue again !!",
					zap.String("ConstructorName", msg.ConstructorNames[req.MessageEnvelope.Constructor]),
					zap.Uint64("RequestID", req.ID),
				)
				ctrl.addToWaitingList(&req)
			}
		} else if req.MessageEnvelope.Constructor == msg.C_MessagesReadHistory {

			log.LOG.Warn("QueueController::executor() :: NOT SENT and request added to queue again !!",
				zap.String("ConstructorName", msg.ConstructorNames[req.MessageEnvelope.Constructor]),
				zap.Uint64("RequestID", req.ID),
			)
			ctrl.addToWaitingList(&req)
		}

	case res := <-reqCallbacks.ResponseChannel:
		log.LOG.Warn("QueueController::executor() :: ResponseChannel received signal",
			zap.String("ConstructorName", msg.ConstructorNames[res.Constructor]),
			zap.Uint64("RequestID", res.RequestID),
		)
		if reqCallbacks.SuccessCallback != nil {
			reqCallbacks.SuccessCallback(res)
		} else {
			log.LOG.Warn("QueueController::executor() :: ResponseChannel received signal SuccessCallback is null",
				zap.String("ConstructorName", msg.ConstructorNames[res.Constructor]),
				zap.Uint64("RequestID", res.RequestID),
			)
		}
	}
	return
}

// ExecuteRealtimeCommand
func (ctrl *QueueController) ExecuteRealtimeCommand(requestID uint64, constructor int64, commandBytes []byte, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler, blockingMode bool) (err error) {

	messageEnvelope := new(msg.MessageEnvelope)
	messageEnvelope.Constructor = constructor
	messageEnvelope.RequestID = requestID
	messageEnvelope.Message = commandBytes

	// Add the callback functions
	domain.AddRequestCallback(requestID, successCB, domain.DEFAULT_WS_REALTIME_TIMEOUT, timeoutCB)

	execBlock := func(reqID uint64, req *msg.MessageEnvelope) error {
		err := ctrl.network.Send(req)
		if err != nil {
			log.LOG.Debug("QueueController::ExecuteRealtimeCommand()->network.Send()",
				zap.String("Error", err.Error()),
				zap.String("ConstructorName", msg.ConstructorNames[req.Constructor]),
				zap.Uint64("RequestID", requestID),
			)
			return err
		}

		reqCB := domain.GetRequestCallback(reqID)
		if reqCB != nil {
			select {
			case <-time.After(domain.DEFAULT_WS_REALTIME_TIMEOUT):
				log.LOG.Debug("QueueController::ExecuteRealtimeCommand()->execBlock() : Timeout",
					zap.String("ConstructorName", msg.ConstructorNames[req.Constructor]),
					zap.Uint64("RequestID", requestID),
				)

				domain.RemoveRequestCallback(reqID)
				if reqCB.TimeoutCallback != nil {
					reqCB.TimeoutCallback()
				}
				err = domain.ErrRequestTimeout
			case res := <-reqCB.ResponseChannel:
				log.LOG.Debug("QueueController::ExecuteRealtimeCommand()->execBlock()  : Success",
					zap.String("ConstructorName", msg.ConstructorNames[req.Constructor]),
					zap.Uint64("RequestID", requestID),
				)
				if reqCB.SuccessCallback != nil {
					reqCB.SuccessCallback(res)
				}
			}
		} else {
			log.LOG.Debug("QueueController::ExecuteRealtimeCommand()->execBlock()  : RequestCallback not found",
				zap.String("ConstructorName", msg.ConstructorNames[req.Constructor]),
				zap.Uint64("RequestID", requestID),
			)
		}
		return err
	}

	if blockingMode {
		err = execBlock(requestID, messageEnvelope)
	} else {
		go execBlock(requestID, messageEnvelope)
	}

	return err
}

// executeRemoteCommand
func (ctrl *QueueController) ExecuteCommand(requestID uint64, constructor int64, requestBytes []byte, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	log.LOG.Debug("QueueController::ExecuteCommand()",
		zap.String("Constructor", msg.ConstructorNames[constructor]),
		zap.Uint64("RequestID", requestID),
	)
	messageEnvelope := new(msg.MessageEnvelope)
	messageEnvelope.RequestID = requestID
	messageEnvelope.Constructor = constructor
	messageEnvelope.Message = requestBytes
	req := request{
		ID:              requestID,
		Timeout:         domain.DEFAULT_REQUEST_TIMEOUT,
		MessageEnvelope: messageEnvelope,
	}

	// Add the callback functions
	domain.AddRequestCallback(requestID, successCB, req.Timeout, timeoutCB)

	// Add the request to the queue
	ctrl.addToWaitingList(&req)
}

// addToWaitingList
func (ctrl *QueueController) addToWaitingList(req *request) {
	jsonRequest, err := req.MarshalJSON()
	if err != nil {
		log.LOG.Debug("QueueController::addToWaitingList()->MarshalJSON()",
			zap.String("Error", err.Error()),
		)
		return
	}
	if _, err := ctrl.waitingList.Enqueue(jsonRequest); err != nil {
		log.LOG.Debug("QueueController::addToWaitingList()->Enqueue()",
			zap.String("Error", err.Error()),
		)
		return
	}
	log.LOG.Debug("QueueController::addToWaitingList() Request added to waiting list",
		zap.String("Constructor", msg.ConstructorNames[req.MessageEnvelope.Constructor]),
		zap.Uint64("RequestID", req.MessageEnvelope.RequestID),
	)
	if !ctrl.isDistributorRunning() {
		go ctrl.distributor()
	}
}

// Start
func (ctrl *QueueController) Start() {
	log.LOG.Info("QueueController::Start()")

	ctrl.reinitializePendingMessages()

	go ctrl.distributor()
}

func (ctrl *QueueController) reinitializePendingMessages() {
	log.LOG.Info("QueueController::reinitializePendingMessages()")
	// Remove all MessageSend requests from queue and add all pending messages back to queue
	items := make([]*goque.Item, 0)
	for {
		if item, err := ctrl.waitingList.Dequeue(); err == nil && item != nil {
			tmp := new(msg.MessageEnvelope)
			tmp.Unmarshal(item.Value)
			if tmp.Constructor != msg.C_MessagesSend {
				items = append(items, item)
			}
		} else {
			break
		}
	}

	// get all pendingMessages
	pendingMessages := repo.Ctx().PendingMessages.GetAllPendingMessages()

	//add pendingMessages to queue
	for _, v := range pendingMessages {
		messageEnvelope := new(msg.MessageEnvelope)
		messageEnvelope.RequestID = uint64(v.RandomID)
		v.RandomID = domain.RandomInt63()
		messageEnvelope.Constructor = msg.C_MessagesSend
		messageEnvelope.Message, _ = v.Marshal()
		req := &request{
			ID:              messageEnvelope.RequestID,
			Timeout:         domain.DEFAULT_REQUEST_TIMEOUT,
			MessageEnvelope: messageEnvelope,
		}

		// add its callback here

		ctrl.addToWaitingList(req)
	}

	//add items to queue
	for _, v := range items {
		ctrl.waitingList.Enqueue(v.Value)
	}

	log.LOG.Info("QueueController::reinitializePendingMessages() Finished",
		zap.Uint64("Queue Length", ctrl.waitingList.Length()),
	)
}

// Stop
func (ctrl *QueueController) Stop() {
	ctrl.waitingList.Close()
}

func (ctrl *QueueController) IsRequestCancelled(reqID int64) bool {
	_, ok := ctrl.cancelledRequest[reqID]
	if ok {
		ctrl.cancellLock.Lock()
		delete(ctrl.cancelledRequest, reqID)
		ctrl.cancellLock.Unlock()
	}
	return ok
}

func (ctrl *QueueController) CancelRequest(reqID int64) {
	ctrl.cancellLock.Lock()
	ctrl.cancelledRequest[reqID] = true
	ctrl.cancellLock.Unlock()
}

func (ctrl *QueueController) DropQueue() (dataDir string, err error) {
	dataDir = ctrl.waitingList.DataDir
	err = ctrl.waitingList.Drop()
	return
}

func (ctrl *QueueController) OpenQueue(dataDir string) (err error) {
	ctrl.waitingList, err = goque.OpenQueue(dataDir)
	return
}
