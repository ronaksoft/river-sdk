package riversdk

/*
   Creation Time: 2019 - Jul - 08
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2018
*/

var (
	funcHolders map[int]func(r *River)
)

func add(ver int, f func(river *River)) {
	funcHolders[ver] = f
}

func init() {
	funcHolders = make(map[int]func(r *River))
	add(0, func(r *River) {
		// r.ClearCache(0, "", true)
	})
	add (1, func(r *River) {
		//
	})

}
