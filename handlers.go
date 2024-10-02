package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

func addTaskInDB(db *sql.DB, task Task) (int64, error) {
	taskInBD, err := db.Exec("INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)", task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		log.Fatal("Что то не таку при добавлении в бд", err)
	}
	id, err := taskInBD.LastInsertId()
	return id, err
}

func PostTaskHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response := Response{}
		var task Task
		err := json.NewDecoder(r.Body).Decode(&task)
		if err != nil {
			response.Error = "Неверный формат джсон"
			w.Header().Set("Content-Type", "application/json")

			json.NewEncoder(w).Encode(response)
			return
		}

		if task.Title == "" {
			response.Error = "Поле тайтл обязательно"
			w.Header().Set("Content-Type", "application/json")

			json.NewEncoder(w).Encode(response)
			return
		}

		now := time.Now()
		dateFormat := "20060102"

		if task.Date == "" {
			task.Date = now.Format(dateFormat)
		} else {
			taskDate, err := time.Parse(dateFormat, task.Date)
			if err != nil {
				response.Error = "Неверный формат даты"
				w.Header().Set("Content-Type", "application/json")

				json.NewEncoder(w).Encode(response)
				return
			}

			if taskDate.Before(now) {
				if task.Repeat == "" {
					task.Date = now.Format(dateFormat)
				} else {
					nextDate, err := NextDate(now, task.Date, task.Repeat)
					if err != nil {
						response.Error = fmt.Sprintf("Ощибка в NextDate:%v", err)
						w.Header().Set("Content-Type", "application/json")

						json.NewEncoder(w).Encode(response)
						return

					}
					task.Date = nextDate
				}

			}
		}

		id, err := addTaskInDB(db, task)
		if err != nil {
			fmt.Println(task)
			fmt.Println(err)
			http.Error(w, "Ошибка при добавлении задачи в бд", http.StatusInternalServerError)
			return
		}
		//response := Response{Result: "Задача успешно добавлена"}
		response = Response{ID: strconv.Itoa(int(id))}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}

}

type Task struct {
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

type Response struct {
	ID string `json:"id,omitempty"`
	//	Result string `json:"result"`

	Error string `json:"error,omitempty"`
}
