package main

import (
	"fmt"
	"net/http"
)

func main() {
	fmt.Println("Zapuskaem server")

	webDir := "./web"
	fmt.Println("Завершаем работу")
	http.Handle("/", http.FileServer(http.Dir(webDir)))
	err := http.ListenAndServe(":7540", nil)
	if err != nil {
		panic(err)
	}
}
