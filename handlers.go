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

func getTasksHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date ASC LIMIT 10")
		if err != nil {
			http.Error(w, fmt.Sprintf("Ошибка при запросу задач из БД: %v", err), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var tasks []Task
		for rows.Next() {
			var task Task
			err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
			if err != nil {
				http.Error(w, fmt.Sprintf("Ошибка при чтении данных задачи: %v", err), http.StatusInternalServerError)
				return
			}
			tasks = append(tasks, task)
		}

		response := TasksResponse{
			Tasks: tasks,
		}
		if len(tasks) == 0 {
			response.Tasks = []Task{}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

type Task struct {
	ID      string `json:"id"` //opasno
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

type TasksResponse struct {
	Tasks []Task `json:"tasks"`
	Error string `json:"error,omitempty"`
}
