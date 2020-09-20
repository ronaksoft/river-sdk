package executor

import (
	"context"
	"github.com/beeker1121/goque"
	"path/filepath"
	"sync"
)

/*
   Creation Time: 2020 - Sep - 20
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2020
*/

const (
	defaultConcurrency = 10
)

type RequestFactoryFunc func() Request

// Executor
type Executor struct {
	name    string
	stack   *goque.Stack
	factory RequestFactoryFunc

	// internals
	ctx     context.Context
	cf      context.CancelFunc
	rt      chan struct{}
	mtx     sync.Mutex
	running bool
}

func NewExecutor(dbPath string, name string, factory RequestFactoryFunc, opts ...Option) (*Executor, error) {
	s, err := goque.OpenStack(filepath.Join(dbPath, name))
	if err != nil {
		return nil, err
	}
	e := &Executor{
		stack:   s,
		name:    name,
		factory: factory,
		rt:      make(chan struct{}, defaultConcurrency),
	}
	e.ctx, e.cf = context.WithCancel(context.Background())

	// Apply options
	for _, opt := range opts {
		opt(e)
	}

	return e, nil
}

func (e *Executor) execute() {
	for {
		e.mtx.Lock()
		if e.stack.Length() == 0 {
			e.running = false
			e.mtx.Unlock()
			break
		}
		e.mtx.Unlock()

		// Pop the next request from the stack
		stackItem, err := e.stack.Pop()
		if err != nil {
			return
		}

		// Check rate limit and if it is ok then run the job in background
		e.rt <- struct{}{}
		go func() {
			defer func() {
				<-e.rt
			}()

			req := e.factory()
			req.Deserialize(stackItem.Value)
			err := req.Prepare()
			if err != nil {
				return
			}
			for {
				act := req.NextAction()
				if act == nil {
					break
				}
				act.Do(e.ctx)
				req.ActionDone(act.ID())
			}
		}()
	}
}

func (e *Executor) Execute(req Request) error {
	_, err := e.stack.Push(req.Serialize())
	if err != nil {
		return err
	}
	e.mtx.Lock()
	if !e.running {
		e.running = true
		go e.execute()
	}
	e.mtx.Unlock()
	return nil
}
