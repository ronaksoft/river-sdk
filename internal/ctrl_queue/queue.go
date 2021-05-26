package queueCtrl

import (
	"git.ronaksoft.com/river/msg/go/msg"
	fileCtrl "git.ronaksoft.com/river/sdk/internal/ctrl_file"
	"git.ronaksoft.com/river/sdk/internal/ctrl_network"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/internal/logs"
	"git.ronaksoft.com/river/sdk/internal/repo"
	"git.ronaksoft.com/river/sdk/internal/request"
	"github.com/beeker1121/goque"
	"github.com/juju/ratelimit"
	"github.com/ronaksoft/rony"
	"github.com/ronaksoft/rony/registry"
	"github.com/ronaksoft/rony/tools"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	logger *logs.Logger
)

func init() {
	logger = logs.With("QueueCtrl")
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
	cancelledRequest map[uint64]bool
}

func New(fileCtrl *fileCtrl.Controller, network *networkCtrl.Controller, dataDir string) *Controller {
	ctrl := new(Controller)
	ctrl.dataDir = filepath.Join(dataDir, "queue")
	ctrl.rateLimiter = ratelimit.NewBucket(time.Second, 20)
	if dataDir == "" {
		panic(domain.ErrQueuePathIsNotSet)
	}

	ctrl.cancelledRequest = make(map[uint64]bool)
	ctrl.networkCtrl = network
	ctrl.fileCtrl = fileCtrl
	return ctrl
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
		reqCB, err := request.UnmarshalCallback(item.Value)
		if err != nil {
			logger.Error("could not unmarshal popped request", zap.Error(err))
			continue
		}

		// If request is already canceled ignore it
		if ctrl.IsRequestCancelled(reqCB.RequestID()) {
			logger.Info("discarded a canceled request",
				zap.Uint64("ReqID", reqCB.RequestID()),
				zap.String("C", registry.ConstructorName(reqCB.Constructor())),
			)
			continue
		}

		go ctrl.executor(reqCB)
	}
}

