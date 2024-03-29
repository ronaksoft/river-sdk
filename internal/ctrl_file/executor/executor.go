package executor

import (
    "context"
    "path/filepath"
    "sync"

    "github.com/beeker1121/goque"
    "github.com/ronaksoft/river-sdk/internal/logs"
    "go.uber.org/zap"
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
    defaultConcurrentRequests = 5
    defaultConcurrentAction   = 5
)

type RequestFactoryFunc func(data []byte) Request

type Executor struct {
    name    string
    stack   *goque.Stack
    factory RequestFactoryFunc
    logger  *logs.Logger

    // internals
    waitGroupsLock sync.Mutex
    waitGroups     map[string]*sync.WaitGroup
    ctx            context.Context
    cf             context.CancelFunc
    rt             chan struct{}
    mtx            sync.Mutex
    running        bool
}

func NewExecutor(dbPath string, name string, factory RequestFactoryFunc, opts ...Option) (*Executor, error) {
    s, err := goque.OpenStack(filepath.Join(dbPath, name))
    if err != nil {
        return nil, err
    }
    e := &Executor{
        stack:      s,
        name:       name,
        factory:    factory,
        rt:         make(chan struct{}, defaultConcurrentRequests),
        waitGroups: make(map[string]*sync.WaitGroup),
        logger:     logs.With("FileExecutor"),
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
            return
        }
        e.mtx.Unlock()

        // Pop the next request from the stack
        stackItem, err := e.stack.Pop()
        if err != nil {
            e.logger.Fatal("got Serious Error", zap.Error(err))
            return
        }

        // Check rate limit and if it is ok then run the job in background
        e.rt <- struct{}{}
        go func() {
            defer func() {
                <-e.rt
            }()

            req := e.factory(stackItem.Value)
            e.waitGroupsLock.Lock()
            waitGroup := e.waitGroups[req.GetID()]
            if waitGroup != nil {
                delete(e.waitGroups, req.GetID())
            }
            e.waitGroupsLock.Unlock()

            oWaitGroup := &sync.WaitGroup{}
            // External loop over requests
            for req != nil {
                err := req.Prepare()
                if err != nil {
                    e.logger.Warn("got error on Prepare",
                        zap.String("ReqID", req.GetID()),
                        zap.Error(err),
                    )
                    req = req.Next()
                    continue
                }

                oWaitGroup.Add(1)
                go func(req Request) {
                    defer oWaitGroup.Done()
                    iMutex := sync.Mutex{}
                    iRateLimit := make(chan struct{}, defaultConcurrentAction)
                    iWaitGroup := sync.WaitGroup{}
                    // Run actions in a loop
                    for {
                        act := req.NextAction()
                        if act == nil {
                            // Wait for all actions to be done before going to next request
                            iWaitGroup.Wait()
                            break
                        }
                        iRateLimit <- struct{}{}
                        iWaitGroup.Add(1)
                        go func() {
                            act.Do(e.ctx)
                            iMutex.Lock()
                            req.ActionDone(act.ID())
                            iMutex.Unlock()
                            iWaitGroup.Done()
                            <-iRateLimit
                        }()
                    }
                }(req)

                // Goto next chained request
                req = req.Next()
            }

            // Wait for all chained requests to finish
            oWaitGroup.Wait()

            // If there is a waiter for this request, then mark it done.
            if waitGroup != nil {
                waitGroup.Done()
            }

        }()
    }
}

// Execute pushes the request into the stack. This function is non-blocking. If you need a blocking call
// you should use ExecuteAndWait function
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

// ExecuteAndWait accepts a waitGroup which its Done() function will be called  when process is done
func (e *Executor) ExecuteAndWait(waitGroup *sync.WaitGroup, req Request) error {
    e.waitGroupsLock.Lock()
    e.waitGroups[req.GetID()] = waitGroup
    e.waitGroupsLock.Unlock()
    return e.Execute(req)
}

// Option to config Executor
type Option func(e *Executor)

func WithConcurrency(c int32) Option {
    return func(e *Executor) {
        e.rt = make(chan struct{}, c)
    }
}
