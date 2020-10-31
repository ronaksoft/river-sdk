package domain

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"math/big"
	"testing"
	"time"
)

/*
   Creation Time: 2020 - Jan - 04
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2018
*/

func TestSplitPQ(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 48)
	if err != nil {
		panic(err)
	}
	x := key.N.Uint64()

	fmt.Println(x)
	startTime := time.Now()
	p, q := SplitPQ(big.NewInt(int64(x)))
	fmt.Println(time.Now().Sub(startTime))
	fmt.Println(p, q)
}

func TestGenerateMessageKey(t *testing.T) {
	dhKey := StrToByte(RandomID(100))
	dhKey = append(dhKey, []byte("1234567890123456789012345678901234567890")...)
	body := []byte("Hello It is Ehsan")
	key := GenerateMessageKey(dhKey, body)
	fmt.Println(key)
}
