package executer

import (
	"sync"
	"time"

	"git.ronaksoftware.com/ronak/riversdk/loadtester/types"
	"git.ronaksoftware.com/ronak/riversdk/msg"
)

const (
	// DefaultTimeout request timeout
	DefaultTimeout = 10 * time.Second
)

// Executer command executer
type Executer struct {
	netCtrl      types.Network
	requestsLock sync.Mutex
	requests     map[uint64]*Request
}

// SuccessCallback timeout
type SuccessCallback func(response *msg.MessageEnvelope, elapsed time.Duration)

// TimeoutCallback function
type TimeoutCallback func(requestID uint64, elapsed time.Duration)

// Request required data to send command to server
type Request struct {
	RequestID           uint64
	CreatedOn           time.Time
	Timeout             time.Duration
	OnSuccess           SuccessCallback
	OnTimeOut           TimeoutCallback
	ResponseWaitChannel chan *msg.MessageEnvelope
}

// NewExecuter create new instance of command executer
func NewExecuter(netCtrl types.Network) *Executer {
	exe := &Executer{
		netCtrl:  netCtrl,
		requests: make(map[uint64]*Request),
	}
	return exe
}

// Exec execute command
func (exec *Executer) Exec(message *msg.MessageEnvelope, onSuccess SuccessCallback, onTimeOut TimeoutCallback, timeout time.Duration) {

	if timeout == 0 {
		timeout = DefaultTimeout
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
