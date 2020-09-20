package executor

import (
	"context"
	"encoding/json"
	"git.ronaksoft.com/river/sdk/internal/logs"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/zap"
	"os"
	"testing"
	"time"
)

/*
   Creation Time: 2020 - Sep - 20
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2020
*/

type dummyAction struct {
	id  int32
	req *dummyRequest
}

func (d *dummyAction) ID() int32 {
	return d.id
}

func (d *dummyAction) Do(ctx context.Context) {
	logs.Info("Do", zap.Int32("ID", d.id))
}

type dummyRequest struct {
	chunks  chan int32
	done    []int32
	hasNext bool
}

func (d *dummyRequest) Prepare() error {
	d.chunks = make(chan int32, 10)
	d.done = d.done[:0]
	for i := int32(0); i < 10; i++ {
		d.chunks <- i
	}
	return nil
}

func (d *dummyRequest) NextAction() Action {
	logs.Debug("NextAction")
	select {
	case id := <-d.chunks:
		logs.Debug("NextAction returns", zap.Int32("ID", id))
		return &dummyAction{
			id:  id,
			req: d,
		}
	default:
		logs.Debug("NextAction returns nil")
		return nil
	}
}

func (d *dummyRequest) ActionDone(id int32) {
	d.done = append(d.done, id)
	logs.Info("Action is done", zap.Int32("ID", id))

	if len(d.chunks) == 0 {
		stopChan <- struct{}{}
	}
}

func (d *dummyRequest) Serialize() []byte {
	logs.Debug("Marshal Called")
	b, err := json.Marshal(d)
	if err != nil {
		panic(err)
	}
	return b
}

func (d *dummyRequest) Next() Request {
	time.Sleep(time.Second)
	if d.hasNext {
		d.hasNext = false
		_ = d.Prepare()
		return d
	}
	return nil
}

var stopChan = make(chan struct{})

func init() {
	logs.SetLogLevel(-1)
}

func TestNewExecutor(t *testing.T) {
	Convey("Executor", t, func(c C) {
		_ = os.MkdirAll("./_hdd", os.ModePerm)
		e, err := NewExecutor("./_hdd", "dummy", func(data []byte) Request {
			r := &dummyRequest{}
			r.chunks = make(chan int32, 10)
			for i := int32(0); i < 10; i++ {
				r.chunks <- i
			}
			err := json.Unmarshal(data, r)
			if err != nil {
				panic(err)
			}
			r.hasNext = true
			return r
		})
		c.So(err, ShouldBeNil)

		r := &dummyRequest{}

		err = e.Execute(r)
		c.So(err, ShouldBeNil)

		<-stopChan
		time.Sleep(time.Second * 3)
	})
}
