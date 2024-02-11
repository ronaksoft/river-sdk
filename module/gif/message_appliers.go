package gif

import (
    "github.com/ronaksoft/river-msg/go/msg"
    "github.com/ronaksoft/river-sdk/internal/domain"
    "github.com/ronaksoft/river-sdk/internal/repo"
    "github.com/ronaksoft/river-sdk/internal/uiexec"
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

func (r *gif) savedGifs(e *rony.MessageEnvelope) {
    u := &msg.SavedGifs{}
    err := u.Unmarshal(e.Message)
    if err != nil {
        r.Log().Error("couldn't unmarshal savedGifs", zap.Error(err))
        return
    }

    accessTime := domain.Now().Unix()
    for _, d := range u.Docs {
        err = repo.Files.SaveGif(d)
        if err != nil {
            r.Log().Warn("got error on applying SavedGifs (Save File)", zap.Error(err))
        }
        if !repo.Gifs.IsSaved(d.Doc.ClusterID, d.Doc.ID) {
            err = repo.Gifs.Save(d)
            if err != nil {
                r.Log().Warn("got error on applying SavedGifs (Save Gif)", zap.Error(err))
            }
            err = repo.Gifs.UpdateLastAccess(d.Doc.ClusterID, d.Doc.ID, accessTime)
            if err != nil {
                r.Log().Warn("got error on applying SavedGifs (Update Access Time)", zap.Error(err))
            }
        }
    }
    oldHash, _ := repo.System.LoadInt(domain.SkGifHash)
    err = repo.System.SaveInt(domain.SkGifHash, uint64(u.Hash))
    if err != nil {
        r.Log().Warn("got error on saving GifHash", zap.Error(err))
    }
    if oldHash != uint64(u.Hash) {
        uiexec.ExecDataSynced(false, false, true)
    }
}
