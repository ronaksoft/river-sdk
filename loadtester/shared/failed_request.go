package shared

import (
	"fmt"
	"strings"
	"sync"
)

var (
	FailedRequests   []*FailedRequest
	FailedRequestsMx sync.Mutex
)

type FailedRequest struct {
	RequestID uint64
	AuthID    int64
	Actor     string
	Comment   string
}

func (f *FailedRequest) String() string {
	return fmt.Sprintf("Actor: %s \t ReqID: %d \t AuthID: %d \t Err: %s \n", f.Actor, f.RequestID, f.AuthID, f.Comment)
}

func init() {
	FailedRequests = make([]*FailedRequest, 0, 1000)
}

func SetFailedRequest(reqID uint64, authID int64, actorName, comment string) {
	req := FailedRequest{
		RequestID: reqID,
		AuthID:    authID,
		Actor:     actorName,
		Comment:   comment,
	}
	FailedRequestsMx.Lock()
	FailedRequests = append(FailedRequests, &req)
	FailedRequestsMx.Unlock()

}

func ClearFailedRequest() {
	FailedRequestsMx.Lock()
	FailedRequests = FailedRequests[:0]
	FailedRequestsMx.Unlock()
}

func PrintFailedRequest() string {
	FailedRequestsMx.Lock()

	sb := strings.Builder{}
	for _, v := range FailedRequests {
		sb.WriteString(v.String())
	}
	FailedRequestsMx.Unlock()

	return sb.String()
}
