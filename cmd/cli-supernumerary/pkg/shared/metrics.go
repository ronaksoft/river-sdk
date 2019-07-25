package shared

import (
	ronak "git.ronaksoftware.com/ronak/toolbox"
)

var (
	Metrics *ronak.Prometheus

	CntInitFailedActor = "init_failed_actors"

	CntRequest         = "request_count"          // vec
	CntResponse        = "response_count"         // vec
	CntDiffered        = "differed_count"         // vec
	CntTimeout         = "timeout_count"          // vec
	CntSuccess         = "success_count"          // vec
	CntError           = "error_count"            // counter
	CntDisconnect      = "disconnect_count"       // counter
	CntFile            = "file_count"             // counter
	CntFileError       = "file_error_count"       // counter
	CntFailedScenario  = "failed_scenario_count"  // counter
	CntSucceedScenario = "succeed_scenario_count" // counter

	// Network Usage
	CntSend    = "send"    // counter
	CntReceive = "receive" // counter

	HistFileLatency     = "file_latency"     // hist
	HistSuccessLatency  = "success_latency"  // hist
	HistTimeoutLatency  = "timeout_latency"  // hist
	HistDifferedLatency = "differed_latency" // hist

	GaugeActors = "actor" // gauge
)

var (
	SizeBucketKB     = []float64{1, 5, 10, 20, 40, 80, 160, 320, 640, 1280, 2560}
	TimeBucketMS     = []float64{0.1, 0.5, 1, 10, 20, 50, 100, 500, 1000, 2000, 3000, 5000, 7000, 10000, 15000, 20000, 30000}
	TimeBucketMicroS = []float64{1, 2, 5, 10, 20, 50, 100, 200, 300, 400, 500, 1000, 2000, 5000, 10000}
)

func InitMetrics(boundleID, instanceID string) {

	Metrics = ronak.NewPrometheus(boundleID, instanceID)

	// Counter vectors
	Metrics.RegisterCounterVec(CntRequest, "Number of message requests", nil, []string{"fn"})            // ok
	Metrics.RegisterCounterVec(CntResponse, "Number of received responses/updates", nil, []string{"fn"}) // ok
	Metrics.RegisterCounterVec(CntDiffered, "Number of late received responses", nil, []string{"fn"})    // ok
	Metrics.RegisterCounterVec(CntTimeout, "Number of timeout requests", nil, []string{"fn"})            // ok
	Metrics.RegisterCounterVec(CntSuccess, "Number of success requests", nil, []string{"fn"})            // ok

	// Counter
	Metrics.RegisterCounter(CntError, "Number of error responses", nil)                   // ok
	Metrics.RegisterCounter(CntDisconnect, "Number of websocket disconnections", nil)     // ok
	Metrics.RegisterCounter(CntFile, "Number of FileSavePart requests ", nil)             // ok
	Metrics.RegisterCounter(CntFileError, "Number of failed FileSavePart requests ", nil) // ok
	Metrics.RegisterCounter(CntFailedScenario, "Number of failed scenarios ", nil)        // ok
	Metrics.RegisterCounter(CntSucceedScenario, "Number of succeed scenarios ", nil)      // ok
	Metrics.RegisterCounter(CntSend, "Upload/SendWebsocket bandwidth", nil)               // ok
	Metrics.RegisterCounter(CntReceive, "Download/Receive bandwidth", nil)                // ok

	// Histogram
	Metrics.RegisterHistogram(HistFileLatency, "Histogram of the FileSavePart request latency", TimeBucketMS, nil)        // ok
	Metrics.RegisterHistogram(HistSuccessLatency, "Histogram of the success requests latency", TimeBucketMS, nil)         // ok
	Metrics.RegisterHistogram(HistTimeoutLatency, "Histogram of the timedout requests latency", TimeBucketMS, nil)        // ok
	Metrics.RegisterHistogram(HistDifferedLatency, "Histogram of the late received responses latency", TimeBucketMS, nil) // ok

	// Gauge
	Metrics.RegisterGauge(GaugeActors, "Number of actors", nil) // ok
}
