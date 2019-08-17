package domain

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"log"
	"math/big"
	"math/rand"
	"regexp"
	"sort"
	"strings"
	"sync/atomic"
	"time"

	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
	"github.com/nyaruka/phonenumbers"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var (
	TimeDelta     time.Duration
	uniqueCounter int64
	_RegExPhone   *regexp.Regexp
)

const (
	ALPHANUMERICS = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
)

func init() {
	exp, err := regexp.Compile("^\\d*$")
	if err != nil {
		log.Fatal(err.Error())
	}
	_RegExPhone = exp

}

// SplitPQ ...
// This function used for proof of work to splits PQ to two prime numbers P and Q
func SplitPQ(pq *big.Int) (p1, p2 *big.Int) {
	value0 := big.NewInt(0)
	value1 := big.NewInt(1)
	value15 := big.NewInt(15)
	value17 := big.NewInt(17)
	rndMax := big.NewInt(0).SetBit(big.NewInt(0), 64, 1)

	what := big.NewInt(0).Set(pq)
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	g := big.NewInt(0)
	i := 0
	for !(g.Cmp(value1) == 1 && g.Cmp(what) == -1) {
		q := big.NewInt(0).Rand(rnd, rndMax)
		q = q.And(q, value15)
		q = q.Add(q, value17)
		q = q.Mod(q, what)

		x := big.NewInt(0).Rand(rnd, rndMax)
		whatNext := big.NewInt(0).Sub(what, value1)
		x = x.Mod(x, whatNext)
		x = x.Add(x, value1)

		y := big.NewInt(0).Set(x)
		lim := 1 << (uint(i) + 18)
		j := 1
		flag := true

		for j < lim && flag {
			a := big.NewInt(0).Set(x)
			b := big.NewInt(0).Set(x)
			c := big.NewInt(0).Set(q)

			for b.Cmp(value0) == 1 {
				b2 := big.NewInt(0)
				if b2.And(b, value1).Cmp(value0) == 1 {
					c.Add(c, a)
					if c.Cmp(what) >= 0 {
						c.Sub(c, what)
					}
				}
				a.Add(a, a)
				if a.Cmp(what) >= 0 {
					a.Sub(a, what)
				}
				b.Rsh(b, 1)
			}
			x.Set(c)

			z := big.NewInt(0)
			if x.Cmp(y) == -1 {
				z.Add(what, x)
				z.Sub(z, y)
			} else {
				z.Sub(x, y)
			}
			g.GCD(nil, nil, z, what)

			if (j & (j - 1)) == 0 {
				y.Set(x)
			}
			j = j + 1

			if g.Cmp(value1) != 0 {
				flag = false
			}
		}
		i = i + 1
	}

	p1 = big.NewInt(0).Set(g)
	p2 = big.NewInt(0).Div(what, g)

	if p1.Cmp(p2) == 1 {
		p1, p2 = p2, p1
	}

	return
}

// GenerateMessageKey ...
// Message Key is: _Sha512(DHKey[100:140], InternalHeader, Payload)[32:64]
func GenerateMessageKey(dhKey, plain []byte) []byte {
	// Message Key is: _Sha512(DHKey[100:140], InternalHeader, Payload)[32:64]
	keyBuffer := make([]byte, 40+len(plain))
	copy(keyBuffer, dhKey[100:140])
	copy(keyBuffer[40:], plain)
	if k, err := Sha512(keyBuffer); err != nil {
		return nil
	} else {
		return k[32:64]

	}
}

// Encrypt ...
// Generate MessageKey, AES IV and AES Key:
// 1. Message Key is: _Sha512(dhKey[100:140], plain)[32:64]
// 2. AES IV: _Sha512 (dhKey[180:220], MessageKey)[:32]
// 3. AES KEY: _Sha512 (MessageKey, dhKey[170:210])[:32]
func Encrypt(dhKey, plain []byte) (encrypted []byte, err error) {
	// Message Key is: _Sha512(DHKey[100:140], InternalHeader, Payload)[32:64]
	msgKey := GenerateMessageKey(dhKey, plain)

	// AES IV: _Sha512 (DHKey[180:220], MessageKey)[:32]
	iv := make([]byte, 72)
	copy(iv, dhKey[180:220])
	copy(iv[40:], msgKey)
	aesIV, err := Sha512(iv)
	if err != nil {
		return nil, err
	}

	// AES KEY: _Sha512 (MessageKey, DHKey[170:210])[:32]
	key := make([]byte, 72)
	copy(key, msgKey)
	copy(key[32:], dhKey[170:210])
	aesKey, err := Sha512(key)
	if err != nil {
		return nil, err
	}

	return AES256GCMEncrypt(
		aesKey[:32],
		aesIV[:12],
		plain,
	)

}

