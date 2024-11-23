package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pterm/pterm"
)

type ConnectionData struct {
	ID                string
	LastHeartbeatTime string
	ConnectionTime    string
}

func CreateTables() bool {
	defer recoverFromPanic()

	const maxRetries = 3
	var attempt int

	for attempt = 1; attempt <= maxRetries; attempt++ {
		log.Printf("Attempt %d to create tables...\n", attempt)
		success := attemptCreateTables()
		if success {
			log.Println("Tables created successfully!")
			return true
		}
		log.Println("Retrying table creation...")
		time.Sleep(1 * time.Second) 
	}

	log.Printf("Failed to create tables after %d attempts.\n", maxRetries)
	return false
}

func attemptCreateTables() bool {
	database, err := sql.Open("sqlite3", "database.db")
	if err != nil {
		log.Println("Error opening database:", err)
		return false
	}
	defer database.Close()

	_, err = database.Exec(`
		CREATE TABLE IF NOT EXISTS connections (
			id VARCHAR(255) PRIMARY KEY, 
			last_heartbeat_time VARCHAR(255), 
			connection_time VARCHAR(255)
		);
		CREATE TABLE IF NOT EXISTS events (
			recipient VARCHAR(255), 
			type VARCHAR(100), 
			extra VARCHAR(500)
		);
		CREATE TABLE IF NOT EXISTS event_responses (
			sender VARCHAR(255), 
			response VARCHAR(255)
		);
	`)
	if err != nil {
		log.Println("Error creating tables:", err)
		return false
	}

	return true
}

func ConnectionNew(ID string) bool {
	defer recoverFromPanic()

	database, err := sql.Open("sqlite3", "database.db")
	if err != nil {
		log.Println("Error opening database:", err)
		return false
	}
	defer database.Close()

	_, err = database.Exec(
		"INSERT INTO connections (id, last_heartbeat_time, connection_time) VALUES (?, ?, ?)",
		ID, time.Now().Format(time.RFC3339), time.Now().Format(time.RFC3339),
	)

	if err != nil {
		log.Println("Error inserting new connection:", err)
		return false
	}

	return true
}

func GetConnectionData(ID string) string {
	defer recoverFromPanic()

	db, err := sql.Open("sqlite3", "database.db")
	if err != nil {
		log.Fatal("Error opening database:", err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT id, last_heartbeat_time FROM connections WHERE id = ?", ID)
	if err != nil {
		log.Println("Error querying database:", err)
		return ""
	}
	defer rows.Close()

	var connectionData ConnectionData
	for rows.Next() {
		err := rows.Scan(&connectionData.ID, &connectionData.LastHeartbeatTime)
		if err != nil {
			pterm.Fatal.WithFatal(true).Println("Error scanning row:", err)
			continue
		}
		return connectionData.LastHeartbeatTime
	}

	if err = rows.Err(); err != nil {
		log.Println("Row iteration error:", err)
	}

	return ""
}

func GetAllAgents() ([]ConnectionData, error) {
	defer recoverFromPanic()

	var agents []ConnectionData

	database, err := sql.Open("sqlite3", "database.db")
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}
	defer database.Close()

	rows, err := database.Query("SELECT id, last_heartbeat_time, connection_time FROM connections")
	if err != nil {
		return nil, fmt.Errorf("error querying database: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var agent ConnectionData
		if err := rows.Scan(&agent.ID, &agent.LastHeartbeatTime, &agent.ConnectionTime); err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}
		agents = append(agents, agent)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return agents, nil
}

func recoverFromPanic() {
	if r := recover(); r != nil {
		log.Println("Recovered from panic:", r)
	}
}
