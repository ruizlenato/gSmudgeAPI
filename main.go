package main

import (
	"context"
	"fmt"
	"gSmudgeAPI/cache"
	"gSmudgeAPI/handler"
	"gSmudgeAPI/handler/instagram"
	"gSmudgeAPI/handler/tiktok"
	"gSmudgeAPI/handler/twitter"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/valyala/fasthttp"
)

func cacheMiddleware(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		var key string
		if strings.HasPrefix(string(ctx.RequestURI()), "/twitter?") || strings.HasPrefix(string(ctx.RequestURI()), "/x?") {
			key = (regexp.MustCompile((`.*(?:twitter|x).com/.+status/([A-Za-z0-9]+)`))).FindStringSubmatch(string(ctx.QueryArgs().Peek("url")))[1]
		} else if strings.HasPrefix(string(ctx.RequestURI()), "/instagram?") {
			key = (regexp.MustCompile((`(?:reel|p)/([A-Za-z0-9_-]+)`))).FindStringSubmatch(string(ctx.QueryArgs().Peek("url")))[1]
		} else if strings.HasPrefix(string(ctx.RequestURI()), "/tiktok?") {
			resp, _ := http.Get(string(ctx.QueryArgs().Peek("url")))
			key = (regexp.MustCompile((`/(?:video|v)/(\d+)`))).FindStringSubmatch(resp.Request.URL.String())[1]
		} else {
			key = string(ctx.RequestURI())
		}

		cachedResponse, err := cache.GetRedisClient().Get(context.Background(), key).Bytes()

		if err == nil {
			ctx.Response.Header.Add("Content-Type", "application/json")
			ctx.Write(cachedResponse)
		} else {
			next(ctx)
		}
	}
}

func main() {
	requestHandler := func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/":
			handler.HandlerIndex(ctx)
		case "/instagram":
			cacheMiddleware(instagram.InstagramIndexer)(ctx)
		case "/twitter":
			twitter.TwitterIndexer(ctx)
		case "/x":
			twitter.TwitterIndexer(ctx)
		case "/tiktok":
			tiktok.TikTokIndexer(ctx)
		default:
			ctx.Error("Not Found", fasthttp.StatusNotFound)
		}
	}

	// Iniciando o servidor fasthttp
	fmt.Print("Starting gSmudgeAPI server on port 6969\n")
	if err := fasthttp.ListenAndServe(":6969", requestHandler); err != nil {
		log.Fatalf("Erro ao iniciar o servidor: %s", err)
	}
}
