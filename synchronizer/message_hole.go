package synchronizer

import (
	"git.ronaksoftware.com/ronak/riversdk/log"
	"git.ronaksoftware.com/ronak/riversdk/repo"
	"go.uber.org/zap"
)

func IsMessageInHole(peerID, minID, maxID int64) bool {
	holes, err := repo.Ctx().MessageHoles.GetHoles(peerID, minID, maxID)
	if err != nil {
		return true
	}
	return len(holes) > 0
}

func fillMessageHoles(peerID, msgMinID, msgMaxID int64) error {
	holes, err := repo.Ctx().MessageHoles.GetHoles(peerID, msgMinID, msgMaxID)
	if err != nil {
		return err
	}
	for _, h := range holes {
		// inside or exact size of hole
		if h.MinID.Int64 <= msgMinID && h.MinID.Int64 < msgMaxID && h.MaxID > msgMinID && h.MaxID >= msgMaxID {

			err := repo.Ctx().MessageHoles.Delete(h.PeerID, h.MinID.Int64) // Delete
			if err != nil {
				fnLogFillMessageHoles("Delete", h.PeerID, h.MinID.Int64, h.MaxID, err)
			}
			err = repo.Ctx().MessageHoles.Save(h.PeerID, h.MinID.Int64, msgMinID-1) // Insert
			if err != nil {
				fnLogFillMessageHoles("Update", h.PeerID, h.MinID.Int64, msgMinID-1, err)
			}
			repo.Ctx().MessageHoles.Save(h.PeerID, msgMaxID+1, h.MaxID) // Insert
			if err != nil {
				fnLogFillMessageHoles("Insert", h.PeerID, msgMaxID+1, h.MaxID, err)
			}
		}
		// minside overlap
		if h.MinID.Int64 > msgMinID && h.MinID.Int64 < msgMaxID && h.MaxID > msgMinID && h.MaxID > msgMaxID {
			err := repo.Ctx().MessageHoles.Delete(h.PeerID, h.MinID.Int64) // Delete
			if err != nil {
				fnLogFillMessageHoles("Delete", h.PeerID, h.MinID.Int64, h.MaxID, err)
			}
			err = repo.Ctx().MessageHoles.Save(h.PeerID, msgMaxID+1, h.MaxID) // Insert
			if err != nil {
				fnLogFillMessageHoles("Insert", h.PeerID, msgMaxID+1, h.MaxID, err)
			}
		}
		// maxside overlap
		if h.MinID.Int64 < msgMinID && h.MinID.Int64 < msgMaxID && h.MaxID > msgMinID && h.MaxID < msgMaxID {
			err := repo.Ctx().MessageHoles.Save(h.PeerID, h.MinID.Int64, msgMinID-1) //Update
			if err != nil {
				fnLogFillMessageHoles("Update", h.PeerID, h.MinID.Int64, msgMinID-1, err)
			}
		}
		// surrendered over hole
		if h.MinID.Int64 > msgMinID && h.MinID.Int64 < msgMaxID && h.MaxID > msgMinID && h.MaxID < msgMaxID {
			err := repo.Ctx().MessageHoles.Delete(h.PeerID, h.MinID.Int64) // Delete
			if err != nil {
				fnLogFillMessageHoles("Delete", h.PeerID, h.MinID.Int64, h.MaxID, err)
			}
		}
	}
	return nil
}

func createMessageHole(peerID, minID, maxID int64) error {
	return repo.Ctx().MessageHoles.Save(peerID, minID, maxID)
}

func deleteMessageHole(peerID int64) error {
	return repo.Ctx().MessageHoles.DeleteAll(peerID)
}

func fnLogFillMessageHoles(operation string, peerID, minID, maxID int64, err error) {
	log.LOG_Warn("fillMessageHoles() :: Failed To "+operation,
		zap.Int64("peerID", peerID),
		zap.Int64("minID", minID),
		zap.Int64("maxID", maxID),
		zap.Error(err),
	)
}
