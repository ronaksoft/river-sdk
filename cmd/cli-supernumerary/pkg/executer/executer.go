package executer

import (
	"sync"
	"time"

	"git.ronaksoftware.com/ronak/riversdk/cmd/cli-supernumerary/pkg/shared"

	msg "git.ronaksoftware.com/ronak/riversdk/msg/chat"
)

// Executor command executor
type Executor struct {
	netCtrl      shared.Neter
	requestsLock sync.Mutex
	requests     map[uint64]*Request
}

// Request required data to send command to server
type Request struct {
	RequestID           uint64
	CreatedOn           time.Time
	Timeout             time.Duration
	OnSuccess           shared.SuccessCallback
	OnTimeOut           shared.TimeoutCallback
	ResponseWaitChannel chan *msg.MessageEnvelope
}

// NewExecutor create new instance of command executor
func NewExecutor(netCtrl shared.Neter) *Executor {
	exe := &Executor{
		netCtrl:  netCtrl,
		requests: make(map[uint64]*Request),
	}
	return exe
}

// Exec execute command
func (exec *Executor) Exec(message *msg.MessageEnvelope, onSuccess shared.SuccessCallback, onTimeOut shared.TimeoutCallback, timeout time.Duration) {
	if timeout == 0 {
		timeout = shared.DefaultTimeout
	}

	r := &Request{
		RequestID:           message.RequestID,
		OnSuccess:           onSuccess,
		OnTimeOut:           onTimeOut,
		ResponseWaitChannel: make(chan *msg.MessageEnvelope),
		CreatedOn:           time.Now(),
		Timeout:             timeout,
	}
	exec.requestsLock.Lock()
	exec.requests[r.RequestID] = r
	exec.requestsLock.Unlock()
	go exec.executor(message, r)
}

// executor send and wait for response, call this on go routine
func (exec *Executor) executor(message *msg.MessageEnvelope, r *Request) {
	if exec.netCtrl == nil {
		out := new(msg.MessageEnvelope)
		out.RequestID = message.RequestID
		msg.ResultError(out, &msg.Error{Code: "00", Items: "network controller is null"})
		r.OnSuccess(out, time.Since(r.CreatedOn))
		return
	}

	err := exec.netCtrl.Send(message)
	if err != nil {
		out := new(msg.MessageEnvelope)
		out.RequestID = message.RequestID
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		r.OnSuccess(out, time.Since(r.CreatedOn))
		return
	}

	select {
	case <-time.After(r.Timeout):
		if r.OnTimeOut != nil {
			r.OnTimeOut(r.RequestID, time.Since(r.CreatedOn))
		}
	case res := <-r.ResponseWaitChannel:
		if r.OnSuccess != nil {
			r.OnSuccess(res, time.Since(r.CreatedOn))
		}
	}
}

// GetRequest returns response callbacks
func (exec *Executor) GetRequest(reqID uint64) *Request {
	exec.requestsLock.Lock()
	req, ok := exec.requests[reqID]
	exec.requestsLock.Unlock()
	if ok {
		return req
	}
	return nil
}

// RemoveRequest removes response callbacks
func (exec *Executor) RemoveRequest(reqID uint64) {
	exec.requestsLock.Lock()
	delete(exec.requests, reqID)
	exec.requestsLock.Unlock()
}
