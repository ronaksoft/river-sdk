package z

import (
	"fmt"
	"testing"
)

/*
   Creation Time: 2019 - Nov - 25
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2018
*/

func TestDeleteItemFromArray(t *testing.T) {
	x := []int32{1, 2, 3, 4, 5, 6, 7, 8}
	fmt.Println("Before:", x)
	DeleteItemFromArray(&x, 3)
	fmt.Println("After:", x)

	for idx := 0; idx < len(x); idx++ {
		DeleteItemFromArray(&x, idx)
		fmt.Println("Deleted:", idx, x)
		idx--
	}
	fmt.Println("After All:", x)
}
