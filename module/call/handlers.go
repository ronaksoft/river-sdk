package call

import (
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"github.com/ronaksoft/rony"
)

func (c *call) toggleVideoHandler(in, out *rony.MessageEnvelope, da domain.Callback) {
	req := &msg.ClientCallToggleVideo{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}

	err := c.toggleVideo(req.Video)
	if err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}

	out.Fill(out.RequestID, msg.C_Bool, &msg.Bool{Result: true})
	da.OnComplete(out)
}

func (c *call) toggleAudioHandler(in, out *rony.MessageEnvelope, da domain.Callback) {
	req := &msg.ClientCallToggleAudio{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}

	err := c.toggleAudio(req.Audio)
	if err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}

	out.Fill(out.RequestID, msg.C_Bool, &msg.Bool{Result: true})
	da.OnComplete(out)
}

func (c *call) tryReconnectHandler(in, out *rony.MessageEnvelope, da domain.Callback) {
	req := &msg.ClientCallTryReconnect{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}

	err := c.tryReconnect(req.ConnId)
	if err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}

	out.Fill(out.RequestID, msg.C_Bool, &msg.Bool{Result: true})
	da.OnComplete(out)
}

func (c *call) destroyHandler(in, out *rony.MessageEnvelope, da domain.Callback) {
	req := &msg.ClientCallDestroy{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}

	c.destroy(req.CallID)

	out.Fill(out.RequestID, msg.C_Bool, &msg.Bool{Result: true})
	da.OnComplete(out)
}

func (c *call) areAllAudioHandler(in, out *rony.MessageEnvelope, da domain.Callback) {
	req := &msg.ClientCallDestroy{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}

	ok, err := c.areAllAudio()
	if err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}

	out.Fill(out.RequestID, msg.C_Bool, &msg.Bool{Result: ok})
	da.OnComplete(out)
}

func (c *call) iceCandidateHandler(in, out *rony.MessageEnvelope, da domain.Callback) {
	req := &msg.ClientCallSendIceCandidate{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}

	err := c.iceCandidate(req.ConnId, req.Candidate)
	if err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}

	out.Fill(out.RequestID, msg.C_Bool, &msg.Bool{Result: true})
	da.OnComplete(out)
}

func (c *call) iceConnectionStatusChangeHandler(in, out *rony.MessageEnvelope, da domain.Callback) {
	req := &msg.ClientCallSendIceConnectionStatus{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}

	err := c.iceConnectionStatusChange(req.ConnId, req.State, req.HasIceError)
	if err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}

	out.Fill(out.RequestID, msg.C_Bool, &msg.Bool{Result: true})
	da.OnComplete(out)
}

func (c *call) mediaSettingsChangeHandler(in, out *rony.MessageEnvelope, da domain.Callback) {
	req := &msg.ClientCallSendMediaSettings{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}

	err := c.mediaSettingsChange(req.ConnId, req.MediaSettings)
	if err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}

	out.Fill(out.RequestID, msg.C_Bool, &msg.Bool{Result: true})
	da.OnComplete(out)
}

func (c *call) startHandler(in, out *rony.MessageEnvelope, da domain.Callback) {
	req := &msg.ClientCallStart{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}

	callID, err := c.start(req.Peer, req.InputUsers, req.CallID)
	if err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}

	out.Fill(out.RequestID, msg.C_ClientCallStarted, &msg.ClientCallStarted{CallID: callID})
	da.OnComplete(out)
}

func (c *call) joinHandler(in, out *rony.MessageEnvelope, da domain.Callback) {
	req := &msg.ClientCallJoin{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}

	err := c.join(req.Peer, req.CallID)
	if err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}

	out.Fill(out.RequestID, msg.C_Bool, &msg.Bool{Result: true})
	da.OnComplete(out)
}

func (c *call) acceptHandler(in, out *rony.MessageEnvelope, da domain.Callback) {
	req := &msg.ClientCallAccept{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}

	err := c.accept(req.CallID, req.Video)
	if err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}

	out.Fill(out.RequestID, msg.C_Bool, &msg.Bool{Result: true})
	da.OnComplete(out)
}

func (c *call) rejectHandler(in, out *rony.MessageEnvelope, da domain.Callback) {
	req := &msg.ClientCallReject{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}

	err := c.reject(req.CallID, req.Duration, req.Reason, req.TargetPeer)
	if err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}

	out.Fill(out.RequestID, msg.C_Bool, &msg.Bool{Result: true})
	da.OnComplete(out)
}

func (c *call) groupGetParticipantByUserIDHandler(in, out *rony.MessageEnvelope, da domain.Callback) {
	req := &msg.ClientCallGetParticipantByUserID{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}

	participant, err := c.getParticipantByUserID(req.CallID, req.UserID)
	if err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}

	out.Fill(out.RequestID, msg.C_CallParticipant, participant)
	da.OnComplete(out)
}

func (c *call) groupGetParticipantByConnIdHandler(in, out *rony.MessageEnvelope, da domain.Callback) {
	req := &msg.ClientCallGetParticipantByConnId{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}

	participant, err := c.getParticipantByConnId(req.ConnId)
	if err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}

	out.Fill(out.RequestID, msg.C_CallParticipant, participant)
	da.OnComplete(out)
}

func (c *call) groupGetParticipantListHandler(in, out *rony.MessageEnvelope, da domain.Callback) {
	req := &msg.ClientCallGetParticipantList{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}

	participants, err := c.getParticipantList(req.CallID, req.ExcludeCurrent)
	if err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}

	out.Fill(out.RequestID, msg.C_CallParticipants, &msg.CallParticipants{
		CallParticipants: participants,
	})
	da.OnComplete(out)
}

func (c *call) groupMuteParticipantHandler(in, out *rony.MessageEnvelope, da domain.Callback) {
	req := &msg.ClientCallMuteParticipant{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}

	err := c.muteParticipant(req.UserID, req.Muted)
	if err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}

	out.Fill(out.RequestID, msg.C_Bool, &msg.Bool{Result: true})
	da.OnComplete(out)
}

func (c *call) groupAddParticipantHandler(in, out *rony.MessageEnvelope, da domain.Callback) {
	req := &msg.ClientCallGroupAddParticipant{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}

	err := c.groupAddParticipant(req.CallID, req.Participants)
	if err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}

	out.Fill(out.RequestID, msg.C_Bool, &msg.Bool{Result: true})
	da.OnComplete(out)
}

func (c *call) groupRemoveParticipantHandler(in, out *rony.MessageEnvelope, da domain.Callback) {
	req := &msg.ClientCallGroupRemoveParticipant{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}

	err := c.groupRemoveParticipant(req.CallID, req.UserIDs, req.Timeout)
	if err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}

	out.Fill(out.RequestID, msg.C_Bool, &msg.Bool{Result: true})
	da.OnComplete(out)
}

func (c *call) groupUpdateAdminHandler(in, out *rony.MessageEnvelope, da domain.Callback) {
	req := &msg.ClientCallGroupUpdateAdmin{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}

	err := c.groupUpdateAdmin(req.CallID, req.UserID, req.Admin)
	if err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}

	out.Fill(out.RequestID, msg.C_Bool, &msg.Bool{Result: true})
	da.OnComplete(out)
}
