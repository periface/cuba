package db

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"
	_ "github.com/mattn/go-sqlite3" // Import the SQLite driver
	"github.com/periface/cuba/internals/utils"
)

type DB struct {
	Database *sql.DB
}

var (
	dbOnce     sync.Once
	dbInstance *DB
)

func GetDBFilename() string {
	FILENAME, err := utils.GetEnvVariable("DB_FILENAME")
	if err != nil {
		fmt.Println("Error getting DB_FILENAME from environment:", err)
		return "../wero/proveedores.db" // Default filename if not set in environment
	}
	if FILENAME == "" {
		FILENAME = "../wero/proveedores.db" // Default filename if not set in environment
	}
	return FILENAME
}
func GetDBInstance() (*DB, error) {
	FILENAME := GetDBFilename()
	connectionString := "file:{DB_FILENAME}?cache=shared&mode=rwc"

	dbOnce.Do(func() {
		database, err := sql.Open("sqlite3", strings.ReplaceAll(connectionString, "{DB_FILENAME}", FILENAME))
		if err != nil {
			fmt.Println("Error opening database:", err)
			return
		}
		dbInstance = &DB{Database: database}
	})

	if dbInstance == nil {
		return nil, fmt.Errorf("database instance is not initialized")
	}

	return dbInstance, nil
}

