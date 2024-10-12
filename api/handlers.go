package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	nextDate "github.com/LingL42/finalGoProject/dateFunction"
	dataBase "github.com/LingL42/finalGoProject/db"
)

func NextDateHandler(w http.ResponseWriter, r *http.Request) {
	nowStr := r.URL.Query().Get("now")
	dateStr := r.URL.Query().Get("date")
	repeatStr := r.URL.Query().Get("repeat")

	if !ValidateDate(w, nowStr) {
		return
	}
	now, _ := time.Parse(nextDate.DateFormat, nowStr)

	nextDate, err := nextDate.NextDate(now, dateStr, repeatStr)
	response := Response{}

	if err != nil {
		response.Error = err.Error()
		http.Error(w, "Ошибка некстДейт "+response.Error, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, nextDate)

}

func PostTaskHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response := Response{}
		var task dataBase.Task
		err := json.NewDecoder(r.Body).Decode(&task)
		if err != nil {
			response.Error = "Неверный формат json"
			w.Header().Set("Content-Type", "application/json")

			json.NewEncoder(w).Encode(response)
			return
		}

		if !validateTitle(w, task.Title) {
			return
		}

		now := time.Now()

		if task.Date == "" {
			task.Date = now.Format(nextDate.DateFormat)
		} else {
			if !ValidateDate(w, task.Date) {
				return
			}

			taskDate, _ := time.Parse(nextDate.DateFormat, task.Date)

			if taskDate.Before(now) {
				if task.Repeat == "" {
					task.Date = now.Format(nextDate.DateFormat)
				} else {
					nextDate, err := nextDate.NextDate(now, task.Date, task.Repeat)
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

		id, err := dataBase.AddTaskInDB(db, task)
		if err != nil {
			http.Error(w, "Ошибка при добавлении задачи в бд", http.StatusInternalServerError)
			return
		}
		response = Response{ID: strconv.Itoa(int(id))}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}

}

func GetTasksHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := dataBase.GetTasks(db)
		if err != nil {
			http.Error(w, fmt.Sprintf("Ошибка при запросу задач из БД: %v", err), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var tasks []dataBase.Task
		for rows.Next() {
			var task dataBase.Task
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
			response.Tasks = []dataBase.Task{}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

func GetTaskHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idParam := r.URL.Query().Get("id")

		if !validateIdParam(w, idParam) {
			return
		}

		id, err := strconv.Atoi(idParam)
		if err != nil {
			writeErrorResponse(w, http.StatusBadRequest, "Неверный формат идентификатора")
			return
		}

		task, err := dataBase.GetTaskById(db, id)
		if err == dataBase.ErrTaskNotFound {
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
		var task dataBase.Task
		err := json.NewDecoder(r.Body).Decode(&task)
		if err != nil {
			http.Error(w, `{"error": "Неверный формат данных"}`, http.StatusBadRequest)
			return
		}

		if !validateIdParam(w, task.ID) {
			return
		}

		if !ValidateDate(w, task.Date) {
			return
		}

		_, err = nextDate.NextDate(time.Now(), task.Date, task.Repeat)
		if err != nil {
			http.Error(w, `{"error": "Ощибка в NextDate"}`, http.StatusBadRequest)
			return
		}

		if !validateTitle(w, task.Title) {
			return
		}

		result, err := dataBase.UpdateTask(db, task)

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

func MarkTaskAsDoneHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")

		if !validateIdParam(w, id) {
			return
		}

		idForDB, _ := strconv.Atoi(id)
		task, err := dataBase.GetTaskById(db, idForDB)
		if err == dataBase.ErrTaskNotFound {
			writeErrorResponse(w, http.StatusNotFound, `{"error": "Задача не найдена"}`)
			return
		} else if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, `{"error": "Ошибка базы данных"}`)
			return
		}

		if task.Repeat == "" {
			_, err := dataBase.DeleteTask(db, id)
			if err != nil {
				http.Error(w, `{"error": "Ошибка при удалении задачи"}`, http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{}`))
			return
		}

		actualDate, _ := time.Parse(nextDate.DateFormat, task.Date)
		nextDate, err := nextDate.NextDate(actualDate.AddDate(0, 0, 1), task.Date, task.Repeat)

		if err != nil {
			http.Error(w, `{"error": "Ошибка при обновлении задачи"}`, http.StatusInternalServerError)
			return
		}

		err = dataBase.UpdateTaskDate(db, nextDate, id)
		if err != nil {
			http.Error(w, `{"error": "Ошибка при обновлении задачи"}`, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{}`))
	}
}

func DeleteTaskHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")

		if !validateIdParam(w, id) {
			return
		}

		result, err := dataBase.DeleteTask(db, id)
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

func TaskHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			PostTaskHandler(db)(w, r)
		case http.MethodGet:
			GetTaskHandler(db)(w, r)
		case http.MethodPut:
			PutTaskHandler(db)(w, r)
		case http.MethodDelete:
			DeleteTaskHandler(db)(w, r)
		default:
			http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		}
	}
}

type Response struct {
	ID    string `json:"id,omitempty"`
	Error string `json:"error,omitempty"`
}

type TasksResponse struct {
	Tasks []dataBase.Task `json:"tasks"`
	Error string          `json:"error,omitempty"`
}
