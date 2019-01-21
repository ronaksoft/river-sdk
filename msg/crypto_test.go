package msg

import (
	"crypto/rand"
	"git.ronaksoftware.com/ronak/toolbox"
	"testing"
)

/*
   Creation Time: 2019 - Jan - 16
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2018
*/

func BenchmarkGenerateMessageKey(b *testing.B) {
	total := 10000
	plain := []byte("This is just a text to be encrypted. We are not changing the text")
	authKeys := make([][]byte, 0, total)
	for i := 0; i < total; i++ {
		x := make([]byte, 256)
		rand.Read(x)
		authKeys = append(authKeys, x)
	}
	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		var msgKey []byte
		for pb.Next() {
			GenerateMessageKey(authKeys[ronak.RandomInt(total)], plain, msgKey)
		}
	})

}

func BenchmarkEncrypt(b *testing.B) {
	total := 10000
	plain := []byte("This is just a text to be encrypted. We are not changing the text")
	authKeys := make([][]byte, 0, total)
	for i := 0; i < total; i++ {
		x := make([]byte, 256)
		rand.Read(x)
		authKeys = append(authKeys, x)
	}
	b.ReportAllocs()
	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		var msgKey []byte
		for pb.Next() {
			_ = GenerateMessageKey(authKeys[0], plain, msgKey)
			_, _ = Encrypt(authKeys[0], msgKey, plain)
		}
	})
}
