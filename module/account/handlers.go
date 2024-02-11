package account

import (
    "encoding/json"

    "github.com/ronaksoft/river-msg/go/msg"
    "github.com/ronaksoft/river-sdk/internal/domain"
    "github.com/ronaksoft/river-sdk/internal/repo"
    "github.com/ronaksoft/river-sdk/internal/request"
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

func (r *account) accountUpdateUsername(da request.Callback) {
    req := &msg.AccountUpdateUsername{}
    if err := da.RequestData(req); err != nil {
        return
    }

    r.SDK().GetConnInfo().ChangeUsername(req.Username)
    r.SDK().GetConnInfo().Save()

    // send the request to server
    r.SDK().QueueCtrl().EnqueueCommand(da)
}

func (r *account) accountRegisterDevice(da request.Callback) {
    req := &msg.AccountRegisterDevice{}
    if err := da.RequestData(req); err != nil {
        return
    }

    val, err := json.Marshal(req)
    if err != nil {
        r.Log().Error("AccountModule::accountRegisterDevice()-> Json Marshal()", zap.Error(err))
        return
    }
    err = repo.System.SaveString(domain.SkDeviceToken, string(val))
    if err != nil {
        r.Log().Error("AccountModule::accountRegisterDevice()-> SaveString()", zap.Error(err))
        return
    }
    // send the request to server
    r.SDK().QueueCtrl().EnqueueCommand(da)
}

func (r *account) accountUnregisterDevice(da request.Callback) {
    req := &msg.AccountUnregisterDevice{}
    if err := da.RequestData(req); err != nil {
        return
    }

    val, err := json.Marshal(&msg.AccountRegisterDevice{})
    if err != nil {
        r.Log().Error("AccountModule::accountUnregisterDevice()-> Json Marshal()", zap.Error(err))
        return
    }
    err = repo.System.SaveString(domain.SkDeviceToken, string(val))
    if err != nil {
        r.Log().Error("AccountModule::accountUnregisterDevice()-> SaveString()", zap.Error(err))
        return
    }

    // send the request to server
    r.SDK().QueueCtrl().EnqueueCommand(da)
}

func (r *account) accountSetNotifySettings(da request.Callback) {
    req := &msg.AccountSetNotifySettings{}
    if err := da.RequestData(req); err != nil {
        return
    }

    dialog, _ := repo.Dialogs.Get(da.TeamID(), req.Peer.ID, int32(req.Peer.Type))
    if dialog == nil {
        return
    }

    dialog.NotifySettings = req.Settings
    _ = repo.Dialogs.Save(dialog)

    // send the request to server
    r.SDK().QueueCtrl().EnqueueCommand(da)

}

func (r *account) accountRemovePhoto(da request.Callback) {
    req := &msg.AccountRemovePhoto{}
    if err := da.RequestData(req); err != nil {
        return
    }

    // send the request to server
    r.SDK().QueueCtrl().EnqueueCommand(da)

    user, err := repo.Users.Get(r.SDK().GetConnInfo().PickupUserID())
    if err != nil {
        return
    }

    if user.Photo != nil && user.Photo.PhotoID == req.PhotoID {
        _ = repo.Users.UpdatePhoto(r.SDK().GetConnInfo().PickupUserID(), &msg.UserPhoto{
            PhotoBig:      &msg.FileLocation{},
            PhotoSmall:    &msg.FileLocation{},
            PhotoBigWeb:   &msg.WebLocation{},
            PhotoSmallWeb: &msg.WebLocation{},
            PhotoID:       0,
        })
    }

    _ = repo.Users.RemovePhotoGallery(r.SDK().GetConnInfo().PickupUserID(), req.PhotoID)
}

func (r *account) accountUpdateProfile(da request.Callback) {
    req := &msg.AccountUpdateProfile{}
    if err := da.RequestData(req); err != nil {
        return
    }

    // TODO : add connInfo Bio and save it too
    r.SDK().GetConnInfo().ChangeFirstName(req.FirstName)
    r.SDK().GetConnInfo().ChangeLastName(req.LastName)
    r.SDK().GetConnInfo().ChangeBio(req.Bio)

    r.SDK().GetConnInfo().Save()

    _ = repo.Users.UpdateProfile(r.SDK().GetConnInfo().PickupUserID(),
        req.FirstName, req.LastName, r.SDK().GetConnInfo().PickupUsername(), req.Bio, r.SDK().GetConnInfo().PickupPhone(),
    )

    // send the request to server
    r.SDK().QueueCtrl().EnqueueCommand(da)
}

func (r *account) accountsGetTeams(da request.Callback) {
    req := &msg.AccountGetTeams{}
    if err := da.RequestData(req); err != nil {
        return
    }

    teams := repo.Teams.List()

    if len(teams) > 0 {
        teamsMany := &msg.TeamsMany{
            Teams: teams,
        }
        da.Response(msg.C_TeamsMany, teamsMany)
        return
    }

    r.SDK().QueueCtrl().EnqueueCommand(da)
}
