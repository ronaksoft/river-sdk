package executer

import (
	"sync"
	"time"

	"git.ronaksoftware.com/ronak/riversdk/loadtester/shared"

	"git.ronaksoftware.com/ronak/riversdk/msg"
)

// Executer command executer
type Executer struct {
	netCtrl      shared.Networker
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

// NewExecuter create new instance of command executer
func NewExecuter(netCtrl shared.Networker) *Executer {
	exe := &Executer{
		netCtrl:  netCtrl,
		requests: make(map[uint64]*Request),
	}
	return exe
}

// Exec execute command
func (exec *Executer) Exec(message *msg.MessageEnvelope, onSuccess shared.SuccessCallback, onTimeOut shared.TimeoutCallback, timeout time.Duration) {

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
	go exec.executer(message, r)
}

// executer send and wait for response, call this on go routine
func (exec *Executer) executer(message *msg.MessageEnvelope, r *Request) {
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
func (exec *Executer) GetRequest(reqID uint64) *Request {
	exec.requestsLock.Lock()
	req, ok := exec.requests[reqID]
	exec.requestsLock.Unlock()
	if ok {
		return req
	}
	return nil
}

// RemoveRequest removes response callbacks
func (exec *Executer) RemoveRequest(reqID uint64) {
	exec.requestsLock.Lock()
	delete(exec.requests, reqID)
	exec.requestsLock.Unlock()
}
