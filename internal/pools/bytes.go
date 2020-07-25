package pools

import "github.com/gobwas/pool/pbytes"

/*
   Creation Time: 2020 - Apr - 08
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2018
*/

var (
	TinyBytes = pbytes.New(32, 256)
	Bytes     = pbytes.DefaultPool
)
