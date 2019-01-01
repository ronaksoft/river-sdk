package shared

import "time"

// Status info
type Status struct {
	LifeTime           time.Duration
	RequestCount       int64
	TimeoutCount       int64
	SuccessCount       int64
	DisconnectCount    int64
	AverageSuccessTime time.Duration
	AverageTimeoutTime time.Duration
}
