package main

import (
	"fmt"
	"github.com/valyala/fasthttp"
	"sync"
	"testing"
	"time"
)

/*
   Creation Time: 2020 - Apr - 27
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2020
*/

func TestLog(t *testing.T) {
	total := 10
	wg := sync.WaitGroup{}
	for i := 0; i < total; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			url := fmt.Sprintf("http://127.0.0.1:2374/p%02d", i)
			for j := 0; j < 100; j++ {
				args := fasthttp.AcquireArgs()
				args.AppendBytes([]byte(fmt.Sprintf("Log from %d in Item %d", i, j)))
				fasthttp.Post(nil, url, args)
				fasthttp.ReleaseArgs(args)
				time.Sleep(time.Second)
			}
		}(i)
	}
	wg.Wait()
}
