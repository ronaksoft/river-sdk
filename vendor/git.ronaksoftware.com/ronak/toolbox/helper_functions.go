package ronak

import (
    "hash/crc32"
    "math/rand"
    "time"
    "crypto/sha256"
    "crypto/sha512"
    "crypto/cipher"
    "crypto/aes"
    "hash/crc64"
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
    _CRC_TABLE_32 *crc32.Table
    _CRC_TABLE_64 *crc64.Table
)

func init() {
    _CRC_TABLE_32 = crc32.MakeTable(0xE2374063)
    _CRC_TABLE_64 = crc64.MakeTable(0x23740630002374)
    rand.Seed(time.Now().UnixNano())
}

// CRC generates an uint32 from in
func CRC32(in []byte) uint32 {
    return crc32.Checksum(in, _CRC_TABLE_32)
}

func CRC64(in []byte) uint64 {
    return crc64.Checksum(in, _CRC_TABLE_64)
}

// RandomID generates a pseudo-random string with length 'n' which characters are alphanumerics.
func RandomID(n int) string {
    b := make([]byte, n)
    for i := range b {
        b[i] = ALPHANUMERICS[rand.Intn(len(ALPHANUMERICS))]
    }
    return string(b)
}

// RandomDigit generates a pseudo-random string with length 'n' which characters are only digits (0-9)
func RandomDigit(n int) string {
    rand.Seed(time.Now().UnixNano())
    b := make([]byte, n)
    for i := range b {
        b[i] = DIGITS[rand.Intn(len(DIGITS))]
    }
    return string(b)
}

// RandomInt64 produces a pseudo-random number, if n == 0 there will be no limit otherwise
// the output will be smaller than n
func RandomInt64(n int64) int64 {
    if n == 0 {
        return rand.Int63()
    } else {
        return rand.Int63n(n)
    }
}

func RandomInt(n int) int {
    if n == 0 {
        return rand.Int()
    } else {
        return rand.Intn(n)
    }
}

// RandUint64 produces a pseudo-random unsigned number
func RandomUint64() uint64 {
    return rand.Uint64()
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
