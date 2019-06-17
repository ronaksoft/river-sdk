package config

import (
	"git.ronaksoftware.com/ronak/riversdk/cmd/cli-supernumerary/pkg/supernumerary"
	"time"
)

/*
   Creation Time: 2019 - May - 13
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2018
*/

type StartCfg struct {
	ServerURL     string        `json:"server-url"`
	FileServerURL string        `json:"file-server-url"`
	Timeout       time.Duration `json:"timeout"`
	MaxInterval   time.Duration `json:"max_interval"`
}

type TickerCfg struct {
	Duration time.Duration              `json:"duration"`
	Action   supernumerary.TickerAction `json:"action"`
}

type PhoneRangeCfg struct {
	StartPhone int64 `json:"start_phone"`
	EndPhone   int64 `json:"end_phone"`
}

type NodeRegisterCmd struct {
	InstanceID string `json:"instance_id"`
}

type CreateGroup struct {
	StartPhone int64 `json:"start_phone"`
	EndPhone   int64 `json:"end_phone"`
	GroupSize  int64 `json:"group_size"`
}
