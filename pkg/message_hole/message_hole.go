package messageHole

import (
	"encoding/json"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo"
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
	minIndex int64
	maxIndex int64
	pts      map[int64]PointType
	bars     []Bar
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
	if len(pts) == 0 {
		return bars
	}

	m.minIndex = pts[0].Index
	m.maxIndex = pts[len(pts)-1].Index

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

func (m *HoleManager) isRangeFilled(min, max int64) bool {
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

func (m *HoleManager) isPointHole(pt int64) bool {
	for idx := range m.bars {
		if pt >= m.bars[idx].Min && pt <= m.bars[idx].Max {
			switch m.bars[idx].Type {
			case Filled:
				return false
			case Hole:
				return true
			}
		}
	}
	return true
}

func (m *HoleManager) getUpperFilled(pt int64) (bool, Bar) {
	for idx := range m.bars {
		if pt >= m.bars[idx].Min && pt <= m.bars[idx].Max {
			switch m.bars[idx].Type {
			case Filled:
				return true, Bar{Min: pt, Max: m.bars[idx].Max, Type: Filled}
			case Hole:
				return false, Bar{}
			}
		}
	}
	return false, Bar{}
}

func (m *HoleManager) getLowerFilled(pt int64) (bool, Bar) {
	for idx := range m.bars {
		if pt >= m.bars[idx].Min && pt <= m.bars[idx].Max {
			switch m.bars[idx].Type {
			case Filled:
				return true, Bar{Min: m.bars[idx].Min, Max: pt, Type: Filled}
			case Hole:
				return false, Bar{}
			}
		}
	}
	return false, Bar{}
}

func loadManager(peerID int64, peerType int32) (*HoleManager, error) {
	b, err := repo.Ctx().MessagesExtra.GetHoles(peerID, peerType)
	if err != nil {
		return nil, err
	}
	hm := newHoleManager()
	err = hm.load(b)
	if err != nil {
		return nil, err
	}
	return hm, nil
}

func saveManager(peerID int64, peerType int32, hm *HoleManager) error {
	b, err := hm.save()
	if err != nil {
		return err
	}

	err = repo.Ctx().MessagesExtra.SaveHoles(peerID, peerType, b)
	if err != nil {
		return err
	}
	return nil
}

func InsertHole(peerID int64, peerType int32, minID, maxID int64) error {
	hm, err := loadManager(peerID, peerType)
	if err != nil {
		return err
	}

	hm.addBar(Bar{Type: Hole, Min: minID, Max: maxID})

	err = saveManager(peerID, peerType, hm)
	if err != nil {
		return err
	}

	return nil
}

func InsertFill(peerID int64, peerType int32, minID, maxID int64) error {
	hm, err := loadManager(peerID, peerType)
	if err != nil {
		return err
	}

	hm.addBar(Bar{Type: Filled, Min: minID, Max: maxID})

	err = saveManager(peerID, peerType, hm)
	if err != nil {
		return err
	}

	return nil
}

func AddFill(peerID int64, peerType int32, msgID int64) error {
	hm, err := loadManager(peerID, peerType)
	if err != nil {
		return err
	}

	if msgID <= hm.maxIndex {
		return nil
	}

	hm.addBar(Bar{Type: Filled, Min: hm.maxIndex, Max: msgID})

	err = saveManager(peerID, peerType, hm)
	if err != nil {
		return err
	}

	return nil
}
func IsHole(peerID int64, peerType int32, minID, maxID int64) (bool, error) {
	hm, err := loadManager(peerID, peerType)
	if err != nil {
		return true, err
	}
	return hm.isRangeFilled(minID, maxID), nil
}

func GetUpperFilled(peerID int64, peerType int32, minID int64) (bool, Bar) {
	hm, err := loadManager(peerID, peerType)
	if err != nil {
		return false, Bar{}
	}
	return hm.getUpperFilled(minID)
}

func GetLowerFilled(peerID int64, peerType int32, maxID int64) (bool, Bar) {
	hm, err := loadManager(peerID, peerType)
	if err != nil {
		return false, Bar{}
	}
	return hm.getLowerFilled(maxID)
}
