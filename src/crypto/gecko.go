package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/des"
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/asn1"
	"encoding/base64"
	"errors"
	"golang.org/x/crypto/pbkdf2"
)


type Gecko struct {
	MasterKey []byte
}

type slatAttr struct {
	EntrySalt      []byte
	IterationCount int
	KeySize        int
	Algorithm      struct {
		asn1.ObjectIdentifier
	}
}

type ivAttr struct {
	asn1.ObjectIdentifier
	IV []byte
}

type algoAttr struct {
	asn1.ObjectIdentifier
	Data struct {
		Data struct {
			asn1.ObjectIdentifier
			SlatAttr slatAttr
		}
		IVData ivAttr
	}
}

type metaPBE struct {
	AlgoAttr  algoAttr
	Encrypted []byte
}

type loginPBE struct {
	CipherText []byte
	Data       struct {
		asn1.ObjectIdentifier
		IV []byte
	}
	Encrypted []byte
}

type nssPBE struct {
	AlgoAttr struct {
		asn1.ObjectIdentifier
		SaltAttr struct {
			EntrySalt []byte
			Len       int
		}
	}
	Encrypted []byte
}

type ASN1PBE interface {
	Decrypt(globalSalt, masterPwd []byte) (key []byte, err error)
}

func GeckoDecrypt(ciphertext []byte, masterKey []byte) (plaintext []byte, err error) {
	PBE, err := NewASN1PBE(ciphertext)
	if err != nil {
		return
	}
	var key []byte
	return PBE.Decrypt(masterKey, key)
}

func NewASN1PBE(data []byte) (pbe ASN1PBE, err error) {
	var (
		n nssPBE
		m metaPBE
		l loginPBE
	)
	if _, err := asn1.Unmarshal(data, &n); err == nil {
		return n, nil
	}
	if _, err := asn1.Unmarshal(data, &m); err == nil {
		return m, nil
	}
	if _, err := asn1.Unmarshal(data, &l); err == nil {
		return l, nil
	}
	return nil, errors.New("decode ASN1 data failed")
}

func (l loginPBE) Decrypt(globalSalt, _ []byte) (key []byte, err error) {
	return des3Decrypt(globalSalt, l.Data.IV, l.Encrypted)
}

func (m metaPBE) Decrypt(globalSalt, _ []byte) (key2 []byte, err error) {
	k := sha1.Sum(globalSalt)
	key := pbkdf2.Key(k[:], m.AlgoAttr.Data.Data.SlatAttr.EntrySalt, m.AlgoAttr.Data.Data.SlatAttr.IterationCount, m.AlgoAttr.Data.Data.SlatAttr.KeySize, sha256.New)
	iv := append([]byte{4, 14}, m.AlgoAttr.Data.IVData.IV...)
	return aes128CBCDecrypt(key, iv, m.Encrypted)
}

func (n nssPBE) Decrypt(globalSalt, masterPwd []byte) (key []byte, err error) {
	masterHash := sha1.Sum(append(globalSalt, masterPwd...))
	s := append(masterHash[:], n.AlgoAttr.SaltAttr.EntrySalt...)
	chp := sha1.Sum(s)
	pes := paddingZero(n.AlgoAttr.SaltAttr.EntrySalt, 20)
	tk := hmac.New(sha1.New, chp[:])
	tk.Write(pes)
	pes = append(pes, n.AlgoAttr.SaltAttr.EntrySalt...)
	k1 := hmac.New(sha1.New, chp[:])
	k1.Write(pes)
	tkPlus := append(tk.Sum(nil), n.AlgoAttr.SaltAttr.EntrySalt...)
	k2 := hmac.New(sha1.New, chp[:])
	k2.Write(tkPlus)
	k := append(k1.Sum(nil), k2.Sum(nil)...)
	iv := k[len(k)-8:]
	return des3Decrypt(k[:24], iv, n.Encrypted)
}

func aes128CBCDecrypt(key, iv, encryptPass []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	encryptLen := len(encryptPass)
	if encryptLen < block.BlockSize() {
		return nil, errors.New("length of encrypted password less than block size")
	}

	dst := make([]byte, encryptLen)
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(dst, encryptPass)
	dst = pkcs5UnPadding(dst, block.BlockSize())
	return dst, nil
}

func pkcs5UnPadding(src []byte, blockSize int) []byte {
	n := len(src)
	paddingNum := int(src[n-1])
	if n < paddingNum || paddingNum > blockSize {
		return src
	}
	return src[:n-paddingNum]
}

func des3Decrypt(key, iv []byte, src []byte) ([]byte, error) {
	block, err := des.NewTripleDESCipher(key)
	if err != nil {
		return nil, err
	}
	blockMode := cipher.NewCBCDecrypter(block, iv)
	sq := make([]byte, len(src))
	blockMode.CryptBlocks(sq, src)
	return pkcs5UnPadding(sq, block.BlockSize()), nil
}

func paddingZero(s []byte, l int) []byte {
	h := l - len(s)
	if h <= 0 {
		return s
	}
	for i := len(s); i < l; i++ {
		s = append(s, 0)
	}
	return s
}

func GeckoDecryptCreditCardNUmber(encryptedNumber string, masterKey []byte) (plaintext []byte, err error) {
	decodedNumber, err := base64.StdEncoding.DecodeString(encryptedNumber)
	if err != nil {
		return
	}
	iv := decodedNumber[:12]
	ciphertext := decodedNumber[12:]
	block, err := aes.NewCipher(masterKey)
	if err != nil {
		return
	}
	mode, err := cipher.NewGCM(block)
	if err != nil {
		return
	}
	plaintext, err = mode.Open(nil, iv, ciphertext, nil)
	return
}
