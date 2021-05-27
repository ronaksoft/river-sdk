package request

import (
	"encoding/json"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/internal/uiexec"
	"github.com/ronaksoft/rony"
	"github.com/ronaksoft/rony/errors"
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

type Unmarshaller interface {
	Unmarshal(data []byte) error
}

type Callback interface {
	Constructor() int64
	CreatedOn() int64
	Discard()
	Envelope() *rony.MessageEnvelope
	Flags() DelegateFlag
	Marshal() ([]byte, error)
	OnComplete(m *rony.MessageEnvelope)
	OnProgress(percent int64)
	OnTimeout()
	SetPreComplete(h domain.MessageHandler) Callback
	RequestData(req Unmarshaller) error
	RequestID() uint64
	ResponseChan() chan *rony.MessageEnvelope
	Response(constructor int64, proto proto.Message)
	SentOn() int64
	TeamAccess() uint64
	TeamID() int64
	Timeout() time.Duration
	UI() bool
}

// serialized
type serializedCallback struct {
	Timeout         time.Duration         `json:"timeout"`
	MessageEnvelope *rony.MessageEnvelope `json:"message_envelope"`
	SerializedOn    int64                 `json:"serialized_on"`
	CreatedOn       int64                 `json:"created_on"`
	UI              bool                  `json:"ui"`
	Flags           DelegateFlag          `json:"flags"`
}

type callback struct {
	envelope    *rony.MessageEnvelope
	preComplete domain.MessageHandler
	onComplete  domain.MessageHandler
	onTimeout   domain.TimeoutCallback
	onProgress  func(percent int64)
	ui          bool
	createdOn   int64
	sentOn      int64
	flags       DelegateFlag
	timeout     time.Duration
	resChan     chan *rony.MessageEnvelope
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
	unregister(c.envelope.RequestID)
	if c.preComplete != nil {
		c.preComplete(m)
	}
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
	unregister(c.envelope.RequestID)
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

func (c *callback) Marshal() ([]byte, error) {
	scb := &serializedCallback{
		Timeout:         c.timeout,
		MessageEnvelope: c.envelope.Clone(),
		SerializedOn:    tools.TimeUnix(),
		CreatedOn:       c.createdOn,
		UI:              c.ui,
		Flags:           c.flags,
	}
	return json.Marshal(scb)
}

func (c *callback) ResponseChan() chan *rony.MessageEnvelope {
	return c.resChan
}

func (c *callback) Discard() {
	unregister(c.envelope.RequestID)
}

func (c *callback) CreatedOn() int64 {
	return c.createdOn
}

func (c *callback) SentOn() int64 {
	return c.sentOn
}

func (c *callback) SetPreComplete(h domain.MessageHandler) Callback {
	c.preComplete = h
	return c
}

func (c *callback) RequestData(u Unmarshaller) error {
	err := u.Unmarshal(c.envelope.Message)
	if err != nil {
		me := &rony.MessageEnvelope{}
		me.Fill(c.envelope.RequestID, rony.C_Error, errors.New(errors.Internal, err.Error()), c.envelope.Header...)
		c.OnComplete(me)
	}
	return err
}

func (c *callback) Response(constructor int64, proto proto.Message) {
	res := &rony.MessageEnvelope{}
	res.Fill(c.envelope.RequestID, constructor, proto, c.envelope.Header...)
	c.OnComplete(res)
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
		sentOn:     tools.NanoTime(),
		flags:      flags,
		timeout:    timeout,
		resChan:    make(chan *rony.MessageEnvelope, 1),
	}
	cb.envelope.Fill(reqID, constructor, req, domain.TeamHeader(teamID, teamAccess)...)
	register(cb)
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
		onComplete: onComplete,
		onTimeout:  onTimeout,
		onProgress: onProgress,
		ui:         ui,
		createdOn:  tools.NanoTime(),
		sentOn:     tools.NanoTime(),
		flags:      flags,
		timeout:    timeout,
		resChan:    make(chan *rony.MessageEnvelope, 1),
	}
	cb.envelope.Message = append(cb.envelope.Message, reqBytes...)
	cb.envelope.Header = domain.TeamHeader(teamID, teamAccess)
	register(cb)
	return cb
}

func UnmarshalCallback(data []byte) (*callback, error) {
	scb := &serializedCallback{}
	err := json.Unmarshal(data, scb)
	if err != nil {
		return nil, err
	}
	callbacksMtx.Lock()
	cb := requestCallbacks[scb.MessageEnvelope.RequestID]
	callbacksMtx.Unlock()
	if cb != nil {
		return cb, nil
	}
	cb = &callback{
		envelope:   scb.MessageEnvelope.Clone(),
		onComplete: nil,
		onTimeout:  nil,
		onProgress: nil,
		ui:         scb.UI,
		createdOn:  scb.CreatedOn,
		sentOn:     tools.NanoTime(),
		flags:      scb.Flags,
		timeout:    scb.Timeout,
		resChan:    make(chan *rony.MessageEnvelope, 1),
	}
	register(cb)
	return cb, nil
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

func register(cb *callback) {
	callbacksMtx.Lock()
	requestCallbacks[cb.RequestID()] = cb
	callbacksMtx.Unlock()
}

func unregister(reqID uint64) {
	callbacksMtx.Lock()
	delete(requestCallbacks, reqID)
	callbacksMtx.Unlock()
}

func GetCallback(reqID uint64) Callback {
	callbacksMtx.Lock()
	cb := requestCallbacks[reqID]
	callbacksMtx.Unlock()
	if cb == nil {
		return nil
	}
	return cb
}
