package dbHelper

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

func DbWorker() {

	appPath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	dbFile := filepath.Join(filepath.Dir(appPath), "db/scheduler.db")
	_, err = os.Stat("db/scheduler.db")

	var install bool
	if err != nil {
		install = true
	}

	if install {
		createDB(dbFile)
	}
}

func createDB(dbFile string) {
	fmt.Println(dbFile + "wot put")
	db, err := sql.Open("sqlite3", "db/scheduler.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	createTableSQL := `
	CREATE TABLE IF NOT EXISTS scheduler (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	date TEXT NOT NULL,
	title TEXT NOT NULL,
	comment TEXT,
	repeat TEXT CHECK(LENGTH(repeat) <= 128)
	);
	CREATE INDEX idx_date ON scheduler (date);
	`
	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Таблица scheduler и индекс по полю date созданы.")
}
