package request

import (
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/internal/uiexec"
	"github.com/ronaksoft/rony"
	"github.com/ronaksoft/rony/tools"
	"google.golang.org/protobuf/proto"
	"sync"
	"time"
)

/*
   Creation Time: 2021 - May - 25
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2020
*/

type Callback interface {
	Constructor() int64
	Flags() DelegateFlag
	OnComplete(m *rony.MessageEnvelope)
	OnProgress(percent int64)
	OnTimeout()
	RequestID() uint64
	Envelope() *rony.MessageEnvelope
	TeamAccess() uint64
	TeamID() int64
	Timeout() time.Duration
	UI() bool
}

type callback struct {
	envelope   *rony.MessageEnvelope
	onComplete func(m *rony.MessageEnvelope)
	onTimeout  func()
	onProgress func(percent int64)
	ui         bool
	createdOn  int64
	flags      DelegateFlag
	timeout    time.Duration

	ResponseChannel chan *rony.MessageEnvelope
	DepartureTime   int64
}

func (c *callback) Flags() DelegateFlag {
	return c.flags
}

func (c *callback) TeamID() int64 {
	return domain.GetTeamID(c.envelope)
}

func (c *callback) TeamAccess() uint64 {
	return domain.GetTeamAccess(c.envelope)
}

func (c *callback) RequestID() uint64 {
	return c.envelope.RequestID
}

func (c *callback) Constructor() int64 {
	return c.envelope.Constructor
}

func (c *callback) OnComplete(m *rony.MessageEnvelope) {
	if c.onComplete == nil {
		return
	}
	if c.ui {
		uiexec.ExecSuccessCB(c.onComplete, m)
	} else {
		c.onComplete(m)
	}
}

func (c *callback) OnTimeout() {
	if c.onTimeout == nil {
		return
	}
	if c.ui {
		uiexec.ExecTimeoutCB(c.onTimeout)
	} else {
		c.onTimeout()
	}

}

func (c *callback) OnProgress(percent int64) {
	if c.onProgress == nil {
		return
	}
	c.onProgress(percent)
}

func (c *callback) UI() bool {
	return c.ui
}

func (c *callback) Timeout() time.Duration {
	if c.timeout == 0 {
		return domain.WebsocketRequestTimeout
	}
	return c.timeout
}

func (c *callback) Envelope() *rony.MessageEnvelope {
	return c.envelope
}

func NewCallback(
	teamID int64, teamAccess uint64,
	reqID uint64, constructor int64, req proto.Message,
	onTimeout domain.TimeoutCallback, onComplete domain.MessageHandler, onProgress func(int64),
	ui bool, flags DelegateFlag, timeout time.Duration,
) *callback {
	cb := &callback{
		envelope:   &rony.MessageEnvelope{},
		onComplete: onComplete,
		onTimeout:  onTimeout,
		onProgress: onProgress,
		ui:         ui,
		createdOn:  tools.NanoTime(),
		flags:      flags,
		timeout:    timeout,

		DepartureTime:   tools.NanoTime(),
		ResponseChannel: make(chan *rony.MessageEnvelope),
	}
	cb.envelope.Fill(reqID, constructor, req, domain.TeamHeader(teamID, teamAccess)...)
	return cb
}

func NewCallbackFromBytes(
	teamID int64, teamAccess uint64,
	reqID uint64, constructor int64, reqBytes []byte,
	onTimeout domain.TimeoutCallback, onComplete domain.MessageHandler, onProgress func(int64),
	ui bool, flags DelegateFlag, timeout time.Duration,
) *callback {
	cb := &callback{
		envelope: &rony.MessageEnvelope{
			RequestID:   reqID,
			Constructor: constructor,
		},
		onComplete:      onComplete,
		onTimeout:       onTimeout,
		onProgress:      onProgress,
		ui:              ui,
		createdOn:       tools.NanoTime(),
		flags:           flags,
		timeout:         timeout,
		DepartureTime:   tools.NanoTime(),
		ResponseChannel: make(chan *rony.MessageEnvelope),
	}
	cb.envelope.Message = append(cb.envelope.Message, reqBytes...)
	cb.envelope.Header = domain.TeamHeader(teamID, teamAccess)
	return cb
}

func EmptyCallback() *callback {
	return &callback{
		envelope: &rony.MessageEnvelope{},
	}
}

var (
	callbacksMtx     sync.Mutex
	requestCallbacks map[uint64]*callback
)

func init() {
	requestCallbacks = make(map[uint64]*callback, 100)
}

func RegisterCallback(
	reqID uint64, constructor int64, completeCB domain.MessageHandler, timeOut time.Duration, timeoutCB domain.TimeoutCallback, isUICallback bool,
) *callback {
	cb := &callback{
		envelope: &rony.MessageEnvelope{
			Constructor: constructor,
			RequestID:   reqID,
		},
		onComplete:      completeCB,
		onTimeout:       timeoutCB,
		ui:              isUICallback,
		ResponseChannel: make(chan *rony.MessageEnvelope, 1),
		createdOn:       tools.NanoTime(),
		DepartureTime:   tools.NanoTime(),
		timeout:         timeOut,
	}
	callbacksMtx.Lock()
	requestCallbacks[reqID] = cb
	callbacksMtx.Unlock()
	return cb
}

func UnregisterCallback(reqID uint64) {
	callbacksMtx.Lock()
	delete(requestCallbacks, reqID)
	callbacksMtx.Unlock()
}

func GetCallback(reqID uint64) (cb *callback) {
	callbacksMtx.Lock()
	cb = requestCallbacks[reqID]
	callbacksMtx.Unlock()
	return
}
