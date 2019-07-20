package mon

import (
	"fmt"
	"testing"
	"time"
)

/*
   Creation Time: 2019 - Jul - 16
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2018
*/

func TestFunctionResponseTime(t *testing.T) {
	for i := 0; i < 10; i++ {
		FunctionResponseTime("F1", time.Second*time.Duration(i))
	}
	fmt.Println(Stats.MaxFunctionResponseTime, Stats.MinFunctionResponseTime)
	fmt.Println(Stats.AvgFunctionResponseTime)
	fmt.Println(Stats.TotalFunctionCalls)

}
