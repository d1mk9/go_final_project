package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

var db *sql.DB

func StartDB() *sql.DB {
	db, err := sql.Open("sqlite", "scheduler.db")
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return db
}

func CreateDB() {
	StartDB()
	appPath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	dbFile := filepath.Join(filepath.Dir(appPath), "scheduler.db")
	_, err = os.Stat(dbFile)

	var install bool
	if err != nil {
		if os.IsNotExist(err) {
			log.Println("DB creating")
			install = true
		} else {
			log.Println("DB create error")
			log.Fatal(err)
		}
	}
	log.Println("DB created before")

	if install {
		createTable()
	} else {
		fmt.Println("Error create table")
	}
}

func createTable() {
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS scheduler (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		date TEXT,
		title TEXT,
		comment TEXT,
		repeat TEXT
	);`

	_, err := db.Exec(createTableSQL)
	if err != nil {
		fmt.Printf("error create table: %v", err)
	}

	createIndexSQL := `
	CREATE INDEX IF NOT EXISTS idx_date ON scheduler(date);
	`

	_, err = db.Exec(createIndexSQL)
	if err != nil {
		fmt.Printf("error create index: %v", err)
	}

	fmt.Println("Table and index created successfully")
}
