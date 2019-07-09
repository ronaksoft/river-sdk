package messageHole

import (
	"fmt"
	"testing"
)

func TestIsHole(t *testing.T) {
	hm := newHoleManager()
	hm.insertBar(Bar{0, 99, Hole})
	fmt.Println(hm.bars)
	hm.insertBar(Bar{10, 30, Filled})
	fmt.Println(hm.bars)
	hm.insertBar(Bar{60, 70, Filled})
	fmt.Println(hm.bars)

	for _, b := range hm.bars {
		t.Log(fmt.Sprintf("%s: %d ---> %d", b.Type.String(), b.Min, b.Max))
	}
}

func TestMessageID(t *testing.T) {
	hm := newHoleManager()
	hm.insertBar(Bar{0, 9873, Hole})
	hm.insertBar(Bar{9872, 9873, Filled})
	hm.insertBar(Bar{8721, 9872, Filled})
	hm.insertBar(Bar{7269, 8167, Filled})
	fmt.Println(hm.bars)
	hm.insertBar(Bar{6977, 7268, Filled})
	fmt.Println(hm.bars)
}

func TestHole1(t *testing.T) {
	hm := newHoleManager()
	hm.insertBar(Bar{0, 100, Hole})
	hm.setUpperFilled(101)
	fmt.Println(hm.bars)
	hm.insertBar(Bar{200, 210, Filled})
	fmt.Println(hm.bars)
	fmt.Println(hm.isRangeFilled(99,101))
	fmt.Println(hm.isRangeFilled(50,100))
	fmt.Println(hm.isRangeFilled(200,210))
	// hm.setUpperFilled(110)
	// fmt.Println(hm.bars)
}