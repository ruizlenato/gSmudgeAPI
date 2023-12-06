package main

import (
	"fmt"
	"gSmudgeAPI/handler"
	"gSmudgeAPI/handler/instagram"
	"net/http"
)

func main() {
	http.HandleFunc("/", handler.HandlerIndex)
	http.HandleFunc("/instagram", instagram.InstagramIndexer)

	fmt.Print("Starting gSmudgeAPI server on port 6969\n")
	err := http.ListenAndServe(":6969", nil)
	if err != nil {
		panic(err)
	}
}
