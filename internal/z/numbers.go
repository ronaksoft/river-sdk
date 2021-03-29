package z

/*
   Creation Time: 2020 - Oct - 05
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2020
*/

func AbsInt64(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}

func AbsInt32(x int32) int32 {
	if x < 0 {
		return -x
	}
	return x
}
