package utils

import (
	"database/sql"
	"fmt"
	"os"

	"service/log"

	_ "github.com/go-sql-driver/mysql"
)

type ModDeveloper struct {
	Username string `json:"username"`
	IsOwner  bool   `json:"is_owner"`
}

type Mod struct {
	ID         string         `json:"id"`
	Developers []ModDeveloper `json:"developers"`
}

type ModRequest struct {
	Error   string `json:"error"`
	Payload Mod    `json:"payload"`
}

// Concurrent database connection
var data *sql.DB

// safely prepare the sql statement
func PrepareStmt(db *sql.DB, sql string) (*sql.Stmt, error) {
	if db != nil {
		log.Debug("Preparing connection for statement %s", sql)
		return db.Prepare(sql)
	} else {
		return nil, fmt.Errorf("database connection non-existent")
	}
}

func Db() *sql.DB {
	return data
}

func init() {
	var err error

	uri := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_NAME"),
	)

	log.Info("Connecting to database with URI: %s", uri)
	data, err = sql.Open("mysql", uri)
	if err != nil {
		log.Error("Failed to establish MariaDB connection: %s", err.Error())
		return
	}

	err = data.Ping()
	if err != nil {
		log.Error("Failed to ping database: %s", err.Error())
		return
	} else if data == nil {
		log.Error("Database connection is nil")
		return
	}

	log.Print("MariaDB connection established.")
}
