package hole

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

type Detector struct {
	mtx      sync.Mutex
	MaxIndex int64
	Bars     []Bar
}

func readFromDB(teamID, peerID int64, peerType int32, cat msg.MediaCategory) *Detector {
	m := &Detector{}
	b := repo.MessagesExtra.GetHoles(teamID, peerID, peerType, cat)
	_ = json.Unmarshal(b, &m.Bars)
	m.MaxIndex = 0
	for idx := range m.Bars {
		if m.Bars[idx].Max > m.MaxIndex {
			m.MaxIndex = m.Bars[idx].Max
		}
	}
	return m
}

func writeToDB(teamID, peerID int64, peerType int32, cat msg.MediaCategory, hm *Detector) {
	b, err := json.Marshal(hm.Bars)
	if err != nil {
		logger.Error("got error on marshalling hole", zap.Error(err))
		return
	}
	repo.MessagesExtra.SaveHoles(teamID, peerID, peerType, cat, b)
}

func load(teamID, peerID int64, peerType int32, cat msg.MediaCategory) *Detector {
	keyID := fmt.Sprintf("%d.%d", peerID, peerType)
	cache.mtx.Lock()
	defer cache.mtx.Unlock()
	hm, ok := cache.list[keyID]
	if !ok {
		hm = readFromDB(teamID, peerID, peerType, cat)
		cache.list[keyID] = hm
	}

	if !hm.Valid() {
		logger.Error("load invalid data, we reset hole",
			zap.Int64("TeamID", teamID),
			zap.Int64("PeerID", peerID),
			zap.String("Dump", hm.String()),
		)
		hm = &Detector{}
		b, _ := json.Marshal(hm)
		repo.MessagesExtra.SaveHoles(teamID, peerID, peerType, cat, b)
		cache.list[keyID] = hm
	}
	return hm
}

func (m *Detector) InsertBar(b Bar) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	// If it is the first bar
	if len(m.Bars) == 0 {
		if b.Min > 0 {
			m.Bars = append(m.Bars, Bar{Min: 0, Max: b.Min - 1, Type: Hole})
		}
		m.MaxIndex = b.Max
		m.Bars = append(m.Bars, b)
		return
	}

	// We sort the bar to find the max point, if b.Max is larger than biggest index, then we have to
	// insert a hole to increase the domain
	// TODO:: We must make sure that Bars are sorted all the time then we don't need to sort every time
	sort.Slice(m.Bars, func(i, j int) bool {
		return m.Bars[i].Min < m.Bars[j].Min
	})
	maxIndex := m.Bars[len(m.Bars)-1].Max
	if b.Max > maxIndex {
		m.Bars = append(m.Bars, Bar{Min: maxIndex + 1, Max: b.Max, Type: Hole})
	}

	currentBars := m.Bars
	m.Bars = make([]Bar, 0, len(currentBars)+1)

	// Initially the biggest index is b.Max. We will update the MaxIndex during the range over Bars if
	// necessary. In the first loop (InsertLoop) we go until we can insert the new bar into the list
	idx := 0
	m.MaxIndex = b.Max
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
				m.MaxIndex = currentBars[idx].Max
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
				m.MaxIndex = currentBars[idx].Max
			default:
				m.appendBar(b)
			}
			break InsertLoop
		}
	}

	// In this loop, we are assured that the new bar has been already added, we try to append the remaining
	// Bars to the list
	for ; idx < len(currentBars); idx++ {
		switch {
		case currentBars[idx].Min < m.MaxIndex:
			switch {
			case currentBars[idx].Max > m.MaxIndex:
				m.appendBar(Bar{Min: m.MaxIndex + 1, Max: currentBars[idx].Max, Type: currentBars[idx].Type})
				m.MaxIndex = currentBars[idx].Max
			}
		case currentBars[idx].Min == m.MaxIndex:
			if currentBars[idx].Max > currentBars[idx].Min {
				m.appendBar(Bar{Min: currentBars[idx].Min + 1, Max: currentBars[idx].Max, Type: currentBars[idx].Type})
				m.MaxIndex = currentBars[idx].Max
			}
		default:
			m.appendBar(currentBars[idx])
			m.MaxIndex = currentBars[idx].Max
		}
	}
}

func (m *Detector) appendBar(bars ...Bar) {
	for _, b := range bars {
		lastIndex := len(m.Bars) - 1
		if lastIndex >= 0 && m.Bars[lastIndex].Type == b.Type {
			m.Bars[lastIndex].Max = b.Max
		} else {
			m.Bars = append(m.Bars, b)
		}
	}
}

