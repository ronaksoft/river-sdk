package request

import (
	"git.ronaksoft.com/river/sdk/internal/domain"
	"github.com/ronaksoft/rony"
	"github.com/ronaksoft/rony/pools"
	"strings"
)

/*
   Creation Time: 2021 - May - 25
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2020
*/

type DelegateFlag = int32

func DelegateFlagToString(rdf DelegateFlag) string {
	sb := strings.Builder{}
	if rdf&ServerForced == ServerForced {
		sb.WriteString("|ServerForced")
	}
	if rdf&Blocking == Blocking {
		sb.WriteString("|Blocking")
	}
	if rdf&SkipWaitForNetwork == SkipWaitForNetwork {
		sb.WriteString("|SkipWaitNetwork")
	}
	if rdf&SkipFlusher == SkipFlusher {
		sb.WriteString("|SkipFlusher")
	}
	if rdf&Realtime == Realtime {
		sb.WriteString("|Realtime")
	}
	if rdf&Batch == Batch {
		sb.WriteString("|Batch")
	}
	sb.WriteRune('|')
	return sb.String()
}

// Request Flags
const (
	ServerForced DelegateFlag = 1 << iota
	Blocking
	SkipWaitForNetwork
	SkipFlusher
	Realtime
	Batch
)

type Delegate interface {
	OnComplete(b []byte)
	OnTimeout(err error)
	OnProgress(percent int64)
	Flags() DelegateFlag
}

type delegateAdapter struct {
	d  Delegate
	ui bool
}

func DelegateAdapter(d Delegate, ui bool) *delegateAdapter {
	return &delegateAdapter{
		d:  d,
		ui: ui,
	}
}

func (rda *delegateAdapter) OnComplete(m *rony.MessageEnvelope) {
	if rda.d == nil {
		return
	}
	buf := pools.Buffer.FromProto(m)
	rda.d.OnComplete(*buf.Bytes())
	pools.Buffer.Put(buf)
}

func (rda *delegateAdapter) OnTimeout() {
	if rda.d == nil {
		return
	}
	rda.d.OnTimeout(domain.ErrRequestTimeout)
}

func (rda *delegateAdapter) OnProgress(percent int64) {
	if rda.d == nil {
		return
	}
	rda.d.OnProgress(percent)
}

func (rda *delegateAdapter) UI() bool {
	return rda.ui
}

type delegate struct {
	onComplete func(b []byte)
	onTimeout  func(error)
	onProgress func(int64)
	flags      DelegateFlag
}

func (r *delegate) OnComplete(b []byte) {
	if r.onComplete != nil {
		r.onComplete(b)
	}
}

func (r *delegate) OnTimeout(err error) {
	if r.onTimeout != nil {
		r.onTimeout(err)
	}
}

func (r *delegate) Flags() DelegateFlag {
	return r.flags
}

func (r *delegate) OnProgress(percent int64) {
	if r.onProgress != nil {
		r.onProgress(percent)
	}
}

func NewRequestDelegate(onComplete func(b []byte), onTimeout func(err error), flags DelegateFlag) *delegate {
	return &delegate{
		onComplete: onComplete,
		onTimeout:  onTimeout,
		onProgress: nil,
		flags:      flags,
	}
}
