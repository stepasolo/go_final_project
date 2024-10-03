package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	dbHelper "github.com/LingL42/finalGoProject/db"
)

func addTaskInDB(db *sql.DB, task dbHelper.Task) (int64, error) {
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
		var task dbHelper.Task
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

		var tasks []dbHelper.Task
		for rows.Next() {
			var task dbHelper.Task
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
			response.Tasks = []dbHelper.Task{}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

func GetTaskHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idParam := r.URL.Query().Get("id")
		if idParam == "" {
			writeErrorResponse(w, http.StatusBadRequest, "Не указан идентификатор")
			return
		}
		id, err := strconv.Atoi(idParam)
		if err != nil {
			writeErrorResponse(w, http.StatusBadRequest, "Неверный формат идентификатора")
			return
		}

		task, err := dbHelper.GetTaskById(db, id)
		if err == dbHelper.ErrTaskNotFound {
			writeErrorResponse(w, http.StatusNotFound, "Задача не найдена")
			return
		} else if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "Ошибка при поиске задачи")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(task)
	}
}

func PutTaskHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var task dbHelper.Task
		err := json.NewDecoder(r.Body).Decode(&task)
		if err != nil {
			http.Error(w, `{"error": "Неверный формат данных"}`, http.StatusBadRequest)
			return
		}
		if task.ID == "" {
			http.Error(w, `{"error": "Не указан идентификатор"}`, http.StatusBadRequest)
			return
		}

		_, err = time.Parse("20060102", task.Date)
		if err != nil {
			http.Error(w, `{"error": "Неверный формат даты"}`, http.StatusBadRequest)
			return
		}

		_, err = NextDate(time.Now(), task.Date, task.Repeat)
		if err != nil {
			http.Error(w, `{"error": "Ощибка в NextDate"}`, http.StatusBadRequest)
			return
		}

		if task.Title == "" {
			http.Error(w, `{"error": "Поле тайтл не может быть пустым"}`, http.StatusBadRequest)
			return
		}
		query := "UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?"
		result, err := db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat, task.ID)
		if err != nil {
			http.Error(w, `{"error": "Ошибка базы данных"}`, http.StatusInternalServerError)
			return
		}
		rowsAffected, err := result.RowsAffected()
		if err != nil || rowsAffected == 0 {
			http.Error(w, `{"error": "Задача не найдена"}`, http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{}`))

	}
}

func writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func taskHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			PostTaskHandler(db)(w, r)
		case http.MethodGet:
			GetTaskHandler(db)(w, r)
		case http.MethodPut:
			PutTaskHandler(db)(w, r)
		default:
			http.Error(w, "Методж не поддерживается", http.StatusMethodNotAllowed)
		}
	}
}

type Response struct {
	ID    string `json:"id,omitempty"`
	Error string `json:"error,omitempty"`
}

type TasksResponse struct {
	Tasks []dbHelper.Task `json:"tasks"`
	Error string          `json:"error,omitempty"`
}
