package messageHole

import (
	"encoding/json"
	"fmt"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo"
	"go.uber.org/zap"
	"sort"
	"strings"
	"sync"
)

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
	return ""
}

type Bar struct {
	Min  int64
	Max  int64
	Type BarType
}

type HoleManager struct {
	mtxLock  sync.Mutex
	maxIndex int64
	bars     []Bar
}

func newHoleManager() *HoleManager {
	m := new(HoleManager)
	return m
}

func (m *HoleManager) LoadFromDB(peerID int64, peerType int32) {
	b := repo.MessagesExtra.GetHoles(peerID, peerType)
	_ = json.Unmarshal(b, &m.bars)
	m.maxIndex = 0
	for idx := range m.bars {
		if m.bars[idx].Max > m.maxIndex {
			m.maxIndex = m.bars[idx].Max
		}
	}

}

func (m *HoleManager) InsertBar(b Bar) {
	m.mtxLock.Lock()
	defer m.mtxLock.Unlock()

	// If it is the first bar
	if len(m.bars) == 0 {
		if b.Min > 0 {
			m.bars = append(m.bars, Bar{Min: 0, Max: b.Min - 1, Type: Hole})
		}
		m.maxIndex = b.Max
		m.bars = append(m.bars, b)
		return
	}

	// Insert hole to increase our domain
	if b.Max > m.maxIndex {
		m.bars = append(m.bars, Bar{Min: m.maxIndex + 1, Max: b.Max, Type: Hole})
		m.maxIndex = b.Max
	}

	sort.Slice(m.bars, func(i, j int) bool {
		return m.bars[i].Min < m.bars[j].Min
	})

	oldBars := m.bars
	m.bars = make([]Bar, 0, len(oldBars))
	newBarAdded := false

	for _, bar := range oldBars {
		if newBarAdded {
			switch {
			case b.Max == bar.Min:
				if bar.Max > bar.Min {
					m.appendBar(Bar{Min: bar.Min + 1, Max: bar.Max, Type: bar.Type})
				}
			case b.Max > bar.Min && b.Max < bar.Max:
				m.appendBar(Bar{Min: b.Max + 1, Max: bar.Max, Type: bar.Type})
			default:
				m.appendBar(bar)
			}
			continue
		}
		switch {
		case b.Min > bar.Max:
			m.appendBar(bar)
		case b.Min > bar.Min:
			switch {
			case b.Max < bar.Max:
				m.appendBar(
					Bar{Min: bar.Min, Max: b.Min - 1, Type: bar.Type},
					b,
					Bar{Min: b.Max + 1, Max: bar.Max, Type: bar.Type},
				)
			case b.Max == bar.Max:
				m.appendBar(
					Bar{Min: bar.Min, Max: b.Min - 1, Type: bar.Type},
					b,
				)
			case b.Max > bar.Max:
				m.appendBar(
					Bar{Min: bar.Min, Max: b.Min - 1, Type: bar.Type},
					b,
				)
			}
			newBarAdded = true
		case b.Min == bar.Min:
			switch {
			case b.Max < bar.Max:
				m.appendBar(
					Bar{Min: b.Min, Max: b.Max, Type: b.Type},
					Bar{Min: b.Max + 1, Max: bar.Max, Type: bar.Type},
				)
			default:
				m.appendBar(b)
			}
			newBarAdded = true
		}
	}
}

func (m *HoleManager) appendBar(bars ...Bar) {
	for _, b := range bars {
		lastIndex := len(m.bars) - 1
		if lastIndex >= 0 && m.bars[lastIndex].Type == b.Type {
			m.bars[lastIndex].Max = b.Max
		} else {
			m.bars = append(m.bars, b)
		}
	}
}

