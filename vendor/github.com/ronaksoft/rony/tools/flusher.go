package tools

import (
	"sync/atomic"
)

/*
   Creation Time: 2020 - Dec - 31
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2020
*/

type FlushEntry interface {
	wait() chan struct{}
	done()
	Value() interface{}
}

type entry struct {
	v  interface{}
	ch chan struct{}
}

func NewEntry(v interface{}) FlushEntry {
	return &entry{
		v:  v,
		ch: make(chan struct{}, 1),
	}
}

func (e *entry) wait() chan struct{} {
	return e.ch
}

func (e *entry) done() {
	e.ch <- struct{}{}
}

func (e *entry) Value() interface{} {
	return e.v
}

type FlusherFunc func(targetID string, entries []FlushEntry)

type FlusherPool struct {
	maxWorkers  int32
	batchSize   int32
	flusherFunc FlusherFunc
	poolMtx     SpinLock
	pool        map[string]*flusher
}

// NewFlusherPool creates a pool of flusher funcs. By calling Enter or EnterAndWait you add
// the item into the flusher which identified by 'targetID'.
func NewFlusherPool(maxWorkers, batchSize int32, f FlusherFunc) *FlusherPool {
	fp := &FlusherPool{
		maxWorkers:  maxWorkers,
		batchSize:   batchSize,
		flusherFunc: f,
		pool:        make(map[string]*flusher, 16),
	}
	return fp
}

func (fp *FlusherPool) getFlusher(targetID string) *flusher {
	fp.poolMtx.Lock()
	f := fp.pool[targetID]
	fp.poolMtx.Unlock()
	if f == nil {
		f = &flusher{
			readyWorkers: fp.maxWorkers,
			batchSize:    fp.batchSize,
			flusherFunc:  fp.flusherFunc,
			entryChan:    make(chan FlushEntry, fp.batchSize),
			targetID:     targetID,
		}
		fp.poolMtx.Lock()
		fp.pool[targetID] = f
		fp.poolMtx.Unlock()
	}
	return f
}

func (fp *FlusherPool) Enter(targetID string, entry FlushEntry) {
	fp.getFlusher(targetID).enter(entry)
}

func (fp *FlusherPool) EnterAndWait(targetID string, entry FlushEntry) {
	fp.getFlusher(targetID).enterAndWait(entry)
}

type flusher struct {
	SpinLock
	readyWorkers int32
	batchSize    int32
	flusherFunc  FlusherFunc
	entryChan    chan FlushEntry
	targetID     string
}

func (f *flusher) startWorker() {
	f.Lock()
	if atomic.AddInt32(&f.readyWorkers, -1) < 0 {
		atomic.AddInt32(&f.readyWorkers, 1)
		f.Unlock()
		return
	}
	f.Unlock()

	w := &worker{
		f:  f,
		bs: int(f.batchSize),
	}
	go w.run()
}

func (f *flusher) enter(entry FlushEntry) {
	f.entryChan <- entry
	f.startWorker()
}

func (f *flusher) enterAndWait(entry FlushEntry) {
	f.entryChan <- entry
	f.startWorker()
	<-entry.wait()
}

type worker struct {
	f  *flusher
	bs int
}

func (w *worker) run() {
	var (
		el = make([]FlushEntry, 0, w.bs)
	)
	for {
		el = el[:0]
		for {
			select {
			case e := <-w.f.entryChan:
				el = append(el, e)
				if len(el) < w.bs {
					continue
				}
			default:
			}
			break
		}

		w.f.Lock()
		if len(el) == 0 {
			// clean up and shutdown the worker
			atomic.AddInt32(&w.f.readyWorkers, 1)
			w.f.Unlock()
			break
		}
		w.f.Unlock()
		w.f.flusherFunc(w.f.targetID, el)
		for idx := range el {
			el[idx].done()
		}
	}

}