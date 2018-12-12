package synchronizer

import (
	"git.ronaksoftware.com/ronak/riversdk/log"
	"git.ronaksoftware.com/ronak/riversdk/repo"
	"go.uber.org/zap"
)

func isMessageInHole(peerID, minID, maxID int64) bool {
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
		if h.MinID <= msgMinID && h.MaxID >= msgMaxID {

			err := repo.Ctx().MessageHoles.Delete(h.PeerID, h.MinID) // Delete
			if err != nil {
				fnLogFillMessageHoles("Delete", h.PeerID, h.MinID, h.MaxID)
			}
			err = repo.Ctx().MessageHoles.Save(h.PeerID, h.MinID, msgMinID-1) // Insert
			if err != nil {
				fnLogFillMessageHoles("Update", h.PeerID, h.MinID, msgMinID-1)
			}
			repo.Ctx().MessageHoles.Save(h.PeerID, msgMaxID+1, h.MaxID) // Insert
			if err != nil {
				fnLogFillMessageHoles("Insert", h.PeerID, msgMaxID+1, h.MaxID)
			}
		}
		// minside overlap
		if h.MinID > msgMinID && h.MaxID > msgMaxID {
			err := repo.Ctx().MessageHoles.Delete(h.PeerID, h.MinID) // Delete
			if err != nil {
				fnLogFillMessageHoles("Delete", h.PeerID, h.MinID, h.MaxID)
			}
			err = repo.Ctx().MessageHoles.Save(h.PeerID, msgMaxID+1, h.MaxID) // Insert
			if err != nil {
				fnLogFillMessageHoles("Insert", h.PeerID, msgMaxID+1, h.MaxID)
			}
		}
		// maxside overlap
		if h.MinID < msgMinID && h.MaxID < msgMaxID {
			repo.Ctx().MessageHoles.Save(h.PeerID, h.MinID, msgMaxID-1) //Update
			fnLogFillMessageHoles("Update", h.PeerID, h.MinID, msgMaxID-1)
		}
		// surrendered over hole
		if h.MinID > msgMinID && h.MaxID < msgMaxID {
			repo.Ctx().MessageHoles.Delete(h.PeerID, h.MinID) // Delete
			fnLogFillMessageHoles("Delete", h.PeerID, h.MinID, h.MaxID)
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

func fnLogFillMessageHoles(operation string, peerID, minID, maxID int64) {
	log.LOG_Warn("fillMessageHoles() :: Failed To "+operation,
		zap.Int64("peerID", peerID),
		zap.Int64("minID", minID),
		zap.Int64("maxID", maxID),
	)
}
