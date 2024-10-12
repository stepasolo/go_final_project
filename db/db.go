package db

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

const limit int = 10

func DbWorker() {

	appPath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	dbFile := filepath.Join(filepath.Dir(appPath), "./scheduler.db")
	_, err = os.Stat("./scheduler.db")

	var install bool
	if err != nil {
		install = true
	}

	if install {
		createDB(dbFile)
	}
}

func createDB(dbFile string) {
	db, err := sql.Open("sqlite3", "./scheduler.db")
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

func OpenDB() *sql.DB {
	db, err := sql.Open("sqlite3", "./scheduler.db")
	if err != nil {
		log.Fatal(err)

	}
	//defer db.Close()
	return db
}

var ErrTaskNotFound = errors.New("task not found")

func GetTaskById(db *sql.DB, id int) (Task, error) {
	var task Task
	query := "SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?"
	err := db.QueryRow(query, id).Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)

	if err == sql.ErrNoRows {
		return Task{}, ErrTaskNotFound
	} else if err != nil {
		return Task{}, err
	}

	return task, nil

}

func GetTasks(db *sql.DB) (*sql.Rows, error) {
	return db.Query("SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date ASC LIMIT ?", limit)
}

func DeleteTask(db *sql.DB, id string) (sql.Result, error) {
	result, err := db.Exec("DELETE FROM scheduler WHERE id = ?", id)
	if err != nil {
		return result, err
	}
	return result, nil
}

func UpdateTaskDate(db *sql.DB, nextDate string, id string) error {
	_, err := db.Exec("UPDATE scheduler SET date = ? WHERE id = ?", nextDate, id)
	if err != nil {
		return err
	}
	return nil
}

func AddTaskInDB(db *sql.DB, task Task) (int64, error) {
	taskInBD, err := db.Exec("INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)", task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		return 0, fmt.Errorf("что то не так при добавлении в бд: %v", err)
	}
	id, err := taskInBD.LastInsertId()
	return id, err
}

func UpdateTask(db *sql.DB, task Task) (sql.Result, error) {
	query := "UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?"
	result, err := db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		return result, err
	}
	return result, nil
}

type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}
