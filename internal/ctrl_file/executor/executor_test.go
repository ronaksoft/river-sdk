package executor

import (
	"context"
	"encoding/json"
	"git.ronaksoft.com/river/sdk/internal/logs"
	"git.ronaksoft.com/river/sdk/internal/testenv"
	"github.com/ronaksoft/rony/tools"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/zap"
	"os"
	"sync"
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
	testenv.Log().Info("Do",
		zap.String("ReqID", d.req.GetID()),
		zap.Int32("ActionID", d.id),
	)
}

type dummyRequest struct {
	ID      string
	chunks  chan int32
	Done    []int32
	NextReq *dummyRequest
}

func (d *dummyRequest) GetID() string {
	return d.ID
}

func (d *dummyRequest) Prepare() error {
	d.chunks = make(chan int32, 10)
	d.Done = d.Done[:0]
	for i := int32(0); i < 10; i++ {
		d.chunks <- i
	}
	return nil
}

func (d *dummyRequest) NextAction() Action {
	select {
	case id := <-d.chunks:
		return &dummyAction{
			id:  id,
			req: d,
		}
	default:
		return nil
	}
}

func (d *dummyRequest) ActionDone(id int32) {
	d.Done = append(d.Done, id)
	if len(d.chunks) == 0 {
		testenv.Log().Info("Request Done", zap.String("ID", d.ID))
	}
}

func (d *dummyRequest) Serialize() []byte {
	b, err := json.Marshal(d)
	if err != nil {
		panic(err)
	}
	return b
}

func (d *dummyRequest) Next() Request {
	testenv.Log().Debug("Next", zap.Bool("Exists", d.NextReq != nil))
	if d.NextReq != nil {
		*d = *d.NextReq
		return d
	}
	return nil
}

func init() {
	logs.SetLogLevel(-1)
}

func TestNewExecutor(t *testing.T) {
	_ = os.MkdirAll("./_hdd", os.ModePerm)
	e, err := NewExecutor("./_hdd", "dummy",
		func(data []byte) Request {
			r := &dummyRequest{}
			err := json.Unmarshal(data, r)
			if err != nil {
				panic(err)
			}
			return r
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	Convey("Executor", t, func(c C) {
		Convey("Execute", func(c C) {
			r := &dummyRequest{
				NextReq: &dummyRequest{},
			}
			err = e.Execute(r)
			c.So(err, ShouldBeNil)

			time.Sleep(time.Second * 3)
		})
		Convey("ExecuteAndWait", func(c C) {
			r := &dummyRequest{
				ID:      tools.RandomID(32),
				NextReq: &dummyRequest{},
			}
			waitGroup := &sync.WaitGroup{}
			waitGroup.Add(1)
			err = e.ExecuteAndWait(waitGroup, r)
			c.So(err, ShouldBeNil)
			waitGroup.Wait()

		})

	})
}
