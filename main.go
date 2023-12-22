package main

import (
	"context"
	"fmt"
	"gSmudgeAPI/cache"
	"gSmudgeAPI/handler"
	"gSmudgeAPI/handler/instagram"
	"gSmudgeAPI/handler/tiktok"
	"gSmudgeAPI/handler/twitter"
	"net/http"
	"regexp"
	"strings"
)

func cacheMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var key string
		if strings.HasPrefix(r.RequestURI, "/twitter?") || strings.HasPrefix(r.RequestURI, "/x?") {
			key = (regexp.MustCompile((`.*(?:twitter|x).com/.+status/([A-Za-z0-9]+)`))).FindStringSubmatch(r.URL.Query().Get("url"))[1]
		} else if strings.HasPrefix(r.RequestURI, "/instagram?") {
			key = (regexp.MustCompile((`(?:reel|p)/([A-Za-z0-9_-]+)`))).FindStringSubmatch(r.URL.Query().Get("url"))[1]
		} else if strings.HasPrefix(r.RequestURI, "/tiktok?") {
			resp, _ := http.Get(r.URL.Query().Get("url"))
			key = (regexp.MustCompile((`/(?:video|v)/(\d+)`))).FindStringSubmatch(resp.Request.URL.String())[1]
		} else {
			key = r.RequestURI
		}

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
	http.HandleFunc("/tiktok", cacheMiddleware(tiktok.TikTokIndexer))

	fmt.Print("Starting gSmudgeAPI server on port 6969\n")
	err := http.ListenAndServe(":6969", nil)
	if err != nil {
		panic(err)
	}
}
