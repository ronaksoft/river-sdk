package tools_test

import (
	"fmt"
	"git.ronaksoft.com/ronak/riversdk/internal/tools"
	"sync"
	"testing"
)

/*
   Creation Time: 2020 - May - 16
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2020
*/

func TestSecureRandomInt63(t *testing.T) {
	mtx := sync.Mutex{}
	wg := sync.WaitGroup{}
	var (
		concurrency = 100
		loopSize    = 1000
		m           = make(map[int64]struct{}, concurrency*loopSize)
	)
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < loopSize; j++ {
				mtx.Lock()
				m[tools.SecureRandomInt63(0)] = struct{}{}
				mtx.Unlock()
			}
		}()
	}
	wg.Wait()
	fmt.Println(len(m))
}
