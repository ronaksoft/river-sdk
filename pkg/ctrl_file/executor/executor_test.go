package executor

import (
	"context"
	"encoding/json"
	"git.ronaksoft.com/river/sdk/internal/logs"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/zap"
	"os"
	"testing"
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
	chunks chan int32
	done   []int32
}

func (d *dummyRequest) Prepare() {}

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

func (d *dummyRequest) Deserialize(b []byte) {
	logs.Debug("Unmarshal Called")
	d.chunks = make(chan int32, 10)
	for i := int32(0); i < 10; i++ {
		d.chunks <- i
	}
	err := json.Unmarshal(b, d)
	if err != nil {
		panic(err)
	}
}

var stopChan = make(chan struct{})

func init() {
	logs.SetLogLevel(-1)
}

func TestNewExecutor(t *testing.T) {
	Convey("Executor", t, func(c C) {
		_ = os.MkdirAll("./_hdd", os.ModePerm)
		e, err := NewExecutor("./_hdd", "dummy", func() Request {
			return &dummyRequest{}
		})
		c.So(err, ShouldBeNil)

		r := &dummyRequest{
			chunks: make(chan int32, 10),
			done:   nil,
		}
		for i := int32(0); i < 10; i++ {
			r.chunks <- i
		}
		err = e.Execute(r)
		c.So(err, ShouldBeNil)

		<-stopChan
	})
}
