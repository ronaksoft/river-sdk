package gif

import (
    "github.com/ronaksoft/river-msg/go/msg"
    "github.com/ronaksoft/river-sdk/internal/domain"
    "github.com/ronaksoft/river-sdk/internal/repo"
    "github.com/ronaksoft/river-sdk/internal/request"
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

func (r *gif) gifSave(da request.Callback) {
    req := &msg.GifSave{}
    if err := da.RequestData(req); err != nil {
        return
    }

    cf, err := repo.Files.Get(req.Doc.ClusterID, req.Doc.ID, req.Doc.AccessHash)
    if err != nil {
        da.Response(rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
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
            da.Response(rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
            return
        }
    }
    _ = repo.Gifs.UpdateLastAccess(cf.ClusterID, cf.FileID, domain.Now().Unix())

    r.SDK().QueueCtrl().EnqueueCommand(da)
}

func (r *gif) gifDelete(da request.Callback) {
    req := &msg.GifDelete{}
    if err := da.RequestData(req); err != nil {
        return
    }

    err := repo.Gifs.Delete(req.Doc.ClusterID, req.Doc.ID)
    if err != nil {
        r.Log().Warn("got error on deleting GIF document", zap.Error(err))
    }

    r.SDK().QueueCtrl().EnqueueCommand(da)
}

func (r *gif) gifGetSaved(da request.Callback) {
    req := &msg.GifGetSaved{}
    if err := da.RequestData(req); err != nil {
        return
    }

    gifHash, _ := repo.System.LoadInt(domain.SkGifHash)

    if gifHash != 0 {
        res, err := repo.Gifs.GetSaved()
        if err != nil {
            da.Response(rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
            return
        }
        da.Response(msg.C_SavedGifs, res)
    }

    // TODO:: set ui to false
    r.SDK().QueueCtrl().EnqueueCommand(da)
}
