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

type HoleManager struct {
	maxIndex int64
	bars     []Bar
}

func newHoleManager() *HoleManager {
	m := new(HoleManager)
	return m
}

func (m *HoleManager) insertBar(b Bar) {
	fmt.Println("Add Bar", b)
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
					m.appendBar( Bar{Min: bar.Min + 1, Max: bar.Max, Type: bar.Type})
				}
			case b.Max > bar.Min && b.Max < bar.Max:
				m.appendBar( Bar{Min: b.Max + 1, Max: bar.Max, Type: bar.Type})
			}
			continue
		}
		switch {
		case b.Min > bar.Max:
			m.appendBar( bar)
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
				m.appendBar( b)
			}
			newBarAdded = true

		}
	}
}

func (m *HoleManager) appendBar(bars ...Bar) {
	for _, b := range bars {
		lastIndex :=len(m.bars)-1
		if lastIndex >= 0 && m.bars[lastIndex].Type == b.Type {
			m.bars[lastIndex].Max = b.Max
		} else {
			m.bars = append(m.bars, b)
		}
	}
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

func (m *HoleManager) setUpperFilled(pt int64) bool {
	if pt <= m.maxIndex {
		return false
	}

	m.insertBar(Bar{Type: Filled, Min: m.maxIndex, Max: pt})
	return true
}

func (m *HoleManager) setLowerFilled() {
	for _, b := range m.bars {
		if b.Type == Filled {
			if b.Min != 0 {
				m.insertBar(Bar{Min: 0, Max: b.Min, Type: Filled})
			}
		}
	}
}

func loadManager(peerID int64, peerType int32) (*HoleManager, error) {
	hm := newHoleManager()
	b, err := repo.MessagesExtra.GetHoles(peerID, peerType)
	if err == nil {
		err = json.Unmarshal(b, &hm.bars)
		if err != nil {
			return nil, err
		}
	}
	return hm, nil
}

func saveManager(peerID int64, peerType int32, hm *HoleManager) error {
	b, err := json.Marshal(hm.bars)
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

	hm.insertBar(Bar{Type: Hole, Min: minID, Max: maxID})

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

	hm.insertBar(Bar{Type: Filled, Min: minID, Max: maxID})

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

	if !hm.setUpperFilled(msgID) {
		return nil
	}

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

	hm.setLowerFilled()

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
	for _, bar := range hm.bars {
		sb.WriteString(fmt.Sprintf("[%s: %d - %d]", bar.Type.String(), bar.Min, bar.Max))
	}
	return sb.String()
}
