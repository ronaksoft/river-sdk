package domain

type MInt64B map[int64]bool

func (m MInt64B) ToArray() []int64 {
	a := make([]int64, 0, len(m))
	for k := range m {
		a = append(a, k)
	}
	return a
}
