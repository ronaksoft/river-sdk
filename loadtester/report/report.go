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
		if r.StoppedActors == 0 {
			r.RequestCount = status.RequestCount
			r.TimedoutRequests = status.TimedoutRequests
			r.SucceedRequests = status.SucceedRequests
			r.ErrorResponses = status.ErrorRespons
			r.NetworkDisconnects = status.NetworkDisconnects
			r.AverageActorLifetime = status.LifeTime
			r.AverageSuccessInterval = status.AverageSuccessInterval
			r.AverageTimeoutInterval = status.AverageTimeoutInterval

		} else {
			r.RequestCount += status.RequestCount
			r.TimedoutRequests += status.TimedoutRequests
			r.SucceedRequests += status.SucceedRequests
			r.ErrorResponses += status.ErrorRespons
			r.NetworkDisconnects += status.NetworkDisconnects
			r.AverageActorLifetime = (r.AverageActorLifetime + status.LifeTime) / 2
			r.AverageSuccessInterval = (r.AverageSuccessInterval + status.AverageSuccessInterval) / 2
			r.AverageTimeoutInterval = (r.AverageTimeoutInterval + status.AverageTimeoutInterval) / 2

		}
		atomic.AddInt64(&r.StoppedActors, 1)
		atomic.AddInt64(&r.ActiveActors, -1)
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
Request Count			: %d
Timedout Requests		: %d
Succeed Requests		: %d
Error Responses			: %d
Network Disconnect		: %d
Avg Actor Lifetime		: %v
Avg Success Interval		: %v
Avg Timeout Interval		: %v

Total Exec Time			: %v

`,
		r.TotalActors,
		r.ActiveActors,
		r.StoppedActors,
		r.RequestCount,
		r.TimedoutRequests,
		r.SucceedRequests,
		r.ErrorResponses,
		r.NetworkDisconnects,
		r.AverageActorLifetime,
		r.AverageSuccessInterval,
		r.AverageTimeoutInterval,
		time.Since(r.Timer),
	)
}

// Print display statistics on screen
func (r *Report) Print() {
	fmt.Printf("\033[0;0H") // print inplace
	fmt.Println(r.String())
	fmt.Println("Total Execution Time : ", time.Since(r.Timer))
	r.isPrinting = false
}

// Clear reset reporter statistics
func (r *Report) Clear() {
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
	r.Timer = time.Now()
}