package ronak

import (
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
	Key   interface{}
	Value interface{}
	Ret   chan interface{}
}

type Flusher struct {
	entries        chan FlusherEntry
	next           chan struct{}
	flushPeriod    time.Duration
	maxWorkers     int32
	maxBatchSize   int
	runningWorkers int32
	workerFunc     FlusherFunc
}

type FlusherFunc func(items []FlusherEntry)

func NewFlusher(maxBatchSize, maxConcurrency int32, flushPeriod time.Duration, ff FlusherFunc) *Flusher {
	f := new(Flusher)
	f.entries = make(chan FlusherEntry, maxBatchSize)
	f.next = make(chan struct{}, maxConcurrency)
	f.flushPeriod = flushPeriod
	f.maxWorkers = maxConcurrency
	f.maxBatchSize = int(maxBatchSize)
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
	items := make([]FlusherEntry, 0)
	timer := time.NewTimer(f.flushPeriod)

	for {
		items = items[:0]

		// Wait for next signal to start the job
		<-f.next
		atomic.AddInt32(&f.runningWorkers, 1)

		// Wait for next entry
		select {
		case item := <-f.entries:
			items = append(items, item)
		}
		if !timer.Stop() {
			<-timer.C
		}
		timer.Reset(f.flushPeriod)
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
			case <-timer.C:
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
