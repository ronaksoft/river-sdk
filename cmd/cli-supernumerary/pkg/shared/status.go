package shared

import "time"

// Status info
type Status struct {
	RequestCount           int64
	TimeoutRequests        int64
	SucceedRequests        int64
	ErrorResponses         int64
	NetworkDisconnects     int64
	ActorSucceed           bool
	LifeTime               time.Duration
	AverageSuccessInterval time.Duration
	AverageTimeoutInterval time.Duration
}
