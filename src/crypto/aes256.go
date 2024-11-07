package crypto

import (
	"crypto/aes"
	"crypto/cipher"
)

var (
	NPUBBYTES = 12
)

func AES256GCM(ciphertext []byte, key []byte) (plaintext string, err error) {
	nonce := ciphertext[3 : 3+NPUBBYTES]
	ciphertext = ciphertext[3+NPUBBYTES:]

	block, err := aes.NewCipher(key)
	if err != nil {
		return
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return
	}

	decrypted, err := gcm.Open(nil, nonce, ciphertext, nil)

	return string(decrypted), nil
}
