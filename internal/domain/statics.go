package domain

import (
    "bufio"
    "bytes"
    "crypto/aes"
    "crypto/cipher"
    "crypto/sha256"
    "fmt"
    "hash/crc32"
    "io"
    "log"
    "math"
    "math/big"
    "math/rand"
    "os"
    "regexp"
    "sort"
    "strings"
    "sync/atomic"
    "time"

    "github.com/nyaruka/phonenumbers"
    "github.com/ronaksoft/river-msg/go/msg"
    "github.com/ronaksoft/rony/pools"
    "github.com/ronaksoft/rony/tools"
)

func init() {
    rand.Seed(time.Now().UnixNano())
}

var (
    WindowLog     func(txt string)
    StartTime     time.Time
    TimeDelta     time.Duration
    uniqueCounter int64
    _RegExPhone   *regexp.Regexp
)

const (
    ALPHANUMERICS = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
)

func init() {
    exp, err := regexp.Compile(`^\d*$`)
    if err != nil {
        log.Fatal(err.Error())
    }
    WindowLog = func(txt string) {}
    _RegExPhone = exp
    rand.Seed(time.Now().UnixNano())
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
    msgKey := make([]byte, 32)
    keyBuffer := pools.Bytes.GetLen(40 + len(plain))
    defer pools.Bytes.Put(keyBuffer)

    copy(keyBuffer, dhKey[100:140])
    copy(keyBuffer[40:], plain)
    var sh512 [64]byte
    err := tools.Sha512(keyBuffer, sh512[:0])
    if err != nil {
        return nil
    }
    copy(msgKey, sh512[32:64])
    return msgKey
}

// Encrypt ...
// Generate MessageKey, AES IV and AES Key:
// 1. Message Key is: _Sha512(dhKey[100:140], plain)[32:64]
// 2. AES IV: _Sha512 (dhKey[180:220], MessageKey)[:32]
// 3. AES KEY: _Sha512 (MessageKey, dhKey[170:210])[:32]
func Encrypt(dhKey, plain []byte) (encrypted []byte, err error) {
    // Message Key is: _Sha512(DHKey[100:140], InternalHeader, Payload)[32:64]
    msgKey := GenerateMessageKey(dhKey, plain)
    var (
        iv, key       [72]byte
        aesIV, aesKey [64]byte
    )

    // AES IV: _Sha512 (DHKey[180:220], MessageKey)[:32]
    copy(iv[:], dhKey[180:220])
    copy(iv[40:], msgKey)
    err = tools.Sha512(iv[:], aesIV[:0])
    if err != nil {
        return nil, err
    }

    // AES KEY: _Sha512 (MessageKey, DHKey[170:210])[:32]
    copy(key[:], msgKey)
    copy(key[32:], dhKey[170:210])
    err = tools.Sha512(key[:], aesKey[:0])
    if err != nil {
        return nil, err
    }

    return aes256GCMEncrypt(
        aesKey[:32],
        aesIV[:12],
        plain,
    )
}
func aes256GCMEncrypt(key, iv []byte, msg []byte) ([]byte, error) {
    var block cipher.Block
    if b, err := aes.NewCipher(key); err != nil {
        return nil, err
    } else {
        block = b
    }

    aesGCM, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }
    encrypted := pools.Bytes.GetCap(len(msg) + aesGCM.Overhead())
    encrypted = aesGCM.Seal(encrypted[:0], iv, msg, nil)
    return encrypted, nil
}

// Decrypt ...
// Decrypts the message:
// 1. AES IV: _Sha512 (dhKey[180:220], MessageKey)[:12]
// 2. AES KEY: _Sha512 (MessageKey, dhKey[170:210])[:32]
func Decrypt(dhKey, msgKey, encrypted []byte) (plain []byte, err error) {
    var (
        iv, key       [72]byte
        aesIV, aesKey [64]byte
    )

    // AES IV: _Sha512 (DHKey[180:220], MessageKey)[:32]
    copy(iv[:], dhKey[180:220])
    copy(iv[40:], msgKey)
    err = tools.Sha512(iv[:], aesIV[:0])
    if err != nil {
        return nil, err
    }

    // AES KEY: _Sha512 (MessageKey, DHKey[170:210])[:32]
    copy(key[:], msgKey)
    copy(key[32:], dhKey[170:210])
    err = tools.Sha512(key[:], aesKey[:0])
    if err != nil {
        return nil, err
    }

    return aes256GCMDecrypt(
        aesKey[:32],
        aesIV[:12],
        encrypted,
    )
}
func aes256GCMDecrypt(key, iv []byte, msg []byte) ([]byte, error) {
    var block cipher.Block
    if b, err := aes.NewCipher(key); err != nil {
        return nil, err
    } else {
        block = b
    }

    aesGCM, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }
    decrypted := pools.Bytes.GetCap(len(msg))
    decrypted, err = aesGCM.Open(decrypted[:0], iv, msg, nil)
    if err != nil {
        return nil, err
    }

    return decrypted, nil
}