func (m *Detector) IsRangeFilled(min, max int64) bool {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	for idx := range m.Bars {
		if m.Bars[idx].Type == Hole {
			continue
		}
		if min >= m.Bars[idx].Min && max <= m.Bars[idx].Max {
			return true
		}
	}
	return false
}

func (m *Detector) IsPointHole(pt int64) bool {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	for idx := range m.Bars {
		if pt >= m.Bars[idx].Min && pt <= m.Bars[idx].Max {
			switch m.Bars[idx].Type {
			case Filled:
				return false
			case Hole:
				return true
			}
		}
	}
	return true
}

func (m *Detector) GetUpperFilled(pt int64) (bool, Bar) {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	for idx := range m.Bars {
		if pt >= m.Bars[idx].Min && pt <= m.Bars[idx].Max {
			switch m.Bars[idx].Type {
			case Filled:
				return true, Bar{Min: pt, Max: m.Bars[idx].Max, Type: Filled}
			case Hole:
				return false, Bar{}
			}
		}
	}
	return false, Bar{}
}

func (m *Detector) GetLowerFilled(pt int64) (bool, Bar) {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	for idx := range m.Bars {
		if pt >= m.Bars[idx].Min && pt <= m.Bars[idx].Max {
			switch m.Bars[idx].Type {
			case Filled:
				return true, Bar{Min: m.Bars[idx].Min, Max: pt, Type: Filled}
			case Hole:
				return false, Bar{}
			}
		}
	}
	return false, Bar{}
}

func (m *Detector) SetUpperFilled(pt int64) bool {
	if pt <= m.MaxIndex {
		return false
	}
	m.InsertBar(Bar{Type: Filled, Min: m.MaxIndex + 1, Max: pt})
	return true
}

func (m *Detector) SetLowerFilled() {
	for _, b := range m.Bars {
		if b.Type == Filled {
			if b.Min != 0 {
				m.InsertBar(Bar{Min: 0, Max: b.Min, Type: Filled})
			}
		}
	}
}

func (m *Detector) String() string {
	sb := strings.Builder{}
	for _, bar := range m.Bars {
		sb.WriteString(fmt.Sprintf("[%s: %d - %d]", bar.Type.String(), bar.Min, bar.Max))
	}
	return sb.String()
}

func (m *Detector) Valid() bool {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	idx := int64(-1)
	for _, bar := range m.Bars {
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

var cache = struct {
	mtx  sync.Mutex
	list map[string]*Detector
}{
	list: make(map[string]*Detector),
}

func Init() {
	cache.list = make(map[string]*Detector)
}

func InsertFill(teamID, peerID int64, peerType int32, cat msg.MediaCategory, minID, maxID int64) {
	if minID > maxID {
		return
	}
	hm := load(teamID, peerID, peerType, cat)
	hm.InsertBar(Bar{Type: Filled, Min: minID, Max: maxID})
	writeToDB(teamID, peerID, peerType, cat, hm)
}

// IsHole Checks if there is any hole in the range [minID-maxID].
func IsHole(teamID, peerID int64, peerType int32, cat msg.MediaCategory, minID, maxID int64) bool {
	hm := load(teamID, peerID, peerType, cat)
	return hm.IsRangeFilled(minID, maxID)
}

// GetUpperFilled It returns a LabelBar starts from minID to the highest possible index,
// which makes a continuous Filled section, otherwise it returns false.
func GetUpperFilled(teamID, peerID int64, peerType int32, cat msg.MediaCategory, minID int64) (bool, Bar) {
	hm := load(teamID, peerID, peerType, cat)
	return hm.GetUpperFilled(minID)
}

// GetLowerFilled It returns a LabelBar starts from the lowest possible index to maxID,
// which makes a continuous Filled section, otherwise it returns false.
func GetLowerFilled(teamID, peerID int64, peerType int32, cat msg.MediaCategory, maxID int64) (bool, Bar) {
	hm := load(teamID, peerID, peerType, cat)
	return hm.GetLowerFilled(maxID)
}

func PrintHole(teamID, peerID int64, peerType int32, cat msg.MediaCategory) string {
	hm := load(teamID, peerID, peerType, cat)
	sb := strings.Builder{}
	for _, bar := range hm.Bars {
		sb.WriteString(fmt.Sprintf("[%s: %d - %d]", bar.Type.String(), bar.Min, bar.Max))
	}
	return sb.String()
}
