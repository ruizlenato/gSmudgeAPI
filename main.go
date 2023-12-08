package main

import (
	"fmt"
	"gSmudgeAPI/handler"
	"gSmudgeAPI/handler/instagram"
	"gSmudgeAPI/handler/twitter"
	"net/http"
)

func main() {
	http.HandleFunc("/", handler.HandlerIndex)
	http.HandleFunc("/instagram", instagram.InstagramIndexer)
	http.HandleFunc("/twitter", twitter.TwitterIndexer)
	http.HandleFunc("/x", twitter.TwitterIndexer)

	fmt.Print("Starting gSmudgeAPI server on port 6969\n")
	err := http.ListenAndServe(":6969", nil)
	if err != nil {
		panic(err)
	}
}
