package database

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func CreateDB() (TaskStore, error) {
	var ts TaskStore
	ts = ts.NewTaskStore(ts.DB)
	ts = ts.OpenDB(DbPath)
	appPath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	dbFile := filepath.Join(filepath.Dir(appPath), DbPath)
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

	install = true

	if install {
		CreateTable(ts)
		return ts, nil
	} else {
		fmt.Println("Error create table")
		return ts, err
	}
}

func CreateTable(ts TaskStore) {
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS scheduler (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		date TEXT,
		title TEXT,
		comment TEXT,
		repeat TEXT
	);`

	_, err := ts.DB.Exec(createTableSQL)
	if err != nil {
		fmt.Printf("error create table: %v", err)
	}

	createIndexSQL := `
	CREATE INDEX IF NOT EXISTS idx_date ON scheduler(date);
	`

	_, err = ts.DB.Exec(createIndexSQL)
	if err != nil {
		fmt.Printf("error create index: %v", err)
	}

	fmt.Println("Table and index created successfully")
}
