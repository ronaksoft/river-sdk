package messageHole

import (
	"encoding/json"
	"fmt"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo"
	"go.uber.org/zap"
	"sort"
	"strings"
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

func (m *HoleManager) isRangeFilled(min, max int64) bool {
	for idx := range m.getBars() {
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
	for idx := range m.getBars() {
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
	for idx := range m.getBars() {
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
	for idx := range m.getBars() {
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
	hm := newHoleManager()
	b, err := repo.MessagesExtra.GetHoles(peerID, peerType)
	if err == nil {
		err = json.Unmarshal(b, &hm.pts)
		if err != nil {
			return nil, err
		}
	}
	hm.getBars()
	return hm, nil
}

func saveManager(peerID int64, peerType int32, hm *HoleManager) error {
	b, err := json.Marshal(hm.pts)
	if err != nil {
		return err
	}

	err = repo.MessagesExtra.SaveHoles(peerID, peerType, b)
	if err != nil {
		return err
	}
	return nil
}

func InsertHole(peerID int64, peerType int32, minID, maxID int64) error {
	logs.Info("Insert Hole",
		zap.Int64("MinID", minID),
		zap.Int64("MaxID", maxID),
	)
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

// SetUpperFilled Marks from the top index to 'msgID' as filled. This could be used
// when UpdateNewMessage arrives we just add Fill bar to the end
func SetUpperFilled(peerID int64, peerType int32, msgID int64) error {
	logs.Info("SetUpperFilled",
		zap.Int64("MsgID", msgID),
	)
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

// SetLowerFilled Marks from the lowest index to zero as filled. This is useful when
// we reached to the first message of the user/group.
func SetLowerFilled(peerID int64, peerType int32) error {
	hm, err := loadManager(peerID, peerType)
	if err != nil {
		return err
	}

	hm.addBar(Bar{Type: Filled, Min: 0, Max: hm.minIndex})

	err = saveManager(peerID, peerType, hm)
	if err != nil {
		return err
	}

	return nil
}

// IsHole Checks if there is any hole in the range [minID-maxID].
func IsHole(peerID int64, peerType int32, minID, maxID int64) (bool, error) {
	hm, err := loadManager(peerID, peerType)
	if err != nil {
		return true, err
	}
	return hm.isRangeFilled(minID, maxID), nil
}

// GetUpperFilled It returns a Bar starts from minID to the highest possible index,
// which makes a continuous Filled section, otherwise it returns false.
func GetUpperFilled(peerID int64, peerType int32, minID int64) (bool, Bar) {
	hm, err := loadManager(peerID, peerType)
	if err != nil {
		return false, Bar{}
	}
	return hm.getUpperFilled(minID)
}

// GetLowerFilled It returns a Bar starts from the lowest possible index to maxID,
// which makes a continuous Filled section, otherwise it returns false.
func GetLowerFilled(peerID int64, peerType int32, maxID int64) (bool, Bar) {
	hm, err := loadManager(peerID, peerType)
	if err != nil {
		logs.Error(err.Error())
		return false, Bar{}
	}
	return hm.getLowerFilled(maxID)
}

func PrintHole(peerID int64, peerType int32) string {
	hm, err := loadManager(peerID, peerType)
	if err != nil {
		return err.Error()
	}
	sb := strings.Builder{}
	for _, bar := range hm.getBars() {
		sb.WriteString(fmt.Sprintf("[%s: %d - %d]", bar.Type.String(), bar.Min, bar.Max))
	}
	return sb.String()
}
