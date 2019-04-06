package synchronizer

import (
	"git.ronaksoftware.com/ronak/riversdk/logs"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo/dto"
	"go.uber.org/zap"
)

// GetHoles get holes between min & max
func GetHoles(peerID, minID, maxID int64) []dto.MessageHoles {
	holes, err := repo.Ctx().MessageHoles.GetHoles(peerID, minID, maxID)
	if err != nil {
		return make([]dto.MessageHoles, 0)
	}
	return holes
}

// GetMinClosestHole find closest hole from lower side
func GetMinClosestHole(minID int64, holes []dto.MessageHoles) *dto.MessageHoles {
	minGapSizeIdx := -1
	minGaSize := int64(^uint64(0) >> 1)

	for idx, h := range holes {
		if h.MaxID < minID {
			continue
		}
		gapSize := h.MinID.Int64 - minID
		if minGaSize > gapSize {
			minGaSize = gapSize
			minGapSizeIdx = idx
		}
	}
	if minGapSizeIdx > -1 {
		return &holes[minGapSizeIdx]
	}
	return nil
}

// GetMaxClosestHole find closest hole from upper side
func GetMaxClosestHole(maxID int64, holes []dto.MessageHoles) *dto.MessageHoles {
	maxGapSizeIdx := -1
	maxGapSize := int64(^uint64(0) >> 1)
	for idx, h := range holes {
		if h.MaxID > maxID {
			continue
		}
		gapSize := maxID - h.MaxID
		if maxGapSize > gapSize {
			maxGapSize = gapSize
			maxGapSizeIdx = idx
		}
	}
	if maxGapSizeIdx > -1 {
		return &holes[maxGapSizeIdx]
	}
	return nil
}

// fillMessageHoles
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
				fnLogFillMessageHoles("Insert", h.PeerID, h.MinID.Int64, msgMinID-1, err)
			}
			err = repo.Ctx().MessageHoles.Save(h.PeerID, msgMaxID+1, h.MaxID) // Insert
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

// CreateMessageHole
func CreateMessageHole(peerID, minID, maxID int64) error {
	return repo.Ctx().MessageHoles.Save(peerID, minID, maxID)
}

// DeleteMessageHole
func DeleteMessageHole(peerID int64) error {
	return repo.Ctx().MessageHoles.DeleteAll(peerID)
}

// fnLogFillMessageHoles
func fnLogFillMessageHoles(operation string, peerID, minID, maxID int64, err error) {
	logs.Warn("fillMessageHoles() :: Failed To "+operation,
		zap.Int64("peerID", peerID),
		zap.Int64("minID", minID),
		zap.Int64("maxID", maxID),
		zap.Error(err),
	)
}