// addToWaitingList
func (ctrl *Controller) addToWaitingList(reqCB request.Callback) {
	jsonRequest, err := reqCB.Marshal()
	if err != nil {
		logger.Warn("couldn't marshal the request", zap.Error(err))
		return
	}
	if _, err := ctrl.waitingList.Enqueue(jsonRequest); err != nil {
		logger.Warn("couldn't enqueue the request", zap.Error(err))
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
func (ctrl *Controller) executor(reqCB request.Callback) {
	defer logger.RecoverPanic(
		"SyncCtrl::executor",
		domain.M{
			"OS":  domain.ClientOS,
			"Ver": domain.ClientVersion,
			"C":   reqCB.Constructor(),
		},
		nil,
	)

	// Try to send it over wire, if error happened put it back into the queue
	if err := ctrl.networkCtrl.WebsocketSend(reqCB.Envelope(), 0); err != nil {
		logger.Info("re-push the request into the queue", zap.Error(err))
		ctrl.addToWaitingList(reqCB)
		return
	}

	select {
	case <-time.After(reqCB.Timeout()):
		switch reqCB.Constructor() {
		case msg.C_MessagesSend, msg.C_MessagesSendMedia:
			pmsg, err := repo.PendingMessages.GetByRandomID(int64(reqCB.RequestID()))
			if err == nil && pmsg != nil {
				ctrl.addToWaitingList(reqCB)
				return
			}
		case msg.C_MessagesReadHistory, msg.C_MessagesGetHistory,
			msg.C_ContactsImport, msg.C_ContactsGet,
			msg.C_AuthSendCode, msg.C_AuthRegister, msg.C_AuthLogin,
			msg.C_LabelsAddToMessage, msg.C_LabelsRemoveFromMessage:
			ctrl.addToWaitingList(reqCB)
			return
		default:
			reqCB.OnTimeout()
		}
	case res := <-reqCB.ResponseChan():
		switch reqCB.Constructor() {
		case msg.C_MessagesSend, msg.C_MessagesSendMedia, msg.C_MessagesForward:
			switch res.Constructor {
			case rony.C_Error:
				errMsg := &rony.Error{}
				_ = errMsg.Unmarshal(res.Message)
				if errMsg.Code == msg.ErrCodeAlreadyExists && errMsg.Items == msg.ErrItemRandomID {
					pm, _ := repo.PendingMessages.GetByRandomID(int64(reqCB.RequestID()))
					if pm != nil {
						_ = repo.PendingMessages.Delete(pm.ID)
					}
				} else if errMsg.Code == msg.ErrCodeAccess && errMsg.Items == "NON_TEAM_MEMBER" {
					pm, _ := repo.PendingMessages.GetByRandomID(int64(reqCB.RequestID()))
					if pm != nil {
						_ = repo.PendingMessages.Delete(pm.ID)
					}
				}
			}
		default:
			switch res.Constructor {
			case rony.C_Error:
				errMsg := &rony.Error{}
				_ = errMsg.Unmarshal(res.Message)
				if errMsg.Code == msg.ErrCodeInvalid && errMsg.Items == msg.ErrItemSalt {
					ctrl.addToWaitingList(reqCB)
					return
				}
			}
		}
		reqCB.OnComplete(res)
	}
	return
}

// EnqueueCommand put request in queue and distributor will execute it later
func (ctrl *Controller) EnqueueCommand(reqCB request.Callback) {
	defer logger.RecoverPanic(
		"SyncCtrl::EnqueueCommandWithTimeout",
		domain.M{
			"OS":  domain.ClientOS,
			"Ver": domain.ClientVersion,
			"C":   reqCB.Constructor(),
		},
		nil,
	)

	logger.Debug("enqueues command",
		zap.Uint64("ReqID", reqCB.RequestID()),
		zap.String("C", registry.ConstructorName(reqCB.Constructor())),
	)

	// Add the request to the queue
	ctrl.addToWaitingList(reqCB)
}

// Start queue
func (ctrl *Controller) Start(resetQueue bool) {
	logger.Info("started")
	if resetQueue {
		_ = os.RemoveAll(ctrl.dataDir)
	}
	err := ctrl.OpenQueue()
	if err != nil {
		logger.Fatal("couldn't initialize the queue", zap.Error(err))
	}

	// Try to resend unsent messages
	for _, pmsg := range repo.PendingMessages.GetAll() {
		if resetQueue {
			_ = repo.PendingMessages.Delete(pmsg.ID)
			continue
		}
		switch pmsg.MediaType {
		case msg.InputMediaType_InputMediaTypeEmpty:
			logger.Info("loads pending messages",
				zap.Int64("ID", pmsg.ID),
				zap.Int64("FileID", pmsg.FileID),
			)
			// it will be MessagesSend
			req := repo.PendingMessages.ToMessagesSend(pmsg)
			ctrl.EnqueueCommand(
				request.NewCallback(
					pmsg.TeamID, pmsg.TeamAccessHash, uint64(req.RandomID), msg.C_MessagesSend, req,
					nil, nil, nil, false, 0, 0,
				),
			)

		default:
			req := &msg.ClientSendMessageMedia{}
			_ = req.Unmarshal(pmsg.Media)
			switch req.MediaType {
			case msg.InputMediaType_InputMediaTypeUploadedDocument:
				checkSha256 := true
				for _, attr := range req.Attributes {
					if attr.Type == msg.DocumentAttributeType_AttributeTypeAudio {
						x := &msg.DocumentAttributeAudio{}
						_ = x.Unmarshal(attr.Data)
						if x.Voice {
							checkSha256 = false
						}
					}
				}
				ctrl.fileCtrl.UploadMessageDocument(pmsg.ID, req.FilePath, req.ThumbFilePath, req.FileID, req.ThumbID, pmsg.Sha256, pmsg.PeerID, checkSha256)
			default:
				// it will be MessagesSendMedia
				req := repo.PendingMessages.ToMessagesSendMedia(pmsg)
				if req == nil {
					continue
				}
				ctrl.EnqueueCommand(
					request.NewCallback(
						pmsg.TeamID, pmsg.TeamAccessHash, uint64(req.RandomID), msg.C_MessagesSendMedia, req,
						nil, nil, nil, false, 0, 0,
					),
				)
			}
		}
	}

	go ctrl.distributor()
}

// Stop queue
func (ctrl *Controller) Stop() {
	logger.Info("stopped")
	ctrl.DropQueue()

}

// IsRequestCancelled handle canceled request to do not process its response
func (ctrl *Controller) IsRequestCancelled(reqID uint64) bool {
	_, ok := ctrl.cancelledRequest[reqID]
	if ok {
		ctrl.cancelLock.Lock()
		delete(ctrl.cancelledRequest, reqID)
		ctrl.cancelLock.Unlock()
	}
	return ok
}

// CancelRequest cancel request
func (ctrl *Controller) CancelRequest(reqID uint64) {
	ctrl.cancelLock.Lock()
	ctrl.cancelledRequest[reqID] = true
	ctrl.cancelLock.Unlock()
}

// DropQueue remove queue from storage
func (ctrl *Controller) DropQueue() {
	err := tools.Try(10, time.Millisecond*100, func() error {
		return ctrl.waitingList.Drop()
	})
	if err != nil {
		logger.Warn("got error on dropping queue")
	}
}

// OpenQueue init queue files in storage
func (ctrl *Controller) OpenQueue() (err error) {
	err = tools.Try(10, 100*time.Millisecond, func() error {
		if q, err := goque.OpenQueue(ctrl.dataDir); err != nil {
			err = os.RemoveAll(ctrl.dataDir)
			if err != nil {
				logger.Warn("got error on removing queue directory", zap.Error(err))
			}
			return err
		} else {
			ctrl.waitingList = q
		}
		return nil
	})
	return
}
