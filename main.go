package main

import (
	"fmt"
	"net/http"
	"os"

	dbHelper "github.com/LingL42/finalGoProject/db"
)

func main() {
	dbHelper.DbWorker()
	fmt.Println("Zapuskaem server")
	//port := fmt.Sprintf(`:%s`, os.Getenv("TODO_PORT"))
	port := fmt.Sprintf(`0.0.0.0:%s`, os.Getenv("TODO_PORT")) //для работы через WSL
	fmt.Println(port)

	webDir := "./web"
	http.Handle("/", http.FileServer(http.Dir(webDir)))
	err := http.ListenAndServe(port, nil)
	if err != nil {
		panic(err)
	}

}
