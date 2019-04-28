package synchronizer

import (
	"encoding/json"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo/dto"
	"go.uber.org/zap"
	"sort"
)

type PointType int

const (
	_ PointType = iota
	HoleStart
	HoleStop
	FillStart
	FillStop
)

func (v PointType) String() string {
	switch v {
	case HoleStart:
		return "HoleStart"
	case HoleStop:
		return "HoleStop"
	case FillStart:
		return "FillStart"
	case FillStop:
		return "FillStop"
	}
	panic("invalid point type")
}

type BarType int

const (
	_ BarType = iota
	Hole
	Filled
)

func (v BarType) String() string {
	switch v {
	case Hole:
		return "H"
	case Filled:
		return "F"
	}
	panic("invalid bar type")
}

type Bar struct {
	Min  int64
	Max  int64
	Type BarType
}

type Point struct {
	Index int64
	Type  PointType
}

type HoleManager struct {
	pts  map[int64]PointType
	bars []Bar
}

func newHoleManager() *HoleManager {
	m := new(HoleManager)
	m.pts = make(map[int64]PointType)
	return m
}

func (m *HoleManager) addBar(b Bar) {
	switch b.Type {
	case Hole:
		m.pts[b.Min] = HoleStart
		m.pts[b.Max] = HoleStop
	case Filled:
		m.pts[b.Min] = FillStart
		if pt, ok := m.pts[b.Max]; ok {
			switch pt {
			case FillStart:
				delete(m.pts, b.Max)
			default:
				m.pts[b.Max] = FillStop
			}
		} else {
			m.pts[b.Max] = FillStop
		}
	}
	m.update()
}

func (m *HoleManager) update() {
	m.bars = m.getBars()
	m.pts = make(map[int64]PointType)
	for _, bar := range m.bars {
		switch bar.Type {
		case Filled:
			m.pts[bar.Min] = FillStart
			m.pts[bar.Max] = FillStop
		case Hole:
			m.pts[bar.Min] = HoleStart
			m.pts[bar.Max] = HoleStop
		}
	}
}

func (m *HoleManager) getBars() []Bar {
	pts := make([]Point, 0, len(m.pts))
	for idx, t := range m.pts {
		pts = append(pts, Point{Index: idx, Type: t})
	}
	sort.Slice(pts, func(i, j int) bool {
		return pts[i].Index < pts[j].Index
	})
	bars := make([]Bar, 0)

	startIdx := 0
	for startIdx < len(pts)-1 {
		switch pts[startIdx].Type {
		case HoleStart:
			endIdx := startIdx + 1
			keepGoing := true
			depth := 0
			for keepGoing {
				if endIdx == len(pts)-1 {
					bars = append(bars, Bar{Min: pts[startIdx].Index, Max: pts[endIdx].Index, Type: Hole})
					startIdx = endIdx
					break
				}
				switch pts[endIdx].Type {
				case HoleStop:
					if depth > 0 {
						depth--
						endIdx++
						break
					}
					bars = append(bars, Bar{Min: pts[startIdx].Index, Max: pts[endIdx].Index, Type: Hole})
					startIdx = endIdx
					pts[startIdx].Index++
					pts[startIdx].Type = FillStart
					keepGoing = false
				case FillStart:
					if depth > 0 {
						depth--
						endIdx++
						break
					}
					bars = append(bars, Bar{Min: pts[startIdx].Index, Max: pts[endIdx].Index - 1, Type: Hole})
					startIdx = endIdx
					keepGoing = false
				case HoleStart, FillStop:
					depth++
					fallthrough
				default:
					endIdx++
				}
			}
		case FillStart:
			endIdx := startIdx + 1
			keepGoing := true
			depth := 0
			for keepGoing {
				if endIdx == len(pts)-1 {
					bars = append(bars, Bar{Min: pts[startIdx].Index, Max: pts[endIdx].Index, Type: Filled})
					startIdx = endIdx
					break
				}
				switch pts[endIdx].Type {
				case FillStop:
					if depth > 0 {
						depth--
						endIdx++
						break
					}
					bars = append(bars, Bar{Min: pts[startIdx].Index, Max: pts[endIdx].Index, Type: Filled})
					startIdx = endIdx
					pts[startIdx].Index++
					pts[startIdx].Type = HoleStart
					keepGoing = false
				case HoleStart:
					if depth > 0 {
						depth--
						endIdx++
						break
					}
					bars = append(bars, Bar{Min: pts[startIdx].Index, Max: pts[endIdx].Index - 1, Type: Filled})
					startIdx = endIdx
					keepGoing = false
				case FillStart, HoleStop:
					depth++
					fallthrough
				default:
					endIdx++
				}
			}
		default:
			startIdx++
		}
	}
	return bars
}

func (m *HoleManager) save() ([]byte, error) {
	b, err := json.Marshal(m.pts)
	return b, err
}

func (m *HoleManager) load(b []byte) error {
	err := json.Unmarshal(b, &m.pts)
	if err != nil {
		return err
	}
	m.getBars()
	return nil
}

func (m *HoleManager) isFilled(min, max int64) bool {
	for idx := range m.bars {
		if m.bars[idx].Type == Hole {
			continue
		}
		if min >= m.bars[idx].Min && max <= m.bars[idx].Max {
			return true
		}
	}
	return false
}

func SaveHole(peerID int64, peerType int32, minID, maxID int64) error {
	b, err := repo.Ctx().MessagesExtra.GetHoles(peerID, peerType)
	if err != nil {
		return err
	}
	hm := newHoleManager()
	err = hm.load(b)
	if err != nil {
		return err
	}

	hm.addBar(Bar{Type:Hole, Min:minID, Max:maxID})
	b, err = hm.save()
	if err != nil {
		return err
	}

	err = repo.Ctx().MessagesExtra.SaveHoles(peerID, peerType, b)
	if err != nil {
		return err
	}

	return nil
}

func FillHole(peerID int64, peerType int32, minID, maxID int64) error {
	b, err := repo.Ctx().MessagesExtra.GetHoles(peerID, peerType)
	if err != nil {
		return err
	}
	hm := newHoleManager()
	err = hm.load(b)
	if err != nil {
		return err
	}

	hm.addBar(Bar{Type:Filled, Min:minID, Max:maxID})
	b, err = hm.save()
	if err != nil {
		return err
	}

	err = repo.Ctx().MessagesExtra.SaveHoles(peerID, peerType, b)
	if err != nil {
		return err
	}

	return nil
}

func IsHole(peerID int64, peerType int32, minID, maxID int64) (bool, error) {
	b, err := repo.Ctx().MessagesExtra.GetHoles(peerID, peerType)
	if err != nil {
		return true, err
	}
	hm := newHoleManager()
	err = hm.load(b)
	if err != nil {
		return true, err
	}

	return hm.isFilled(minID, maxID), nil
}




// GetHoles get holes between min & max
func GetHoles(peerID, minID, maxID int64) []dto.MessagesHole {
	holes, err := repo.Ctx().MessageHoles.GetHoles(peerID, minID, maxID)
	if err != nil {
		return make([]dto.MessagesHole, 0)
	}
	return holes
}

// GetMinClosestHole find closest hole from lower side
func GetMinClosestHole(minID int64, holes []dto.MessagesHole) *dto.MessagesHole {
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
func GetMaxClosestHole(maxID int64, holes []dto.MessagesHole) *dto.MessagesHole {
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
			err := repo.Ctx().MessageHoles.Save(h.PeerID, h.MinID.Int64, msgMinID-1) // Update
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