// RandomUint64 produces a pseudo-random unsigned number
func RandomUint64() uint64 {
    return rand.Uint64()
}

func RandomInt64(n int64) (x int64) {
    if n <= 0 {
        return rand.Int63()
    } else {
        return rand.Int63n(n)
    }
}

func RandomInt(n int) (x int) {
    return rand.Intn(n)
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

func NextRequestID() uint64 {
    return uint64(SequentialUniqueID())
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
        bb.Write(tools.StrToByte(phones[idx].Phone))
        bb.Write(tools.StrToByte(phones[idx].FirstName))
        bb.Write(tools.StrToByte(phones[idx].LastName))
    }
    crc32Hash := crc32.ChecksumIEEE(bb.Bytes())
    return uint64(crc32Hash)
}

// ExtractsContactsDifference remove items from newContacts that already exist in oldContacts
func ExtractsContactsDifference(oldContacts, newContacts []*msg.PhoneContact) []*msg.PhoneContact {
    mapOld := make(map[string]*msg.PhoneContact, len(oldContacts))
    mapNew := make(map[string]*msg.PhoneContact, len(newContacts))
    for _, c := range oldContacts {
        c.Phone = SanitizePhone(c.Phone)
        mapOld[c.Phone] = c
    }
    for _, c := range newContacts {
        c.Phone = SanitizePhone(c.Phone)
        mapNew[c.Phone] = c
    }

    for key, oldContact := range mapOld {
        newContact, ok := mapNew[key]
        if ok {
            if newContact.LastName == oldContact.LastName && newContact.FirstName == oldContact.FirstName {
                delete(mapNew, key)
            }
        }
    }

    result := make([]*msg.PhoneContact, 0, len(mapNew))
    for _, v := range mapNew {
        result = append(result, v)
    }
    return result
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

func GetCountryCode(phone string) string {
    ph, err := phonenumbers.Parse(phone, "")
    if err != nil {
        return ""
    }
    return phonenumbers.GetRegionCodeForNumber(ph)
}

func Now() time.Time {
    return time.Now().Add(TimeDelta)
}

func CopyFile(inPath, outPath string) error {
    buf := make([]byte, 1<<13)
    src, err := os.Open(inPath)
    if err != nil {
        return err
    }
    dst, err := os.Create(outPath)
    if err != nil {
        return err
    }
    for {
        n, err := src.Read(buf)
        if err != nil && err != io.EOF {
            return err
        }
        if n == 0 {
            break
        }

        if _, err := dst.Write(buf[:n]); err != nil {
            return err
        }
    }
    return nil
}

func GetExponentialTime(min time.Duration, max time.Duration, attempts int) time.Duration {
    if min <= 0 {
        min = 100 * time.Millisecond
    }
    if max <= 0 {
        max = 10 * time.Second
    }
    minFloat := float64(min)
    napDuration := minFloat * math.Pow(2, float64(attempts-1))
    // add some jitter
    napDuration += rand.Float64()*minFloat - (minFloat / 2)
    if napDuration > float64(max) {
        return max
    }
    return time.Duration(napDuration)
}

func CalculateSha256(filePath string) ([]byte, error) {
    f, err := os.Open(filePath)
    if err != nil {
        return nil, err
    }
    h := sha256.New()
    buf := make([]byte, 1<<20)
    rd := bufio.NewReader(f)
    for {
        n, err := rd.Read(buf)
        if err != nil {
            if err == io.EOF {
                break
            }
            return nil, err
        }
        h.Write(buf[:n])
    }
    return h.Sum(nil), nil
}
