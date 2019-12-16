package labelHole_test

import (
	"fmt"
	labelHole "git.ronaksoftware.com/ronak/riversdk/pkg/label_hole"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo"
	"testing"
)

/*
   Creation Time: 2019 - Dec - 16
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2018
*/

func init() {
	repo.InitRepo("./_data", false)
}

func TestGetLowerFilled(t *testing.T) {
	b, bar := labelHole.GetLowerFilled(100)
	fmt.Println(b, bar)
	labelHole.Fill(10, 100)
	b, bar = labelHole.GetLowerFilled(100)
	fmt.Println(b, bar)
	labelHole.Fill(50, 150)
	b, bar = labelHole.GetLowerFilled(100)
	fmt.Println(b, bar)

}
