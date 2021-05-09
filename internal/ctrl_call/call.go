/*
   Creation Time: 2021 - April - 04
   Created by:  (hamidrezakk)
   Maintainers:
      1.  HamidrezaKK (hamidrezakks@gmail.com)
   Auditor: HamidrezaKK
   Copyright Ronak Software Group 2021
*/

package callCtrl

import (
	"git.ronaksoft.com/river/msg/go/msg"
)

const (
	RetryInterval = 10000
	RetryLimit    = 6
)

type CallController interface {
}

func NewCallController() CallController {
	return &callController{
		peerConnections: nil,
		peer:            nil,
		activeCallID:    0,
		callInfo:        nil,
		callAPI:         nil,
	}
}

type callController struct {
	peerConnections map[int64]CallConnection
	peer            *msg.InputPeer
	activeCallID    int64
	callInfo        map[int64]CallInfo

	callAPI *CallAPI
}

func (c *callController) ToggleVide(enable bool) {

}

func (c *callController) ToggleAudio(enable bool) {

}

func (c *callController) TryReconnect(connId int32) {

}

func (c *callController) CallStart(peer *msg.InputPeer, participants []*msg.InputUser, callID int64) {
	c.peer = peer

}
