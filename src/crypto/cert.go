package crypto

import (
	"bytes"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
)

var (
	PADDING = []byte{0x00, 0x00, 0x00, 0x00}
)

func randomKey(length int) (key []byte, err error) {
	key = make([]byte, length)
	_, err = rand.Read(key)
	if err != nil {
		return
	}
	return key, nil
}

func aes256CbcEncrypt(plaintext string, key []byte, iv []byte, blockSize int) (ciphertext []byte, err error) {
	bPlaintext := PKCS5Padding([]byte(plaintext), blockSize)
	block, _ := aes.NewCipher(key)
	encrypted := make([]byte, len(bPlaintext))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(encrypted, bPlaintext)
	encryptednew := append(iv, encrypted...)
	return encryptednew, nil
}

func aes256CbcDecrypt(ciphertext []byte, key []byte, iv []byte) (plaintext string, err error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return
	}
	decrypted := make([]byte, len(ciphertext))
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(decrypted, ciphertext)
	return string(decrypted), nil
}

func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func RSAEncrypt(pubkey *rsa.PublicKey, data string) (ciphertext string, err error) {
	key, err := randomKey(32)
	iv, err := randomKey(16)
	encryptedData, err := aes256CbcEncrypt(data, key, iv, aes.BlockSize)
	if err != nil {
		return
	}
	encryptedKey, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, pubkey, key, nil)
	if err != nil {
		return
	}
	paddedData := append(encryptedKey, PADDING...)
	paddedData = append(paddedData, encryptedData...)
	encoded := base64.StdEncoding.EncodeToString(paddedData)
	return encoded, nil
}

func RSADecrypt(privkey rsa.PrivateKey, data []byte) (plaintext string, err error) {
	unpadded := bytes.Split(data, PADDING)
	encryptedKey := unpadded[0]
	ciphertext := unpadded[1]

	decryptedKey, err := privkey.Decrypt(nil, encryptedKey, &rsa.OAEPOptions{Hash: crypto.SHA256})
	if err != nil {
		return
	}

	iv := ciphertext[:16]
	ciphertext = ciphertext[16:]
	decryptedText, err := aes256CbcDecrypt(ciphertext, decryptedKey, iv)
	if err != nil {
		return
	}
	return decryptedText, nil
}

func EncodeToPEM(key rsa.PublicKey) (certificate string, err error) {
	pubASN1, err := x509.MarshalPKIXPublicKey(&key)
	if err != nil {
		return
	}
	pubBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: pubASN1,
	})
	return string(pubBytes), nil
}

func DecodePublicPEM(certificate string) (key *rsa.PublicKey, err error) {
	pemBlock, _ := pem.Decode([]byte(certificate))
	pub, err := x509.ParsePKCS1PublicKey(pemBlock.Bytes)
	if err != nil {
		return
	}
	return pub, nil
}

func DecodePrivatePEM(certificate string) (key *rsa.PrivateKey, err error) {
	pemBlock, _ := pem.Decode([]byte(certificate))
	priv, err := x509.ParsePKCS1PrivateKey(pemBlock.Bytes)
	if err != nil {
		return
	}
	return priv, nil
}
