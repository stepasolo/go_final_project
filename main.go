package main

import (
	"fmt"
	"net/http"
	"os"

	myApi "github.com/LingL42/finalGoProject/api"
	dataBase "github.com/LingL42/finalGoProject/db"
)

func main() {
	var port string
	dataBase.DbWorker()
	db := dataBase.OpenDB()
	defer db.Close()
	fmt.Println("Запускаем сервер")
	http.HandleFunc("/api/nextdate", myApi.NextDateHandler)
	osPort := os.Getenv("TODO_PORT")
	if osPort != "" {
		//port := fmt.Sprintf(`:%s`, os.Getenv("TODO_PORT"))
		port = fmt.Sprintf(`0.0.0.0:%s`, os.Getenv("TODO_PORT")) //для работы через WSL
	} else {
		port = ":80"
	}

	fmt.Println(port)

	http.HandleFunc("/api/tasks", myApi.GetTasksHandler(db))
	http.HandleFunc("/api/task", myApi.TaskHandler(db))
	http.HandleFunc("/api/task/done", myApi.MarkTaskAsDoneHandler(db))

	webDir := "./web"
	http.Handle("/", http.FileServer(http.Dir(webDir)))
	err := http.ListenAndServe(port, nil)
	if err != nil {
		panic(err)
	}

}
