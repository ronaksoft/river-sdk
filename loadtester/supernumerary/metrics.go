package supernumerary

import (
	"os"

	ronak "git.ronaksoftware.com/ronak/toolbox"
)

var (
	Metrics *ronak.Prometheus

	CntInitFailedActor = "init_failed_actors"

	CntRequestCount       = "request"
	CntTimedoutRequests   = "timeout"
	CntSucceedRequests    = "succeed"
	CntErrorRespons       = "error"
	CntNetworkDisconnects = "disconnect"

	HistLifeTime        = "lifetime_interval"
	HistSuccessInterval = "success_interval"
	HistTimeoutInterval = "timeout_interval"
)

func init() {
	boundleID := os.Getenv("CFG_BOUNDLE_ID")
	instanceID := os.Getenv("CFG_INSTANCE_ID")
	Metrics = ronak.NewPrometheus(boundleID, instanceID)
	// register metrics
	Metrics.RegisterCounter(
		CntInitFailedActor, "Number of actors that failed to create new instance", nil,
	)

	Metrics.RegisterCounter(
		CntRequestCount, "Number of requests that sent to server", nil,
	)
	Metrics.RegisterCounter(
		CntTimedoutRequests, "Number of requests that timedout", nil,
	)
	Metrics.RegisterCounter(
		CntSucceedRequests, "Number of succedd requests", nil,
	)
	Metrics.RegisterCounter(
		CntErrorRespons, "Number received error response from server", nil,
	)
	Metrics.RegisterCounter(
		CntNetworkDisconnects, "Number of websocket error", nil,
	)
}