// Decrypt ...
// Decrypts the message:
// 1. AES IV: _Sha512 (dhKey[180:220], MessageKey)[:12]
// 2. AES KEY: _Sha512 (MessageKey, dhKey[170:210])[:32]
func Decrypt(dhKey, msgKey, encrypted []byte) (plain []byte, err error) {
	// AES IV: _Sha512 (DHKey[180:220], MessageKey)[:32]
	iv := make([]byte, 40, 72)
	copy(iv, dhKey[180:220])
	iv = append(iv, msgKey...)
	aesIV, _ := Sha512(iv)

	// AES KEY: _Sha512 (MessageKey, DHKey[170:210])[:32]
	key := make([]byte, 32, 72)
	copy(key, msgKey)
	key = append(key, dhKey[170:210]...)
	aesKey, err := Sha512(key)
	if err != nil {
		return nil, err
	}

	return AES256GCMDecrypt(
		aesKey[:32],
		aesIV[:12],
		encrypted,
	)
}

// AES256GCMEncrypt encrypts the msg according with key and iv
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

// RandomUint64 produces a pseudo-random unsigned number
func RandomUint64() uint64 {
	return rand.Uint64()
}

// RandomInt63 produces a pseudo-random signed number
func RandomInt63() int64 {
	return rand.Int63()
}

// RandomID generates a pseudo-random string with length 'n' which characters are alphanumerics.
func RandomID(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = ALPHANUMERICS[rand.Intn(len(ALPHANUMERICS))]
	}
	return string(b)
}

// SequentialUniqueID generate sequential and unique int64 with UnixNano time for sequential and atomic counter for uniqueness
func SequentialUniqueID() int64 {
	counter := atomic.AddInt64(&uniqueCounter, 1)
	nanoTime := time.Now().UnixNano()
	res := nanoTime + counter

	// reset counter we need counter only for requests that created in same nano second
	if counter > 16384 {
		atomic.StoreInt64(&uniqueCounter, 0)
	}
	return res
}

// CalculateContactsGetHash crc32 of UserIDs
func CalculateContactsGetHash(userIDs []int64) uint32 {
	// sort ASC
	sort.Slice(userIDs, func(i, j int) bool { return userIDs[i] < userIDs[j] })
	buff := bytes.Buffer{}
	b := make([]byte, 8)
	for _, id := range userIDs {
		binary.BigEndian.PutUint64(b, uint64(id))
		buff.Write(b)
	}
	crc32Hash := crc32.ChecksumIEEE(buff.Bytes())
	return crc32Hash
}

// CalculateContactsImportHash crc32 of phones
func CalculateContactsImportHash(req *msg.ContactsImport) uint64 {
	phoneContacts := make(map[string]*msg.PhoneContact)
	for _, c := range req.Contacts {
		phoneContacts[c.Phone] = c
	}
	phones := make([]*msg.PhoneContact, 0)
	for _, p := range phoneContacts {
		phones = append(phones, p)
	}
	sort.Slice(phones, func(i, j int) bool { return phones[i].Phone < phones[j].Phone })
	count := len(phones)
	bb := bytes.Buffer{}
	for idx := 0; idx < count; idx++ {
		bb.Write([]byte(phones[idx].Phone))
	}
	crc32Hash := crc32.ChecksumIEEE(bb.Bytes())
	return uint64(crc32Hash)
}

// SanitizePhone copy of server side function
func SanitizePhone(phoneNumber string) string {
	phoneNumber = strings.TrimLeft(phoneNumber, " +0")
	if !_RegExPhone.MatchString(phoneNumber) {
		return ""
	}
	if strings.HasPrefix(phoneNumber, "237400") {
		return phoneNumber
	}
	phone, err := phonenumbers.Parse(phoneNumber, "IR")
	if err != nil {
		return phoneNumber
	}
	return fmt.Sprintf("%d%d", *phone.CountryCode, *phone.NationalNumber)

}

// ExtractsContactsDifference remove items from newContacts that already exist in oldContacts
func ExtractsContactsDifference(oldContacts, newContacts []*msg.PhoneContact) []*msg.PhoneContact {

	mapOld := make(map[string]*msg.PhoneContact)
	mapNew := make(map[string]*msg.PhoneContact)
	for _, c := range oldContacts {
		c.Phone = SanitizePhone(c.Phone)
		mapOld[c.Phone] = c
	}
	for _, c := range newContacts {
		c.Phone = SanitizePhone(c.Phone)
		mapNew[c.Phone] = c
	}

	ok := false
	for key := range mapOld {
		_, ok = mapNew[key]
		if ok {
			delete(mapNew, key)
		}
	}

	result := make([]*msg.PhoneContact, 0, len(mapNew))
	for _, v := range mapNew {
		result = append(result, v)
	}
	return result
}

func Now() time.Time {
	return time.Now().Add(TimeDelta)
}
