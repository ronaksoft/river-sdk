package queueCtrl

import (
	"git.ronaksoftware.com/river/msg/msg"
	"git.ronaksoftware.com/ronak/riversdk/internal/logs"
	fileCtrl "git.ronaksoftware.com/ronak/riversdk/pkg/ctrl_file"
	"git.ronaksoftware.com/ronak/riversdk/pkg/ctrl_network"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo"
	"git.ronaksoftware.com/ronak/riversdk/pkg/uiexec"
	"github.com/beeker1121/goque"
	"github.com/juju/ratelimit"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"sync"
	"time"
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
	networkCtrl *networkCtrl.Controller
	fileCtrl    *fileCtrl.Controller

	// Internal Flags
	distributorLock    sync.Mutex
	distributorRunning bool

	// Cancelled request
	cancelLock       sync.Mutex
	cancelledRequest map[int64]bool
}

// New
func New(fileCtrl *fileCtrl.Controller, network *networkCtrl.Controller, dataDir string) (*Controller, error) {
	ctrl := new(Controller)
	ctrl.dataDir = filepath.Join(dataDir, "queue")
	ctrl.rateLimiter = ratelimit.NewBucket(time.Second, 20)
	if dataDir == "" {
		return nil, domain.ErrQueuePathIsNotSet
	}

	ctrl.cancelledRequest = make(map[int64]bool)
	ctrl.networkCtrl = network
	ctrl.fileCtrl = fileCtrl
	return ctrl, nil
}

