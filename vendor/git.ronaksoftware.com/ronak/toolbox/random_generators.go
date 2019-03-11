package ronak

import (
	"math/rand"
	"sync"
	"time"
)

type randomGenerator struct {
	sync.Pool
}

func (rg *randomGenerator) GetRand() *rand.Rand {
	return rg.Get().(*rand.Rand)
}

func (rg *randomGenerator) PutRand(r *rand.Rand) {
	rg.Put(r)
}

var rndGen randomGenerator

func init() {
	rndGen.New = func() interface{} {
		x := rand.New(rand.NewSource(time.Now().UnixNano()))
		return x
	}
}

// RandomID generates a pseudo-random string with length 'n' which characters are alphanumerics.
func RandomID(n int) string {
	rnd := rndGen.GetRand()
	defer rndGen.PutRand(rnd)
	b := make([]byte, n)
	for i := range b {
		b[i] = ALPHANUMERICS[rnd.Intn(len(ALPHANUMERICS))]
	}
	return string(b)
}

// RandomDigit generates a pseudo-random string with length 'n' which characters are only digits (0-9)
func RandomDigit(n int) string {
	rnd := rndGen.GetRand()
	defer rndGen.PutRand(rnd)
	b := make([]byte, n)
	for i := 0; i < len(b); i++ {
		b[i] = DIGITS[rnd.Intn(len(DIGITS))]
	}
	return string(b)
}

// RandomInt64 produces a pseudo-random number, if n == 0 there will be no limit otherwise
// the output will be smaller than n
func RandomInt64(n int64) (x int64) {
	rnd := rndGen.GetRand()
	if n == 0 {
		x = rnd.Int63()
	} else {
		x = rnd.Int63n(n)
	}
	rndGen.PutRand(rnd)
	return
}

func RandomInt(n int) (x int) {
	rnd := rndGen.GetRand()

	if n == 0 {
		x = rnd.Int()
	} else {
		x = rnd.Intn(n)
	}
	rndGen.PutRand(rnd)
	return
}

// RandUint64 produces a pseudo-random unsigned number
func RandomUint64() (x uint64) {
	rnd := rndGen.GetRand()
	x = rnd.Uint64()
	rndGen.PutRand(rnd)
	return
}
