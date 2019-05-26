package ronak

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"crypto/sha512"
	"hash/crc32"
	"hash/crc64"
	"math/rand"
	"reflect"
	"strconv"
	"sync"
	"time"
	"unsafe"
)

/*
   Creation Time: 2018 - Apr - 07
   Created by:  Ehsan N. Moosa (ehsan)
   Maintainers:
       1.  Ehsan N. Moosa (ehsan)
   Auditor: Ehsan N. Moosa
   Copyright Ronak Software Group 2018
*/

const (
	DIGITS        = "0123456789"
	ALPHANUMERICS = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
)

var (
	crcTable32 *crc32.Table
	crcTable64 *crc64.Table
)

func init() {
	crcTable32 = crc32.MakeTable(0xE2374063)
	crcTable64 = crc64.MakeTable(0x23740630002374)
	rand.Seed(time.Now().UnixNano())
}

// CRC generates an uint32 from in
func CRC32(in []byte) uint32 {
	return crc32.Checksum(in, crcTable32)
}

func CRC64(in []byte) uint64 {
	return crc64.Checksum(in, crcTable64)
}

// AES256CCMEncrypt encrypts the msg according with key and iv
func AES256GCMEncrypt(key, iv []byte, msg []byte) ([]byte, error) {
	var block cipher.Block
	if b, err := aes.NewCipher(key); err != nil {
		return nil, err
	} else {
		block = b
	}
	var encrypted []byte
	if aesGCM, err := cipher.NewGCM(block); err != nil {
		return nil, err
	} else {
		encrypted = aesGCM.Seal(msg[:0], iv, msg, nil)
	}
	return encrypted, nil
}

// AES256GCMDecrypt decrypts the msg according with key and iv
func AES256GCMDecrypt(key, iv []byte, msg []byte) ([]byte, error) {
	var block cipher.Block
	if b, err := aes.NewCipher(key); err != nil {
		return nil, err
	} else {
		block = b
	}
	var decrypted []byte
	if aesGCM, err := cipher.NewGCM(block); err != nil {
		return nil, err
	} else {
		decrypted, err = aesGCM.Open(nil, iv, msg, nil)
		if err != nil {
			return nil, err
		}
	}
	return decrypted, nil
}

// Sha256 returns a 32bytes array which is sha256(in)
func Sha256(in []byte) ([]byte, error) {
	h := sha256.New()
	if _, err := h.Write(in); err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}

// Sha512 returns a 64bytes array which is sha512(in)
func Sha512(in []byte) ([]byte, error) {
	h := sha512.New()
	if _, err := h.Write(in); err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}

func StrToInt64(s string) int64 {
	v, _ := strconv.ParseInt(s, 10, 64)
	return v
}

func StrToInt32(s string) int32 {
	v, _ := strconv.ParseInt(s, 10, 32)
	return int32(v)
}

func StrToUInt64(s string) uint64 {
	v, _ := strconv.ParseInt(s, 10, 64)
	return uint64(v)
}

func StrToUInt32(s string) uint32 {
	v, _ := strconv.ParseInt(s, 10, 32)
	return uint32(v)
}

// ByteToStr converts byte slice to a string without memory allocation.
// Note it may break if string and/or slice header will change
// in the future go versions.
func ByteToStr(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// StrToByte converts string to a byte slice without memory allocation.
// Note it may break if string and/or slice header will change
// in the future go versions.
func StrToByte(s string) []byte {
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh := reflect.SliceHeader{
		Data: sh.Data,
		Len:  sh.Len,
		Cap:  sh.Len,
	}
	return *(*[]byte)(unsafe.Pointer(&bh))
}


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
