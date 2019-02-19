package report

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"git.ronaksoftware.com/ronak/riversdk/loadtester/shared"
	"git.ronaksoftware.com/ronak/riversdk/log"
)

// Report collects reporters statistics
type Report struct {
	mx          sync.Mutex
	isPrinting  bool
	isActive    bool
	Timer       time.Time
	ActorStatus map[string]*shared.Status

	TotalActors   int64
	ActiveActors  int64
	StoppedActors int64

	// statistics
	RequestCount           int64
	TimedoutRequests       int64
	SucceedRequests        int64
	ErrorResponses         int64
	NetworkDisconnects     int64
	AverageActorLifetime   time.Duration
	AverageSuccessInterval time.Duration
	AverageTimeoutInterval time.Duration
	MinActorLifetime       time.Duration
	MaxActorLifetime       time.Duration
	MinSuccessInterval     time.Duration
	MaxSuccessInterval     time.Duration

	SucceedActors int64
	FailedActors  int64

	totalActorLifetime   time.Duration
	totalSuccessInterval time.Duration
	totalTimeoutInterval time.Duration

	lastPrintTime time.Time
}

// NewReport create new instance
func NewReport() shared.Reporter {
	r := new(Report)
	r.ActorStatus = make(map[string]*shared.Status)
	return r
}

// IsActive return active flag
func (r *Report) IsActive() bool {
	return r.isActive
}

// SetIsActive change active flag
func (r *Report) SetIsActive(isActive bool) {
	r.isActive = isActive
}

// Register add actor to be monitored by reporter
func (r *Report) Register(act shared.Acter) {
	r.mx.Lock()
	r.ActorStatus[act.GetPhone()] = act.GetStatus()
	r.mx.Unlock()
	act.SetStopHandler(r.onActorStop)
	atomic.AddInt64(&r.TotalActors, 1)
	atomic.AddInt64(&r.ActiveActors, 1)
}

// onActorStop called to collect statistics data
func (r *Report) onActorStop(phone string) {
	r.mx.Lock()
	status, ok := r.ActorStatus[phone]
	r.mx.Unlock()

	if ok {

		atomic.AddInt64(&r.StoppedActors, 1)
		atomic.AddInt64(&r.ActiveActors, -1)
		r.RequestCount += status.RequestCount
		r.TimedoutRequests += status.TimedoutRequests
		r.SucceedRequests += status.SucceedRequests
		r.ErrorResponses += status.ErrorRespons
		r.NetworkDisconnects += status.NetworkDisconnects

		r.totalActorLifetime += status.LifeTime
		r.totalSuccessInterval += status.AverageSuccessInterval
		r.totalTimeoutInterval += status.AverageTimeoutInterval

		if r.MinActorLifetime > status.LifeTime && status.LifeTime > 0 {
			r.MinActorLifetime = status.LifeTime
		}
		if r.MaxActorLifetime < status.LifeTime && status.LifeTime > 0 {
			r.MaxActorLifetime = status.LifeTime
		}

		if r.MinSuccessInterval > status.AverageSuccessInterval && status.AverageSuccessInterval > 0 {
			r.MinSuccessInterval = status.AverageSuccessInterval
		}
		if r.MaxSuccessInterval < status.AverageSuccessInterval && status.AverageSuccessInterval > 0 {
			r.MaxSuccessInterval = status.AverageSuccessInterval
		}

		if status.ActorSucceed {
			atomic.AddInt64(&r.SucceedActors, 1)
		} else {
			atomic.AddInt64(&r.FailedActors, 1)
		}

		if r.StoppedActors > 0 {
			r.AverageActorLifetime = r.totalActorLifetime / time.Duration(r.StoppedActors)
		}
		if r.SucceedRequests > 0 {
			r.AverageSuccessInterval = r.totalSuccessInterval / time.Duration(r.SucceedRequests)
		}
		if r.TimedoutRequests > 0 {
			r.AverageTimeoutInterval = r.totalTimeoutInterval / time.Duration(r.TimedoutRequests)
		}

		r.mx.Lock()
		delete(r.ActorStatus, phone)
		r.mx.Unlock()
	} else {
		log.LOG_Warn(fmt.Sprintf("Actor(%s) not found in reporter", phone))
	}

	// Print does not need to be thread safe :/
	if r.isActive && (!r.isPrinting || r.StoppedActors == r.TotalActors) {
		r.isPrinting = true
		r.Print()
	}

}

// String format data for print
func (r *Report) String() string {
	return fmt.Sprintf(""+
		`
Total Actors			: %d
Active Actors			: %d
Stopped Actors			: %d
Succeed Actors			: %d
Failed Actors			: %d
Request Count			: %d
Timedout Requests		: %d
Succeed Requests		: %d
Error Responses			: %d
Network Disconnect		: %d
Avg Actor Lifetime		: %v (%v - %v)
Avg Success Interval		: %v (%v - %v)
Avg Timeout Interval		: %v

Total Exec Time			: %v

`,
		r.TotalActors,
		r.ActiveActors,
		r.StoppedActors,
		r.SucceedActors,
		r.FailedActors,
		r.RequestCount,
		r.TimedoutRequests,
		r.SucceedRequests,
		r.ErrorResponses,
		r.NetworkDisconnects,
		r.AverageActorLifetime, r.MinActorLifetime, r.MaxActorLifetime,
		r.AverageSuccessInterval, r.MinSuccessInterval, r.MaxSuccessInterval,
		r.AverageTimeoutInterval,
		time.Since(r.Timer),
	)
}

// Print display statistics on screen
func (r *Report) Print() {
	fmt.Printf("\033[0;0H") // print inplace
	fmt.Println(r.String())
	fmt.Println("Total Execution Time :", time.Since(r.Timer))
	r.isPrinting = false
}

// Clear reset reporter statistics
func (r *Report) Clear() {

	shared.ClearFailedRequest()

	r.ActorStatus = make(map[string]*shared.Status)
	r.TotalActors = 0
	r.ActiveActors = 0
	r.StoppedActors = 0
	r.AverageActorLifetime = 0
	r.RequestCount = 0
	r.TimedoutRequests = 0
	r.SucceedRequests = 0
	r.NetworkDisconnects = 0
	r.AverageSuccessInterval = 0
	r.AverageTimeoutInterval = 0
	r.ErrorResponses = 0
	r.SucceedActors = 0
	r.FailedActors = 0
	r.totalActorLifetime = 0
	r.totalSuccessInterval = 0
	r.totalTimeoutInterval = 0
	// r.lastPrintTime = 0

	r.MinActorLifetime = time.Duration(^uint64(0) >> 1)
	r.MinSuccessInterval = time.Duration(^uint64(0) >> 1)
	r.MaxActorLifetime = 0
	r.MaxSuccessInterval = 0

	r.Timer = time.Now()
}
