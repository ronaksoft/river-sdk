package call

import (
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/request"
	"github.com/ronaksoft/rony"
)

/*
   Creation Time: 2021 - May - 19
   Created by:  (Hamidrezakk)
   Maintainers:
      1.  Hamidrezakk
   Auditor: Hamidrezakk
   Copyright Ronak Software Group 2021
*/

func (c *call) toggleVideoHandler(da request.Callback) {
	req := &msg.ClientCallToggleVideo{}
	if err := da.RequestData(req); err != nil {
		return
	}

	err := c.toggleVideo(req.Video)
	if err != nil {
		da.Response(rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		return
	}

	da.Response(msg.C_Bool, &msg.Bool{Result: true})
}

func (c *call) toggleAudioHandler(da request.Callback) {
	req := &msg.ClientCallToggleAudio{}
	if err := da.RequestData(req); err != nil {
		return
	}

	err := c.toggleAudio(req.Audio)
	if err != nil {
		da.Response(rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		return
	}

	da.Response(msg.C_Bool, &msg.Bool{Result: true})
}

func (c *call) tryReconnectHandler(da request.Callback) {
	req := &msg.ClientCallTryReconnect{}
	if err := da.RequestData(req); err != nil {
		return
	}

	err := c.tryReconnect(req.ConnId)
	if err != nil {
		da.Response(rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		return
	}

	da.Response(msg.C_Bool, &msg.Bool{Result: true})
}

func (c *call) destroyHandler(da request.Callback) {
	req := &msg.ClientCallDestroy{}
	if err := da.RequestData(req); err != nil {
		return
	}

	c.destroy(req.CallID)

	da.Response(msg.C_Bool, &msg.Bool{Result: true})
}

func (c *call) areAllAudioHandler(da request.Callback) {
	req := &msg.ClientCallAreAllAudio{}
	if err := da.RequestData(req); err != nil {
		return
	}

	ok, err := c.areAllAudio()
	if err != nil {
		da.Response(rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		return
	}

	da.Response(msg.C_Bool, &msg.Bool{Result: ok})
}

func (c *call) iceCandidateHandler(da request.Callback) {
	req := &msg.ClientCallSendIceCandidate{}
	if err := da.RequestData(req); err != nil {
		return
	}

	err := c.iceCandidate(req.ConnId, req.Candidate)
	if err != nil {
		da.Response(rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		return
	}

	da.Response(msg.C_Bool, &msg.Bool{Result: true})
}

func (c *call) iceConnectionStatusChangeHandler(da request.Callback) {
	req := &msg.ClientCallSendIceConnectionStatus{}
	if err := da.RequestData(req); err != nil {
		return
	}

	err := c.iceConnectionStatusChange(req.ConnId, req.State, req.HasIceError)
	if err != nil {
		da.Response(rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		return
	}

	da.Response(msg.C_Bool, &msg.Bool{Result: true})
}

func (c *call) trackUpdateHandler(da request.Callback) {
	req := &msg.ClientCallSendTrack{}
	if err := da.RequestData(req); err != nil {
		return
	}

	// TODO: (@hamidrezakk)
	err := c.trackUpdate(req.ConnId, req.StreamID)
	if err = da.RequestData(req); err != nil {
		return
	}

	da.Response(msg.C_Bool, &msg.Bool{Result: true})
}

func (c *call) mediaSettingsChangeHandler(da request.Callback) {
	req := &msg.ClientCallSendMediaSettings{}
	if err := da.RequestData(req); err != nil {
		return
	}

	err := c.mediaSettingsChange(req.MediaSettings)
	if err != nil {
		da.Response(rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		return
	}

	da.Response(msg.C_Bool, &msg.Bool{Result: true})
}

func (c *call) startHandler(da request.Callback) {
	req := &msg.ClientCallStart{}
	if err := da.RequestData(req); err != nil {
		return
	}

	callID, err := c.start(req.Peer, req.InputUsers, req.Video, req.CallID)
	if err != nil {
		da.Response(rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		return
	}

	da.Response(msg.C_ClientCallStarted, &msg.ClientCallStarted{CallID: callID})
}

func (c *call) joinHandler(da request.Callback) {
	req := &msg.ClientCallJoin{}
	if err := da.RequestData(req); err != nil {
		return
	}

	err := c.join(req.Peer, req.CallID)
	if err != nil {
		da.Response(rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		return
	}

	da.Response(msg.C_Bool, &msg.Bool{Result: true})
}

func (c *call) acceptHandler(da request.Callback) {
	req := &msg.ClientCallAccept{}
	if err := da.RequestData(req); err != nil {
		return
	}

	err := c.accept(req.CallID, req.Video)
	if err != nil {
		da.Response(rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		return
	}

	da.Response(msg.C_Bool, &msg.Bool{Result: true})
}

func (c *call) rejectHandler(da request.Callback) {
	req := &msg.ClientCallReject{}
	if err := da.RequestData(req); err != nil {
		return
	}

	err := c.reject(req.CallID, req.Duration, req.Reason, req.TargetPeer)
	if err != nil {
		da.Response(rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		return
	}

	da.Response(msg.C_Bool, &msg.Bool{Result: true})
}

func (c *call) getParticipantByUserIDHandler(da request.Callback) {
	req := &msg.ClientCallGetParticipantByUserID{}
	if err := da.RequestData(req); err != nil {
		return
	}

	participant, err := c.getParticipantByUserID(req.CallID, req.UserID)
	if err != nil {
		da.Response(rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		return
	}

	da.Response(msg.C_CallParticipant, participant)
}

func (c *call) getParticipantByConnIdHandler(da request.Callback) {
	req := &msg.ClientCallGetParticipantByConnId{}
	if err := da.RequestData(req); err != nil {
		return
	}

	participant, err := c.getParticipantByConnId(req.ConnId)
	if err != nil {
		da.Response(rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		return
	}

	da.Response(msg.C_CallParticipant, participant)
}

func (c *call) getParticipantListHandler(da request.Callback) {
	req := &msg.ClientCallGetParticipantList{}
	if err := da.RequestData(req); err != nil {
		return
	}

	participants, err := c.getParticipantList(req.CallID, req.ExcludeCurrent)
	if err != nil {
		da.Response(rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		return
	}

	da.Response(msg.C_CallParticipants, &msg.CallParticipants{
		CallParticipants: participants,
	})
}

func (c *call) muteParticipantHandler(da request.Callback) {
	req := &msg.ClientCallMuteParticipant{}
	if err := da.RequestData(req); err != nil {
		return
	}

	err := c.muteParticipant(req.UserID, req.Muted)
	if err != nil {
		da.Response(rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		return
	}

	da.Response(msg.C_Bool, &msg.Bool{Result: true})
}

func (c *call) groupAddParticipantHandler(da request.Callback) {
	req := &msg.ClientCallGroupAddParticipant{}
	if err := da.RequestData(req); err != nil {
		return
	}

	err := c.groupAddParticipant(req.CallID, req.Participants)
	if err != nil {
		da.Response(rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		return
	}

	da.Response(msg.C_Bool, &msg.Bool{Result: true})
}

func (c *call) groupRemoveParticipantHandler(da request.Callback) {
	req := &msg.ClientCallGroupRemoveParticipant{}
	if err := da.RequestData(req); err != nil {
		return
	}

	err := c.groupRemoveParticipant(req.CallID, req.UserIDs, req.Timeout)
	if err != nil {
		da.Response(rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		return
	}

	da.Response(msg.C_Bool, &msg.Bool{Result: true})
}

func (c *call) groupUpdateAdminHandler(da request.Callback) {
	req := &msg.ClientCallGroupUpdateAdmin{}
	if err := da.RequestData(req); err != nil {
		return
	}

	err := c.groupUpdateAdmin(req.CallID, req.UserID, req.Admin)
	if err != nil {
		da.Response(rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		return
	}

	da.Response(msg.C_Bool, &msg.Bool{Result: true})
}
