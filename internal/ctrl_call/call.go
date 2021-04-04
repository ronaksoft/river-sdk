/*
   Creation Time: 2021 - April - 04
   Created by:  (hamidrezakk)
   Maintainers:
      1.  HamidrezaKK (hamidrezakks@gmail.com)
   Auditor: HamidrezaKK
   Copyright Ronak Software Group 2021
*/

package callCtrl

import "git.ronaksoft.com/river/msg/go/msg"

const (
	RetryInterval = 10000
	RetryLimit    = 6
)

type UpdatePhoneCall struct {
	msg.UpdatePhoneCall
	Data interface{}
}

type MediaSettings struct {
	Audio       bool
	ScreenShare bool
	Video       bool
}

type CallParticipant struct {
	msg.PhoneParticipant
	DeviceType msg.CallDeviceType
	MediaSettings
}

type RTCIceCandidate struct {
	Candidate        string
	SDPMid           string
	SDPMLineIndex    int64
	UsernameFragment string
}

type CallConnection struct {
	Accepted bool
	IceQueue []RTCIceCandidate
	Interval interface{}
	Try      int64
}

type callController struct {
}

func (c *callController) ToggleVide(enable bool) {

}

func (c *callController) ToggleAudio(enable bool) {

}
