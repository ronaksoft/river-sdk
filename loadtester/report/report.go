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

	TotalActor    int64
	ActiveActors  int64
	StoppedActors int64

	// statistics
	AverageLifeTime    time.Duration
	RequestCount       int64
	TimeoutCount       int64
	SuccessCount       int64
	ErrorCount         int64
	DisconnectCount    int64
	AverageSuccessTime time.Duration
	AverageTimeoutTime time.Duration

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
	atomic.AddInt64(&r.TotalActor, 1)
	atomic.AddInt64(&r.ActiveActors, 1)
}

// onActorStop called to collect statistics data
func (r *Report) onActorStop(phone string) {
	r.mx.Lock()
	status, ok := r.ActorStatus[phone]
	r.mx.Unlock()

	if ok {
		if r.StoppedActors == 0 {
			r.AverageLifeTime = status.LifeTime
			r.RequestCount = status.RequestCount
			r.TimeoutCount = status.TimeoutCount
			r.SuccessCount = status.SuccessCount
			r.ErrorCount = status.ErrorRespons
			r.DisconnectCount = status.DisconnectCount
			r.AverageSuccessTime = status.AverageSuccessTime
			r.AverageTimeoutTime = status.AverageTimeoutTime

		} else {
			r.AverageLifeTime = (r.AverageLifeTime + status.LifeTime) / 2
			r.RequestCount += status.RequestCount
			r.TimeoutCount += status.TimeoutCount
			r.SuccessCount += status.SuccessCount
			r.ErrorCount += status.ErrorRespons
			r.DisconnectCount += status.DisconnectCount
			r.AverageSuccessTime = (r.AverageSuccessTime + status.AverageSuccessTime) / 2
			r.AverageTimeoutTime = (r.AverageTimeoutTime + status.AverageTimeoutTime) / 2

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
	if r.isActive && (!r.isPrinting || r.StoppedActors == r.TotalActor) {
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
		r.TotalActor,
		r.ActiveActors,
		r.StoppedActors,
		r.RequestCount,
		r.TimeoutCount,
		r.SuccessCount,
		r.ErrorCount,
		r.DisconnectCount,
		r.AverageLifeTime,
		r.AverageSuccessTime,
		r.AverageTimeoutTime,
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
	r.TotalActor = 0
	r.ActiveActors = 0
	r.StoppedActors = 0
	r.AverageLifeTime = 0
	r.RequestCount = 0
	r.TimeoutCount = 0
	r.SuccessCount = 0
	r.DisconnectCount = 0
	r.AverageSuccessTime = 0
	r.AverageTimeoutTime = 0
	r.Timer = time.Now()
}
