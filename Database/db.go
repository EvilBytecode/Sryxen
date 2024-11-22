package database

import (
	"database/sql"
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
	database, err := sql.Open("sqlite3", "database.db")

	if err != nil {
		return false
	} else {
		_, err := database.Exec(`
                                 CREATE TABLE IF NOT EXISTS connections(id VARCHAR(255), last_heartbeat_time VARCHAR(255), connection_time VARCHAR(255));
								 CREATE TABLE IF NOT EXISTS events(recipient VARCHAR(255), type VARCHAR(100), extra VARCHAR(500));
							     CREATE TABLE IF NOT EXISTS event_responses(sender VARCHAR(255), response VARCHAR(255))`)

		if err != nil {
			return false
		} else {
			return true
		}
	}
}

func ConnectionNew(ID string) bool {
	database, err := sql.Open("sqlite3", "database.db")

	if err != nil {
		return false
	} else {
		_, err := database.Exec("INSERT INTO connections(id, last_heartbeat_time, connection_time) VALUES (?,?,?)", ID, time.Now(), time.Now())

		if err != nil {
			return false
		} else {
			return true
		}
	}
}

func GetConnectionData(ID string) string {
	db, err := sql.Open("sqlite3", "database.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT id, last_heartbeat_time FROM connections WHERE id = ?", ID)

	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var connectionData ConnectionData

		err := rows.Scan(&connectionData.ID, &connectionData.LastHeartbeatTime)

		if err != nil {
			pterm.Fatal.WithFatal(true).Println(err)
		} else {
			return connectionData.LastHeartbeatTime
		}

	}
	return ""
}



func GetAllAgents() ([]ConnectionData, error) {
    var agents []ConnectionData

    database, err := sql.Open("sqlite3", "database.db")
    if err != nil {
        return nil, err
    }
    defer database.Close()

    rows, err := database.Query("SELECT id, last_heartbeat_time, connection_time FROM connections")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    for rows.Next() {
        var agent ConnectionData
        if err := rows.Scan(&agent.ID, &agent.LastHeartbeatTime, &agent.ConnectionTime); err != nil {
            return nil, err
        }
        agents = append(agents, agent)
    }

    if err := rows.Err(); err != nil {
        return nil, err
    }

    return agents, nil
}
