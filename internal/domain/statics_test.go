package domain

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"github.com/ronaksoft/rony/tools"
	. "github.com/smartystreets/goconvey/convey"
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
	Convey("Generating MessageKey" , t, func(c C) {
		dhKey := tools.StrToByte(RandomID(256))
		body := []byte("Hello It is Ehsan")
		key := GenerateMessageKey(dhKey, body)
		c.Println(key)
	})

}

func TestEncrypt(t *testing.T) {
	Convey("Encrypt/Decrypt Test",t, func(c C) {
		dhKey := tools.StrToByte(RandomID(256))
		body := []byte("Hello It is Ehsan")
		key := GenerateMessageKey(dhKey, body)
		encrypted, err := Encrypt(dhKey, body)
		c.So(err, ShouldBeNil)
		dec, err := Decrypt(dhKey, key, encrypted)
		c.So(err, ShouldBeNil)
		c.So(dec, ShouldResemble, body)
	})


}