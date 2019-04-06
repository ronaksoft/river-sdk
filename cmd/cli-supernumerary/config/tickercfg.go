package config

import (
	"time"

	"git.ronaksoftware.com/ronak/riversdk/cmd/cli-loadtester/supernumerary"
)

type TickerCfg struct {
	Duration time.Duration              `json:"duration"`
	Action   supernumerary.TickerAction `json:"action"`
}