func (m *HoleManager) IsRangeFilled(min, max int64) bool {
	m.mtxLock.Lock()
	defer m.mtxLock.Unlock()
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

func (m *HoleManager) IsPointHole(pt int64) bool {
	m.mtxLock.Lock()
	defer m.mtxLock.Unlock()
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

func (m *HoleManager) GetUpperFilled(pt int64) (bool, Bar) {
	m.mtxLock.Lock()
	defer m.mtxLock.Unlock()
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

func (m *HoleManager) GetLowerFilled(pt int64) (bool, Bar) {
	m.mtxLock.Lock()
	defer m.mtxLock.Unlock()
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

func (m *HoleManager) SetUpperFilled(pt int64) bool {
	if pt <= m.maxIndex {
		return false
	}
	m.InsertBar(Bar{Type: Filled, Min: m.maxIndex + 1, Max: pt})
	return true
}

func (m *HoleManager) SetLowerFilled() {
	for _, b := range m.bars {
		if b.Type == Filled {
			if b.Min != 0 {
				m.InsertBar(Bar{Min: 0, Max: b.Min, Type: Filled})
			}
		}
	}
}

func (m *HoleManager) String() string {
	sb := strings.Builder{}
	for _, bar := range m.bars {
		sb.WriteString(fmt.Sprintf("[%s: %d - %d]", bar.Type.String(), bar.Min, bar.Max))
	}
	return sb.String()
}

func (m *HoleManager) Valid() bool {
	m.mtxLock.Lock()
	defer m.mtxLock.Unlock()
	idx := int64(-1)
	for _, bar := range m.bars {
		if bar.Min > bar.Max {
			return false
		}
		if bar.Min <= idx {
			return false
		}
		idx = bar.Max
	}
	return true
}

var holder = struct {
	mtx  sync.Mutex
	list map[string]*HoleManager
}{
	list: make(map[string]*HoleManager),
}

func loadManager(peerID int64, peerType int32) *HoleManager {
	keyID := fmt.Sprintf("%d.%d", peerID, peerType)
	holder.mtx.Lock()
	defer holder.mtx.Unlock()
	hm, ok := holder.list[keyID]
	if !ok {
		hm = newHoleManager()
		hm.LoadFromDB(peerID, peerType)
		holder.list[keyID] = hm
	}

	if !hm.Valid() {
		logs.Error("HoleManager Not Valid", zap.String("Dump", hm.String()))
		hm = newHoleManager()
		b, _ := json.Marshal(hm)
		repo.MessagesExtra.SaveHoles(peerID, peerType, b)
		holder.list[keyID] = hm
	}
	return hm
}

func saveManager(peerID int64, peerType int32, hm *HoleManager) {
	b, err := json.Marshal(hm.bars)
	if err != nil {
		logs.Error("Error On HoleManager", zap.Error(err))
		return
	}
	repo.MessagesExtra.SaveHoles(peerID, peerType, b)
	return
}

func InsertHole(peerID int64, peerType int32, minID, maxID int64) {
	logs.Info("Insert Hole",
		zap.Int64("MinID", minID),
		zap.Int64("MaxID", maxID),
	)
	if minID > maxID {
		return
	}
	hm := loadManager(peerID, peerType)

	hm.InsertBar(Bar{Type: Hole, Min: minID, Max: maxID})

	saveManager(peerID, peerType, hm)

	return
}

func InsertFill(peerID int64, peerType int32, minID, maxID int64) {
	logs.Info("Insert Fill",
		zap.Int64("MinID", minID),
		zap.Int64("MaxID", maxID),
	)
	if minID > maxID {
		return
	}
	hm := loadManager(peerID, peerType)
	logs.Info("Before", zap.String("Obj", hm.String()))
	hm.InsertBar(Bar{Type: Filled, Min: minID, Max: maxID})
	logs.Info("After", zap.String("Obj", hm.String()))

	saveManager(peerID, peerType, hm)
	return
}

// SetUpperFilled Marks from the top index to 'msgID' as filled. This could be used
// when UpdateNewMessage arrives we just add Fill bar to the end
func SetUpperFilled(peerID int64, peerType int32, msgID int64) {
	logs.Info("SetUpperFilled", zap.Int64("MsgID", msgID))
	hm := loadManager(peerID, peerType)

	if !hm.SetUpperFilled(msgID) {
		return
	}

	saveManager(peerID, peerType, hm)
	return
}

// SetLowerFilled Marks from the lowest index to zero as filled. This is useful when
// we reached to the first message of the user/group.
func SetLowerFilled(peerID int64, peerType int32) {
	hm := loadManager(peerID, peerType)
	hm.SetLowerFilled()
	saveManager(peerID, peerType, hm)
	return
}

// IsHole Checks if there is any hole in the range [minID-maxID].
func IsHole(peerID int64, peerType int32, minID, maxID int64) bool {
	hm := loadManager(peerID, peerType)
	return hm.IsRangeFilled(minID, maxID)
}

// GetUpperFilled It returns a Bar starts from minID to the highest possible index,
// which makes a continuous Filled section, otherwise it returns false.
func GetUpperFilled(peerID int64, peerType int32, minID int64) (bool, Bar) {
	hm := loadManager(peerID, peerType)
	return hm.GetUpperFilled(minID)
}

// GetLowerFilled It returns a Bar starts from the lowest possible index to maxID,
// which makes a continuous Filled section, otherwise it returns false.
func GetLowerFilled(peerID int64, peerType int32, maxID int64) (bool, Bar) {
	hm := loadManager(peerID, peerType)
	return hm.GetLowerFilled(maxID)
}

func PrintHole(peerID int64, peerType int32) string {
	hm := loadManager(peerID, peerType)
	sb := strings.Builder{}
	for _, bar := range hm.bars {
		sb.WriteString(fmt.Sprintf("[%s: %d - %d]", bar.Type.String(), bar.Min, bar.Max))
	}
	return sb.String()
}
