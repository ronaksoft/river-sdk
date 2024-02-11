package request_test

import (
    "sync"
    "testing"

    "github.com/ronaksoft/river-msg/go/msg"
    "github.com/ronaksoft/river-sdk/internal/request"
    "github.com/ronaksoft/rony"
    "github.com/ronaksoft/rony/tools"
    . "github.com/smartystreets/goconvey/convey"
)

/*
   Creation Time: 2021 - May - 27
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2020
*/

func TestNewCallback(t *testing.T) {
    Convey("Callback", t, func(c C) {

        reqID := tools.RandomUint64(0)
        Convey("Test OnComplete", func(c C) {
            waitGroup := &sync.WaitGroup{}
            waitGroup.Add(1)
            cb := request.NewCallback(
                0, 0, reqID,
                msg.C_TestRequest, &msg.TestRequest{Hash: tools.S2B(tools.RandomID(10))},
                func() {},
                func(m *rony.MessageEnvelope) {
                    c.So(m.Constructor, ShouldEqual, msg.C_TestResponse)
                    c.So(m.RequestID, ShouldEqual, reqID)
                    waitGroup.Done()
                },
                nil, true, 0, 0,
            )
            me := &rony.MessageEnvelope{}
            me.Fill(reqID, msg.C_TestResponse, &msg.TestResponse{})
            cb.OnComplete(me)
            waitGroup.Wait()
        })
        Convey("Test OnComplete with Replace", func(c C) {
            waitGroup := &sync.WaitGroup{}
            waitGroup.Add(1)
            cb := request.NewCallback(
                0, 0, reqID,
                msg.C_TestRequest, &msg.TestRequest{Hash: tools.S2B(tools.RandomID(10))},
                func() {},
                func(m *rony.MessageEnvelope) {
                    c.So(m.Constructor, ShouldEqual, msg.C_TestResponse)
                    c.So(m.RequestID, ShouldEqual, reqID+1)
                    waitGroup.Done()
                },
                nil, true, 0, 0,
            )
            me := &rony.MessageEnvelope{}
            me.Fill(reqID, msg.C_TestResponse, &msg.TestResponse{})
            cb.SetPreComplete(func(m *rony.MessageEnvelope) {
                m.RequestID = m.RequestID + 1
            })
            cb.OnComplete(me)
            waitGroup.Wait()
        })

    })
}
