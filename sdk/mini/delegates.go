package mini

import (
	"git.ronaksoft.com/river/sdk/internal/domain"
	"github.com/ronaksoft/rony"
	"github.com/ronaksoft/rony/pools"
)

// MainDelegate external (UI) handler will listen to this function to receive data from SDK
type MainDelegate interface {
	OnNetworkStatusChanged(status int)
	OnSyncStatusChanged(status int)
	OnUpdates(constructor int64, b []byte)
	OnGeneralError(b []byte)
	OnSessionClosed(res int)
	ShowLoggerAlert()
	AddLog(text string)
	AppUpdate(version string, updateAvailable, force bool)
}

// RequestDelegate each request should have this callbacks
type RequestDelegate interface {
	OnComplete(b []byte)
	OnTimeout(err error)
	Flags() int32
	OnProgress(percent int64)
}

// Request Flags
const (
	RequestServerForced int32 = 1 << iota
	RequestBlocking
	RequestDontWaitForNetwork
	RequestTeamForce
)

type DelegateAdapter struct {
	d RequestDelegate
}

func NewDelegateAdapter(d RequestDelegate) *DelegateAdapter {
	return &DelegateAdapter{
		d: d,
	}
}

func (rda *DelegateAdapter) OnComplete(m *rony.MessageEnvelope) {
	buf := pools.Buffer.FromProto(m)
	rda.d.OnComplete(*buf.Bytes())
	pools.Buffer.Put(buf)
}

func (rda *DelegateAdapter) OnTimeout() {
	rda.d.OnTimeout(domain.ErrRequestTimeout)
}

func (rda *DelegateAdapter) OnProgress(percent int64) {
	rda.d.OnProgress(percent)
}
