package main

import (
	"fmt"
	"gSmudgeAPI/handler"
	"net/http"
)

func main() {
	http.HandleFunc("/", handler.HandlerIndex)

	fmt.Print("Starting gSmudgeAPI server on port 6969\n")
	err := http.ListenAndServe(":6969", nil)
	if err != nil {
		panic(err)
	}
}
