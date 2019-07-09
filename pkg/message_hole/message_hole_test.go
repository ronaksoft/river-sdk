package messageHole

import (
	"fmt"
	"testing"
)

func TestIsHole(t *testing.T) {
	hm := newHoleManager()
	hm.InsertBar(Bar{0, 99, Hole})
	fmt.Println(hm.bars)
	hm.InsertBar(Bar{10, 30, Filled})
	fmt.Println(hm.bars)
	hm.InsertBar(Bar{60, 70, Filled})
	fmt.Println(hm.bars)

	for _, b := range hm.bars {
		t.Log(fmt.Sprintf("%s: %d ---> %d", b.Type.String(), b.Min, b.Max))
	}
}

func TestMessageID(t *testing.T) {
	hm := newHoleManager()
	hm.InsertBar(Bar{0, 9873, Hole})
	hm.InsertBar(Bar{9872, 9873, Filled})
	hm.InsertBar(Bar{8721, 9872, Filled})
	hm.InsertBar(Bar{7269, 8167, Filled})
	fmt.Println(hm.bars)
	hm.InsertBar(Bar{6977, 7268, Filled})
	fmt.Println(hm.bars)
}

func TestHole1(t *testing.T) {
	hm := newHoleManager()
	hm.InsertBar(Bar{0, 100, Hole})
	hm.SetUpperFilled(101)
	fmt.Println(hm.bars)
	hm.InsertBar(Bar{200, 210, Filled})
	fmt.Println(hm.bars)
	fmt.Println(hm.IsRangeFilled(99,101))
	fmt.Println(hm.IsRangeFilled(50,100))
	fmt.Println(hm.IsRangeFilled(200,210))
	// hm.SetUpperFilled(110)
	// fmt.Println(hm.bars)
}