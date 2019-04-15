package ronak

import (
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

func (f *Flusher) worker() {
	items := make([]FlusherEntry, 0)
	ticker := time.NewTicker(f.flushPeriod)
	defer ticker.Stop()
	for {
		items = items[:0]

		// Wait for next signal to start the job
		<-f.next

		// Wait for next entry
		select {
		case item := <-f.entries:
			items = append(items, item)
		}
	InnerLoop:
		for {
			select {
			case item := <-f.entries:
				items = append(items, item)
			case <-ticker.C:
				// Send signal for the next worker to listen for entries
				f.next <- struct{}{}

				// Run the job synchronously
				f.workerFunc(items)

				break InnerLoop
			}
		}
	}
}
