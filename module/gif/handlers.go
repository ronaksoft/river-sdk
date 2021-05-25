package gif

import (
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/internal/repo"
	"git.ronaksoft.com/river/sdk/internal/request"
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

func (r *gif) gifSave(in, out *rony.MessageEnvelope, da request.Callback) {
	req := &msg.GifSave{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}

	cf, err := repo.Files.Get(req.Doc.ClusterID, req.Doc.ID, req.Doc.AccessHash)
	if err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}

	r.Log().Info("are saving GIF",
		zap.Int64("FileID", cf.FileID),
		zap.Uint64("AccessHash", cf.AccessHash),
		zap.Int32("ClusterID", cf.ClusterID),
	)
	if !repo.Gifs.IsSaved(cf.ClusterID, cf.FileID) {
		md := &msg.MediaDocument{
			Doc: &msg.Document{
				ID:          cf.FileID,
				AccessHash:  cf.AccessHash,
				Date:        0,
				MimeType:    cf.MimeType,
				FileSize:    int32(cf.FileSize),
				Version:     cf.Version,
				ClusterID:   cf.ClusterID,
				Attributes:  req.Attributes,
				MD5Checksum: cf.MD5Checksum,
			},
		}
		err = repo.Gifs.Save(md)
		if err != nil {
			out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
			da.OnComplete(out)
			return
		}
	}
	_ = repo.Gifs.UpdateLastAccess(cf.ClusterID, cf.FileID, domain.Now().Unix())

	r.SDK().QueueCtrl().EnqueueCommand(in, da.OnTimeout, da.OnComplete, da.UI())
}

func (r *gif) gifDelete(in, out *rony.MessageEnvelope, da request.Callback) {
	req := &msg.GifDelete{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}

	err := repo.Gifs.Delete(req.Doc.ClusterID, req.Doc.ID)
	if err != nil {
		r.Log().Warn("got error on deleting GIF document", zap.Error(err))
	}

	r.SDK().QueueCtrl().EnqueueCommand(in, da.OnTimeout, da.OnComplete, da.UI())
}

func (r *gif) gifGetSaved(in, out *rony.MessageEnvelope, da request.Callback) {
	req := &msg.GifGetSaved{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}
	gifHash, _ := repo.System.LoadInt(domain.SkGifHash)

	var enqueueSuccessCB domain.MessageHandler

	if gifHash != 0 {
		res, err := repo.Gifs.GetSaved()
		if err != nil {
			out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
			da.OnComplete(out)
			return
		}
		out.Fill(out.RequestID, msg.C_SavedGifs, res)
		da.OnComplete(out)

		// ignore success cb because we notify views on message hanlder
		enqueueSuccessCB = func(m *rony.MessageEnvelope) {

		}
	} else {
		enqueueSuccessCB = da.OnComplete
	}

	r.SDK().QueueCtrl().EnqueueCommand(in, da.OnTimeout, enqueueSuccessCB, true)
}
