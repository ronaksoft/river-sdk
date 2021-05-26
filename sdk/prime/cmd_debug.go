package riversdk

import (
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/internal/repo"
	"git.ronaksoft.com/river/sdk/internal/request"
	"git.ronaksoft.com/river/sdk/module"
	"github.com/ronaksoft/rony"
	"github.com/ronaksoft/rony/tools"
	"go.uber.org/zap"
)

/*
   Creation Time: 2019 - Jul - 28
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2018
*/

func (r *River) HandleDebugActions(txt string) {
	req := &msg.MessagesSend{
		RandomID:   tools.RandomInt64(0),
		Peer:       &msg.InputPeer{ID: r.ConnInfo.UserID},
		Body:       txt,
		ReplyTo:    0,
		ClearDraft: false,
		Entities:   nil,
	}
	in := &rony.MessageEnvelope{}
	in.Fill(domain.NextRequestID(), msg.C_MessagesSend, req)
	r.Module(module.Message).Execute(
		request.NewCallback(
			0, 0, domain.NextRequestID(), msg.C_MessagesSend, req,
			nil, nil, nil, false, 0, 0,
		),
	)
}

func (r *River) GetHole(peerID int64, peerType int32) []byte {
	return repo.MessagesExtra.GetHoles(domain.GetCurrTeamID(), peerID, peerType, 0)
}

func (r *River) CancelFileRequest(reqID string) {
	err := repo.Files.DeleteFileRequest(reqID)
	if err != nil {
		logger.Warn("got error on delete file request", zap.Error(err))
	}
}

func (r *River) DeleteAllPendingMessages() {
	for _, p := range repo.PendingMessages.GetAll() {
		if p.FileID != 0 {
			r.fileCtrl.CancelUploadRequest(p.FileID)
		}
		_ = repo.PendingMessages.Delete(p.ID)
	}
}

func (r *River) SetUpdateState(newUpdateID int64) {
	_ = r.syncCtrl.SetUpdateID(newUpdateID)
	go r.syncCtrl.Sync()
}

func (r *River) GetUpdateState() int64 {
	return r.syncCtrl.GetUpdateID()
}
