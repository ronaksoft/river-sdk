package domain

// MInt64B simple type to get distinct IDs
type MInt64B map[int64]bool

// ToArray get values
func (m MInt64B) ToArray() []int64 {
	a := make([]int64, 0, len(m))
	for k := range m {
		a = append(a, k)
	}
	return a
}
type Slt struct {
	Value      int64 `json:"value"`
	Timestamp int64  `json:"timestamp"`
}