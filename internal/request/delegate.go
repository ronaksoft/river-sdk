package request

import (
    "strings"

    "github.com/ronaksoft/river-sdk/internal/domain"
    "github.com/ronaksoft/rony"
    "github.com/ronaksoft/rony/pools"
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
    if rdf&RetryUntilCanceled == RetryUntilCanceled {
        sb.WriteString("|RetryUntilCancel")
    }
    if rdf > 0 {
        sb.WriteRune('|')
    }
    return sb.String()
}

// Request Flags
const (
    // ServerForced sends the request to the server even  if there is a local handler registered for it.
    ServerForced DelegateFlag = 1 << iota
    // Blocking blocks the caller until response arrived from the server
    Blocking
    // SkipWaitForNetwork starts the timeout timer right after submitting the request, and does not
    // wait until network connection is established.
    SkipWaitForNetwork
    // SkipFlusher prevents sending the request in a container.
    SkipFlusher
    // Realtime sends the request directly to the network and skips the persistent queue. Such requests
    // are forgotten after app restart.
    Realtime
    // Batch waits longer than usual. This is good for burst request. i.e. Call module uses this flag to prevent
    // flooding server with individual updates.
    Batch
    // RetryUntilCanceled makes the request to be retried in case of timeout.
    RetryUntilCanceled
)

type Delegate interface {
    OnComplete(b []byte)
    OnTimeout(err error)
    Flags() DelegateFlag
}

func DelegateAdapter(
        teamID int64, teamAccess uint64, reqID uint64, constructor int64, reqBytes []byte, d Delegate, progressFunc func(int64),
) *callback {
    onTimeout := func() {}
    onComplete := func(m *rony.MessageEnvelope) {}
    onProgress := func(progress int64) {}
    flags := DelegateFlag(0)
    if d != nil {
        onTimeout = func() {
            d.OnTimeout(domain.ErrRequestTimeout)
        }
        onComplete = func(m *rony.MessageEnvelope) {
            buf := pools.Buffer.FromProto(m)
            d.OnComplete(*buf.Bytes())
            pools.Buffer.Put(buf)
        }
        flags = d.Flags()
    }
    if progressFunc != nil {
        onProgress = progressFunc
    }
    return NewCallbackFromBytes(
        teamID, teamAccess, reqID, constructor, reqBytes,
        onTimeout, onComplete, onProgress, true, flags,
        domain.WebsocketRequestTimeout,
    )
}
