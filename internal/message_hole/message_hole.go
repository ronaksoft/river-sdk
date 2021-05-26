package messageHole

import (
	"encoding/json"
	"fmt"
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/logs"
	"git.ronaksoft.com/river/sdk/internal/repo"
	"go.uber.org/zap"
	"sort"
	"strings"
	"sync"
)

var (
	logger *logs.Logger
)

func init() {
	logger = logs.With("MessageHole")
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

func (m *HoleManager) LoadFromDB(teamID, peerID int64, peerType int32, cat msg.MediaCategory) {
	b := repo.MessagesExtra.GetHoles(teamID, peerID, peerType, cat)
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

	// We sort the bar to find the max point, if b.Max is larger than biggest index, then we have to
	// insert a hole to increase the domain
	// TODO:: We must make sure that bars are sorted all the time then we don't need to sort every time
	sort.Slice(m.bars, func(i, j int) bool {
		return m.bars[i].Min < m.bars[j].Min
	})
	maxIndex := m.bars[len(m.bars)-1].Max
	if b.Max > maxIndex {
		m.bars = append(m.bars, Bar{Min: maxIndex + 1, Max: b.Max, Type: Hole})
	}

	currentBars := m.bars
	m.bars = make([]Bar, 0, len(currentBars)+1)

	// Initially the biggest index is b.Max. We will update the maxIndex during the range over bars if
	// necessary. In the first loop (InsertLoop) we go until we can insert the new bar into the list
	idx := 0
	m.maxIndex = b.Max
InsertLoop:
	for idx := 0; idx < len(currentBars); idx++ {
		switch {
		case b.Min > currentBars[idx].Max:
			m.appendBar(currentBars[idx])
		case b.Min > currentBars[idx].Min:
			switch {
			case b.Max < currentBars[idx].Max:
				m.appendBar(
					Bar{Min: currentBars[idx].Min, Max: b.Min - 1, Type: currentBars[idx].Type},
					b,
					Bar{Min: b.Max + 1, Max: currentBars[idx].Max, Type: currentBars[idx].Type},
				)
				m.maxIndex = currentBars[idx].Max
			case b.Max == currentBars[idx].Max:
				m.appendBar(
					Bar{Min: currentBars[idx].Min, Max: b.Min - 1, Type: currentBars[idx].Type},
					b,
				)
			case b.Max > currentBars[idx].Max:
				m.appendBar(
					Bar{Min: currentBars[idx].Min, Max: b.Min - 1, Type: currentBars[idx].Type},
					b,
				)
			}
			break InsertLoop
		case b.Min == currentBars[idx].Min:
			switch {
			case b.Max < currentBars[idx].Max:
				m.appendBar(
					Bar{Min: b.Min, Max: b.Max, Type: b.Type},
					Bar{Min: b.Max + 1, Max: currentBars[idx].Max, Type: currentBars[idx].Type},
				)
				m.maxIndex = currentBars[idx].Max
			default:
				m.appendBar(b)
			}
			break InsertLoop
		}
	}

	// In this loop, we are assured that the new bar has been already added, we try to append the remaining
	// bars to the list
	for ; idx < len(currentBars); idx++ {
		switch {
		case currentBars[idx].Min < m.maxIndex:
			switch {
			case currentBars[idx].Max > m.maxIndex:
				m.appendBar(Bar{Min: m.maxIndex + 1, Max: currentBars[idx].Max, Type: currentBars[idx].Type})
				m.maxIndex = currentBars[idx].Max
			}
		case currentBars[idx].Min == m.maxIndex:
			if currentBars[idx].Max > currentBars[idx].Min {
				m.appendBar(Bar{Min: currentBars[idx].Min + 1, Max: currentBars[idx].Max, Type: currentBars[idx].Type})
				m.maxIndex = currentBars[idx].Max
			}
		default:
			m.appendBar(currentBars[idx])
			m.maxIndex = currentBars[idx].Max
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

func Init() {
	holder.list = make(map[string]*HoleManager)
}

func loadManager(teamID, peerID int64, peerType int32, cat msg.MediaCategory) *HoleManager {
	keyID := fmt.Sprintf("%d.%d", peerID, peerType)
	holder.mtx.Lock()
	defer holder.mtx.Unlock()
	hm, ok := holder.list[keyID]
	if !ok {
		hm = newHoleManager()
		hm.LoadFromDB(teamID, peerID, peerType, cat)
		holder.list[keyID] = hm
	}

	if !hm.Valid() {
		logger.Error("HoleManager Not Valid", zap.String("Dump", hm.String()))
		hm = newHoleManager()
		b, _ := json.Marshal(hm)
		repo.MessagesExtra.SaveHoles(teamID, peerID, peerType, cat, b)
		holder.list[keyID] = hm
	}
	return hm
}

func saveManager(teamID, peerID int64, peerType int32, cat msg.MediaCategory, hm *HoleManager) {
	b, err := json.Marshal(hm.bars)
	if err != nil {
		logger.Error("Error On HoleManager", zap.Error(err))
		return
	}
	repo.MessagesExtra.SaveHoles(teamID, peerID, peerType, cat, b)
}

func InsertFill(teamID, peerID int64, peerType int32, cat msg.MediaCategory, minID, maxID int64) {
	if minID > maxID {
		return
	}
	hm := loadManager(teamID, peerID, peerType, cat)
	hm.InsertBar(Bar{Type: Filled, Min: minID, Max: maxID})
	saveManager(teamID, peerID, peerType, cat, hm)
}

// IsHole Checks if there is any hole in the range [minID-maxID].
func IsHole(teamID, peerID int64, peerType int32, cat msg.MediaCategory, minID, maxID int64) bool {
	hm := loadManager(teamID, peerID, peerType, cat)
	return hm.IsRangeFilled(minID, maxID)
}

// GetUpperFilled It returns a LabelBar starts from minID to the highest possible index,
// which makes a continuous Filled section, otherwise it returns false.
func GetUpperFilled(teamID, peerID int64, peerType int32, cat msg.MediaCategory, minID int64) (bool, Bar) {
	hm := loadManager(teamID, peerID, peerType, cat)
	return hm.GetUpperFilled(minID)
}

// GetLowerFilled It returns a LabelBar starts from the lowest possible index to maxID,
// which makes a continuous Filled section, otherwise it returns false.
func GetLowerFilled(teamID, peerID int64, peerType int32, cat msg.MediaCategory, maxID int64) (bool, Bar) {
	hm := loadManager(teamID, peerID, peerType, cat)
	return hm.GetLowerFilled(maxID)
}

func PrintHole(teamID, peerID int64, peerType int32, cat msg.MediaCategory) string {
	hm := loadManager(teamID, peerID, peerType, cat)
	sb := strings.Builder{}
	for _, bar := range hm.bars {
		sb.WriteString(fmt.Sprintf("[%s: %d - %d]", bar.Type.String(), bar.Min, bar.Max))
	}
	return sb.String()
}
