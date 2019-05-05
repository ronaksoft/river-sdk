package messageHole

import (
	"fmt"
	"testing"
)


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
