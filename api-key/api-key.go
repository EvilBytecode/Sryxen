package apiKey

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
	"golang.org/x/crypto/scrypt"
)

const apiKeyFile = "api_key.json"

var (
	currentAPIKey string
	apiKeyMutex   sync.Mutex
)

type APIKeyData struct {
	Key string `json:"key"`
}

func GenerateAPIKey() string {
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil {
		log.Fatalf("Failed to generate salt: %v", err)
	}

	password := make([]byte, 32)
	_, err = rand.Read(password)
	if err != nil {
		log.Fatalf("Failed to generate password for API key: %v", err)
	}

	key, err := scrypt.Key(password, salt, 16384, 8, 1, 32)
	if err != nil {
		log.Fatalf("Failed to generate API key: %v", err)
	}

	apiKey := "SRYXEN_" + base64.RawURLEncoding.EncodeToString(key)
	return apiKey
}

func SaveAPIKeyToFile(apiKey string) error {
	apiKeyData := APIKeyData{Key: apiKey}
	data, err := json.Marshal(apiKeyData)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(apiKeyFile, data, 0644)
}

func LoadAPIKeyFromFile() (string, error) {
	data, err := ioutil.ReadFile(apiKeyFile)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil 
		}
		return "", err
	}

	var apiKeyData APIKeyData
	if err := json.Unmarshal(data, &apiKeyData); err != nil {
		return "", err
	}
	return apiKeyData.Key, nil
}

func InitializeAPIKey() (string, error) {
	apiKeyMutex.Lock()
	defer apiKeyMutex.Unlock()

	var err error
	currentAPIKey, err = LoadAPIKeyFromFile()
	if err != nil {
		return "", err
	}

	if currentAPIKey == "" {
		currentAPIKey = GenerateAPIKey()
		if err := SaveAPIKeyToFile(currentAPIKey); err != nil {
			return "", err
		}
	}

	return currentAPIKey, nil
}

func GetCurrentAPIKey() (string, error) {
	apiKeyMutex.Lock()
	defer apiKeyMutex.Unlock()

	if currentAPIKey != "" {
		return currentAPIKey, nil
	}

	apiKey, err := LoadAPIKeyFromFile()
	if err != nil {
		return "", fmt.Errorf("error loading API key from file: %v", err)
	}

	if apiKey == "" {
		return "", errors.New("API key is not initialized and not found in the file")
	}

	currentAPIKey = apiKey
	return currentAPIKey, nil
}


func ValidateAPIKey(r *http.Request) bool {
	apiKey := r.Header.Get("X-API-Key")
	if apiKey == "" {
		log.Println("API key missing in request headers.")
		return false
	}
	log.Printf("Received API Key: %s\n", apiKey)

	currentKey, err := GetCurrentAPIKey()
	if err != nil {
		log.Println("Error loading current API key:", err)
		return false
	}
	log.Printf("Stored API Key: %s\n", currentKey)

	if apiKey != currentKey {
		log.Println("API key does not match.")
		return false
	}
	return true
}
