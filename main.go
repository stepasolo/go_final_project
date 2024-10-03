package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	dbHelper "github.com/LingL42/finalGoProject/db"
)

func main() {
	dbHelper.DbWorker()
	db := dbHelper.OpenDB()
	fmt.Println("Zapuskaem server")
	http.HandleFunc("/api/nextdate", nextDateHandler)
	//port := fmt.Sprintf(`:%s`, os.Getenv("TODO_PORT"))
	port := fmt.Sprintf(`0.0.0.0:%s`, os.Getenv("TODO_PORT")) //для работы через WSL
	fmt.Println(port)

	http.HandleFunc("/api/tasks", getTasksHandler(db))
	http.HandleFunc("/api/task", taskHandler(db))

	webDir := "./web"
	http.Handle("/", http.FileServer(http.Dir(webDir)))
	err := http.ListenAndServe(port, nil)
	if err != nil {
		panic(err)
	}

}

func nextDateHandler(w http.ResponseWriter, r *http.Request) {
	nowStr := r.URL.Query().Get("now")
	dateStr := r.URL.Query().Get("date")
	repeatStr := r.URL.Query().Get("repeat")

	now, err := time.Parse("20060102", nowStr)
	if err != nil {
		http.Error(w, "Неверный now", http.StatusBadRequest)
		return
	}

	nextDate, err := NextDate(now, dateStr, repeatStr)
	response := Response{}

	if err != nil {
		response.Error = err.Error()
		http.Error(w, "Ошибка некстДейт", http.StatusBadRequest)
		return
	} //else {
	// 	response.Result = nextDate
	// }

	w.Header().Set("Content-Type", "application/json")
	//json.NewEncoder(w).Encode(response)
	fmt.Fprint(w, nextDate)

}

// type Response struct {
// 	Result string `json:"result"`
// 	Error  string `json:"error,omitempty"`
// }
