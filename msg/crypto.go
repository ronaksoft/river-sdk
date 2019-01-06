package msg

import (
	"git.ronaksoftware.com/ronak/toolbox"
	"github.com/gobwas/pool/pbytes"
)

// GenerateMessageKey
// Message Key is: Sha512(DHKey[100:140], InternalHeader, Payload)[32:64]
func GenerateMessageKey(authKey, plain []byte, msgKey []byte) error {
	// Message Key is: Sha512(DHKey[100:140], InternalHeader, Payload)[32:64]
	keyBuffer := pbytes.GetLen(40 + len(plain))
	defer pbytes.Put(keyBuffer)
	copy(keyBuffer, authKey[100:140])
	copy(keyBuffer[40:], plain)
	if k, err := ronak.Sha512(keyBuffer); err != nil {
		return err
	} else {
		copy(msgKey, k[32:64])
		return nil
	}
}

// Encrypt
// Generate MessageKey, AES IV and AES Key:
// 1. Message Key is: Sha512(dhKey[100:140], plain)[32:64]
// 2. AES IV: Sha512 (dhKey[180:220], MessageKey)[:32]
// 3. AES KEY: Sha512 (MessageKey, dhKey[170:210])[:32]
func Encrypt(authKey, msgKey, plain []byte) (encrypted []byte, err error) {
	// AES IV: Sha512 (DHKey[180:220], MessageKey)[:32]
	iv := make([]byte, 72)
	copy(iv, authKey[180:220])
	copy(iv[40:], msgKey)
	aesIV, err := ronak.Sha512(iv)
	if err != nil {
		return nil, err
	}

	// AES KEY: Sha512 (MessageKey, DHKey[170:210])[:32]
	key := make([]byte, 72)
	copy(key, msgKey)
	copy(key[32:], authKey[170:210])
	aesKey, err := ronak.Sha512(key)
	if err != nil {
		return nil, err
	}

	return ronak.AES256GCMEncrypt(
		aesKey[:32],
		aesIV[:12],
		plain,
	)

}

// Decrypt
// Decrypts the message:
// 1. AES IV: Sha512 (dhKey[180:220], MessageKey)[:32]
// 2. AES KEY: Sha512 (MessageKey, dhKey[170:210])[:32]
func Decrypt(authKey, msgKey, encrypted []byte) (plain []byte, err error) {
	// AES IV: Sha512 (DHKey[180:220], MessageKey)[:32]
	iv := make([]byte, 40, 72)
	copy(iv, authKey[180:220])
	iv = append(iv, msgKey...)
	aesIV, _ := ronak.Sha512(iv)

	// AES KEY: Sha512 (MessageKey, DHKey[170:210])[:32]
	key := make([]byte, 32, 72)
	copy(key, msgKey)
	key = append(key, authKey[170:210]...)
	aesKey, err := ronak.Sha512(key)
	if err != nil {
		return nil, err
	}

	return ronak.AES256GCMDecrypt(
		aesKey[:32],
		aesIV[:12],
		encrypted,
	)
}
