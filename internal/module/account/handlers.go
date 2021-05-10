package account

import (
	"encoding/json"
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/internal/logs"
	"git.ronaksoft.com/river/sdk/internal/repo"
	"github.com/ronaksoft/rony"
	"go.uber.org/zap"
)

/*
   Creation Time: 2021 - May - 10
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2020
*/

func (r *account) accountUpdateUsername(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.AccountUpdateUsername{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	r.sdk.GetConnInfo().ChangeUsername(req.Username)
	r.sdk.GetConnInfo().Save()

	// send the request to server
	r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)
}

func (r *account) accountRegisterDevice(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.AccountRegisterDevice{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	val, err := json.Marshal(req)
	if err != nil {
		logs.Error("River::accountRegisterDevice()-> Json Marshal()", zap.Error(err))
		return
	}
	err = repo.System.SaveString(domain.SkDeviceToken, string(val))
	if err != nil {
		logs.Error("River::accountRegisterDevice()-> SaveString()", zap.Error(err))
		return
	}
	// send the request to server
	r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)
}

func (r *account) accountUnregisterDevice(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.AccountUnregisterDevice{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "E00", Items: err.Error()})
		successCB(out)
		return
	}

	val, err := json.Marshal(&msg.AccountRegisterDevice{})
	if err != nil {
		logs.Error("River::accountUnregisterDevice()-> Json Marshal()", zap.Error(err))
		return
	}
	err = repo.System.SaveString(domain.SkDeviceToken, string(val))
	if err != nil {
		logs.Error("River::accountUnregisterDevice()-> SaveString()", zap.Error(err))
		return
	}

	// send the request to server
	r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)
}

func (r *account) accountSetNotifySettings(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.AccountSetNotifySettings{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	dialog, _ := repo.Dialogs.Get(domain.GetTeamID(in), req.Peer.ID, int32(req.Peer.Type))
	if dialog == nil {
		return
	}

	dialog.NotifySettings = req.Settings
	_ = repo.Dialogs.Save(dialog)

	// send the request to server
	r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)

}

func (r *account) accountRemovePhoto(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	x := &msg.AccountRemovePhoto{}
	_ = x.Unmarshal(in.Message)

	// send the request to server
	r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)

	user, err := repo.Users.Get(r.sdk.GetConnInfo().PickupUserID())
	if err != nil {
		return
	}

	if user.Photo != nil && user.Photo.PhotoID == x.PhotoID {
		_ = repo.Users.UpdatePhoto(r.sdk.GetConnInfo().PickupUserID(), &msg.UserPhoto{
			PhotoBig:      &msg.FileLocation{},
			PhotoSmall:    &msg.FileLocation{},
			PhotoBigWeb:   &msg.WebLocation{},
			PhotoSmallWeb: &msg.WebLocation{},
			PhotoID:       0,
		})
	}

	repo.Users.RemovePhotoGallery(r.sdk.GetConnInfo().PickupUserID(), x.PhotoID)
}

func (r *account) accountUpdateProfile(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.AccountUpdateProfile{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	// TODO : add connInfo Bio and save it too
	r.sdk.GetConnInfo().ChangeFirstName(req.FirstName)
	r.sdk.GetConnInfo().ChangeLastName(req.LastName)
	r.sdk.GetConnInfo().ChangeBio(req.Bio)

	r.sdk.GetConnInfo().Save()

	_ = repo.Users.UpdateProfile(r.sdk.GetConnInfo().PickupUserID(),
		req.FirstName, req.LastName, r.sdk.GetConnInfo().PickupUsername(), req.Bio, r.sdk.GetConnInfo().PickupPhone(),
	)

	// send the request to server
	r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)
}

func (r *account) accountsGetTeams(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.AccountGetTeams{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	teams := repo.Teams.List()

	if len(teams) > 0 {
		teamsMany := &msg.TeamsMany{
			Teams: teams,
		}
		out.Fill(out.RequestID, msg.C_TeamsMany, teamsMany)
		successCB(out)
		return
	}

	r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)
}
