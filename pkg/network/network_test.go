package network_test

import (
	"git.ronaksoftware.com/ronak/riversdk/pkg/network"
	"testing"
	"time"
)

/*
   Creation Time: 2019 - Apr - 15
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2018
*/

func TestNewController(t *testing.T) {
	ctrl := network.NewController(network.Config{
		ServerEndpoint: "ws://new.river.im",
		PingTime:       30 * time.Second,
		PongTimeout:    30 * time.Second,
	})
}
