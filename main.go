package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	fmt.Println("Zapuskaem server")
	port := fmt.Sprintf(`:%s`, os.Getenv("TODO_PORT"))

	webDir := "./web"
	http.Handle("/", http.FileServer(http.Dir(webDir)))
	err := http.ListenAndServe(port, nil)
	if err != nil {
		panic(err)
	}
}