// distributor
// Pulls the next request from the waitingList and pass it to the executor. It uses
// a rate limiter to throttle the throughput
func (ctrl *Controller) distributor() {
	for {
		// Wait until network is available
		ctrl.networkCtrl.WaitForNetwork(true)

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

		if !ctrl.IsRequestCancelled(int64(req.ID)) {
			go ctrl.executor(req)
		} else {
			logs.Info("QueueController discarded a canceled request",
				zap.Uint64("ReqID", req.ID),
				zap.String("C", msg.ConstructorNames[req.MessageEnvelope.Constructor]),
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
	reqCB := domain.GetRequestCallback(req.ID)
	if reqCB == nil {
		reqCB = domain.AddRequestCallback(
			req.ID, req.MessageEnvelope.Constructor, nil, domain.WebsocketRequestTime, nil, false,
		)
	}
	reqCB.DepartureTime = time.Now()

	// Try to send it over wire, if error happened put it back into the queue
	if err := ctrl.networkCtrl.SendWebsocket(req.MessageEnvelope, false); err != nil {
		logs.Error("QueueCtrl got error from NetCtrl", zap.Error(err))
		logs.Info("QueueCtrl re-push the request into the queue")
		ctrl.addToWaitingList(&req)
		return
	}

	select {
	case <-time.After(req.Timeout):
		reqCB := domain.GetRequestCallback(req.ID)
		if reqCB == nil {
			return
		}
		switch req.MessageEnvelope.Constructor {
		case msg.C_MessagesSend, msg.C_MessagesSendMedia:

			pmsg, err := repo.PendingMessages.GetByRandomID(int64(req.ID))
			if err == nil && pmsg != nil {
				ctrl.addToWaitingList(&req)
				return
			}
		case msg.C_MessagesReadHistory, msg.C_MessagesGetHistory,
			msg.C_ContactsImport, msg.C_ContactsGet,
			msg.C_AuthSendCode, msg.C_AuthRegister, msg.C_AuthLogin,
			msg.C_LabelsAddToMessage, msg.C_LabelsRemoveFromMessage:
			ctrl.addToWaitingList(&req)
			return
		default:
			if reqCB.TimeoutCallback != nil {
				if reqCB.IsUICallback {
					uiexec.ExecTimeoutCB(reqCB.TimeoutCallback)
				} else {
					reqCB.TimeoutCallback()
				}
			}
		}
	case res := <-reqCB.ResponseChannel:
		switch req.MessageEnvelope.Constructor {
		case msg.C_MessagesSend, msg.C_MessagesSendMedia:
			switch res.Constructor {
			case msg.C_Error:
				errMsg := &msg.Error{}
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
		if reqCB.SuccessCallback != nil {
			if reqCB.IsUICallback {
				uiexec.ExecSuccessCB(reqCB.SuccessCallback, res)
			} else {
				reqCB.SuccessCallback(res)
			}
		} else {
			logs.Warn("QueueCtrl received response but no callback exists!!!",
				zap.String("C", msg.ConstructorNames[res.Constructor]),
				zap.Uint64("ReqID", res.RequestID),
			)
		}
	}
	domain.RemoveRequestCallback(req.ID)
	return
}

// RealtimeCommand run request immediately and do not save it in queue
func (ctrl *Controller) RealtimeCommand(
	messageEnvelope *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler,
	blockingMode, isUICallback bool,
) {
	logs.Debug("QueueCtrl fires realtime command",
		zap.Uint64("ReqID", messageEnvelope.RequestID),
		zap.String("C", msg.ConstructorNames[messageEnvelope.Constructor]),
	)

	// Add the callback functions
	reqCB := domain.AddRequestCallback(
		messageEnvelope.RequestID, messageEnvelope.Constructor, successCB, domain.WebsocketRequestTime, timeoutCB, isUICallback,
	)
	execBlock := func(reqID uint64, req *msg.MessageEnvelope) {
		err := ctrl.networkCtrl.SendWebsocket(req, blockingMode)
		if err != nil {
			logs.Warn("QueueCtrl got error from NetCtrl",
				zap.String("Error", err.Error()),
				zap.String("C", msg.ConstructorNames[req.Constructor]),
				zap.Uint64("ReqID", req.RequestID),
			)
			if timeoutCB != nil {
				timeoutCB()
			}
			return
		}

		select {
		case <-time.After(reqCB.Timeout):
			logs.Debug("QueueCtrl got timeout on realtime command",
				zap.String("C", msg.ConstructorNames[req.Constructor]),
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
			logs.Debug("QueueCtrl got response on realtime command",
				zap.Uint64("ReqID", req.RequestID),
				zap.String("ReqC", msg.ConstructorNames[req.Constructor]),
				zap.String("ResC", msg.ConstructorNames[res.Constructor]),
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

// EnqueueCommand put request in queue and distributor will execute it later
func (ctrl *Controller) EnqueueCommand(
	messageEnvelope *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler,
	isUICallback bool,
) {
	logs.Debug("QueueCtrl enqueues command",
		zap.Uint64("ReqID", messageEnvelope.RequestID),
		zap.String("C", msg.ConstructorNames[messageEnvelope.Constructor]),
	)

	// Add the callback functions
	_ = domain.AddRequestCallback(
		messageEnvelope.RequestID, messageEnvelope.Constructor, successCB, domain.WebsocketRequestTime, timeoutCB, isUICallback,
	)

	// Add the request to the queue
	ctrl.addToWaitingList(&request{
		ID:              messageEnvelope.RequestID,
		Timeout:         domain.WebsocketRequestTime,
		MessageEnvelope: messageEnvelope,
	})
}

// Start queue
func (ctrl *Controller) Start(resetQueue bool) {
	logs.Info("QueueCtrl started")
	if resetQueue {
		_ = os.RemoveAll(ctrl.dataDir)
	}
	err := ctrl.OpenQueue()
	if err != nil {
		logs.Fatal("We couldn't initialize the queue", zap.Error(err))
	}

	// Try to resend unsent messages
	for _, pmsg := range repo.PendingMessages.GetAll() {
		switch pmsg.MediaType {
		case msg.InputMediaTypeEmpty:
			if resetQueue {
				repo.PendingMessages.Delete(pmsg.ID)
			} else {
				// it will be MessagesSend
				req := repo.PendingMessages.ToMessagesSend(pmsg)
				reqBytes, _ := req.Marshal()
				ctrl.EnqueueCommand(&msg.MessageEnvelope{
					Constructor: msg.C_MessagesSend,
					RequestID:   uint64(req.RandomID),
					Message:     reqBytes,
				}, nil, nil, false)
			}

		default:
			if resetQueue {
				_ = repo.PendingMessages.Delete(pmsg.ID)
				uploadRequest, ok := ctrl.fileCtrl.GetUploadRequest(pmsg.FileID)
				if !ok {
					continue
				}
				ctrl.fileCtrl.CancelUploadRequest(uploadRequest.GetID())
			} else {
				// it will be MessagesSendMedia
				req := repo.PendingMessages.ToMessagesSendMedia(pmsg)
				if req == nil {
					continue
				}
				reqBytes, _ := req.Marshal()
				ctrl.EnqueueCommand(&msg.MessageEnvelope{
					Constructor: msg.C_MessagesSendMedia,
					RequestID:   uint64(req.RandomID),
					Message:     reqBytes,
				}, nil, nil, false)
			}

		}
	}

	go ctrl.distributor()
}

// Stop queue
func (ctrl *Controller) Stop() {
	logs.Info("QueueCtrl stopped")
	ctrl.DropQueue()

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
	err := domain.Try(10, time.Millisecond*100, func() error {
		return ctrl.waitingList.Drop()
	})
	if err != nil {
		logs.Warn("QueueCtrl got error on dropping queue")
	}
}

// OpenQueue init queue files in storage
func (ctrl *Controller) OpenQueue() (err error) {
	err = domain.Try(10, 100*time.Millisecond, func() error {
		if q, err := goque.OpenQueue(ctrl.dataDir); err != nil {
			err = os.RemoveAll(ctrl.dataDir)
			if err != nil {
				logs.Warn("QueueCtrl we got error on removing queue directory", zap.Error(err))
			}
			return err
		} else {
			ctrl.waitingList = q
		}
		return nil
	})
	return
}
