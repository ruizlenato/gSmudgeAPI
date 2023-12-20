package main

import (
	"context"
	"fmt"
	"gSmudgeAPI/cache"
	"gSmudgeAPI/handler"
	"gSmudgeAPI/handler/instagram"
	"gSmudgeAPI/handler/twitter"
	"net/http"
)

func cacheMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.RequestURI
		cachedResponse, err := cache.GetRedisClient().Get(context.Background(), key).Bytes()

		if err == nil {
			w.Header().Add("Content-Type", "application/json")
			w.Write(cachedResponse)
		} else {
			next.ServeHTTP(w, r)
		}
	}
}

func main() {
	http.HandleFunc("/", handler.HandlerIndex)
	http.HandleFunc("/instagram", cacheMiddleware(instagram.InstagramIndexer))
	http.HandleFunc("/twitter", cacheMiddleware(twitter.TwitterIndexer))
	http.HandleFunc("/x", cacheMiddleware(twitter.TwitterIndexer))

	fmt.Print("Starting gSmudgeAPI server on port 6969\n")
	err := http.ListenAndServe(":6969", nil)
	if err != nil {
		panic(err)
	}
}
