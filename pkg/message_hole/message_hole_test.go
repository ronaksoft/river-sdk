package messageHole

import (
	"fmt"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	"testing"
)

func TestMessageHole(t *testing.T) {
	logs.SetLogLevel(-1)
	hm := newHoleManager()
	hm.addBar(Bar{0, 100, Hole})
	hm.addBar(Bar{10, 30, Filled})
	hm.addBar(Bar{60, 70, Filled})
	b, err := hm.save()
	if err != nil {
		t.Error(err)
	}

	hm2 := newHoleManager()
	err = hm2.load(b)
	if err != nil {
		t.Error(err)
	}

	if len(hm.pts) != len(hm.pts) {
		t.Fail()
	}

	if hm.isRangeFilled(20, 40) {
		t.Fail()
	}
	if !hm.isRangeFilled(65, 70) {
		t.Fail()
	}
}

func TestIsHole(t *testing.T) {
	hm := newHoleManager()
	hm.addBar(Bar{0, 99, Hole})
	hm.addBar(Bar{10, 30, Filled})
	hm.addBar(Bar{60, 70, Filled})

	bars := hm.getBars()
	for _, b := range bars {
		t.Log(fmt.Sprintf("%s: %d ---> %d", b.Type.String(), b.Min, b.Max))
	}

	t.Log("Point: 20 [Lower Filled]")
	b, bar := hm.getLowerFilled(20)
	if b {
		t.Log(fmt.Sprintf("%s: %d ---> %d", bar.Type.String(), bar.Min, bar.Max))
	} else {
		t.Log("It is Hole")
	}
	t.Log("Point: 20 [Upper Filled]")
	b, bar = hm.getUpperFilled(20)
	if b {
		t.Log(fmt.Sprintf("%s: %d ---> %d", bar.Type.String(), bar.Min, bar.Max))
	} else {
		t.Log("It is Hole")
	}

	t.Log("Point: 35 [Lower Filled]")
	b, bar = hm.getLowerFilled(35)
	if b {
		t.Log(fmt.Sprintf("%s: %d ---> %d", bar.Type.String(), bar.Min, bar.Max))
	} else {
		t.Log("It is Hole")
	}

	t.Log("Point: 101 [Lower Filled]")
	b, bar = hm.getLowerFilled(101)
	if b {
		t.Log(fmt.Sprintf("%s: %d ---> %d", bar.Type.String(), bar.Min, bar.Max))
	} else {
		t.Log("It is Hole")
	}

	t.Log("Point: 101 [Upper Filled]")
	b, bar = hm.getUpperFilled(101)
	if b {
		t.Log(fmt.Sprintf("%s: %d ---> %d", bar.Type.String(), bar.Min, bar.Max))
	} else {
		t.Log("It is Hole")
	}
}
