package domain

import (
	"sync"
	"sync/atomic"
	"time"
)

/*
   Creation Time: 2019 - Feb - 28
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2018
*/

type FlusherEntry struct {
	Key            interface{}
	Value          interface{}
	Ret            chan interface{}
	CallbackSynced bool
	Callback       func(ret interface{})
}

type Flusher struct {
	entries        chan FlusherEntry
	next           chan struct{}
	nextID         int64
	flushPeriod    time.Duration
	maxWorkers     int32
	maxBatchSize   int
	runningWorkers int32
	workerFunc     FlusherFunc

	// Ticket Storage System
	ticket uint64
	done   uint64
	lock   int32
}

type FlusherFunc func(items []FlusherEntry)

func NewFlusher(maxBatchSize, maxConcurrency int32, flushPeriod time.Duration, ff FlusherFunc) *Flusher {
	f := new(Flusher)
	f.flushPeriod = flushPeriod
	f.maxWorkers = maxConcurrency
	f.maxBatchSize = int(maxBatchSize)
	f.entries = make(chan FlusherEntry, f.maxBatchSize)
	f.next = make(chan struct{}, f.maxWorkers)
	f.workerFunc = ff

	// Run the 1st instance in the background
	for i := int32(0); i < f.maxWorkers; i++ {
		go f.worker()
	}

	// Set the first worker to listen on the channel
	f.next <- struct{}{}

	return f
}

// EnterWithChan
func (f *Flusher) EnterWithChan(key, value interface{}) chan interface{} {
	ch := make(chan interface{}, 1)
	f.entries <- FlusherEntry{
		Key:   key,
		Value: value,
		Ret:   ch,
	}
	return ch
}

// EnterWithResult
func (f *Flusher) EnterWithResult(key, value interface{}) (res interface{}) {
	waitGroup := waitGroupPool.Get().(*sync.WaitGroup)
	waitGroup.Add(1)
	f.entries <- FlusherEntry{
		Key:            key,
		Value:          value,
		CallbackSynced: true,
		Callback: func(ret interface{}) {
			res = ret
			waitGroup.Done()
		},
	}
	waitGroup.Wait()
	waitGroupPool.Put(waitGroup)
	return res
}

// EnterWithSyncCallback
// Use this function if the cbFunc is fast enough since, they will be called synchronously
func (f *Flusher) EnterWithSyncCallback(key, value interface{}, cbFunc func(interface{})) {
	f.entries <- FlusherEntry{
		Key:            key,
		Value:          value,
		CallbackSynced: true,
		Callback:       cbFunc,
	}
}

// EnterWithAsyncCallback
func (f *Flusher) EnterWithAsyncCallback(key, value interface{}, cbFunc func(interface{})) {
	f.entries <- FlusherEntry{
		Key:            key,
		Value:          value,
		CallbackSynced: false,
		Callback:       cbFunc,
	}
}

// Enter
// If you calling this function make sure that the FlusherFunc registered with this Flusher
// can handle nil channels otherwise you may encounter undetermined results
func (f *Flusher) Enter(key, value interface{}) {
	f.entries <- FlusherEntry{
		Key:   key,
		Value: value,
		Ret:   nil,
	}
}

// PendingItems returns the number of items are still in the channel and are not picked by any job worker
func (f *Flusher) PendingItems() int {
	return len(f.entries)
}

// RunningJobs returns the number of job workers currently picking up items from the channel and/or running the flusher func
func (f *Flusher) RunningJobs() int {
	return int(f.runningWorkers)
}

func (f *Flusher) worker() {
	items := make([]FlusherEntry, 0, f.maxBatchSize)
	t := AcquireTimer(f.flushPeriod)
	defer ReleaseTimer(t)

	for {
		items = items[:0]

		// Wait for next signal to run the job
		<-f.next
		atomic.AddInt32(&f.runningWorkers, 1)

		// Wait for next entry
		select {
		case item := <-f.entries:
			items = append(items, item)
		}
		ResetTimer(t, f.flushPeriod)
	InnerLoop:
		for {
			select {
			case item := <-f.entries:
				items = append(items, item)
				if len(items) >= f.maxBatchSize {
					// Send signal for the next worker to listen for entries
					f.next <- struct{}{}

					// Run the job synchronously
					f.workerFunc(items)

					atomic.AddInt32(&f.runningWorkers, -1)
					break InnerLoop
				}
			case <-t.C:
				// Send signal for the next worker to listen for entries
				f.next <- struct{}{}

				// Run the job synchronously
				f.workerFunc(items)

				atomic.AddInt32(&f.runningWorkers, -1)
				break InnerLoop
			}
		}
	}
}

const (
	maxWorkerIdleTime = 3 * time.Second
)

var waitGroupPool = sync.Pool{
	New: func() interface{} {
		return &sync.WaitGroup{}
	},
}
