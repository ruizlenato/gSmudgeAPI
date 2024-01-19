package main

import (
	"context"
	"fmt"
	"gSmudgeAPI/cache"
	"gSmudgeAPI/handler"
	"gSmudgeAPI/handler/instagram"
	"gSmudgeAPI/handler/tiktok"
	"gSmudgeAPI/handler/twitter"
	"gSmudgeAPI/utils"
	"log"
	"regexp"
	"strings"

	"github.com/valyala/fasthttp"
)

func cacheMiddleware(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		var matches []string
		url := string(ctx.QueryArgs().Peek("url"))
		if strings.HasPrefix(string(ctx.RequestURI()), "/twitter?") || strings.HasPrefix(string(ctx.RequestURI()), "/x?") {
			re := regexp.MustCompile((`.*(?:twitter|x).com/.+status/([A-Za-z0-9]+)`))
			matches = re.FindStringSubmatch(url)
		} else if strings.HasPrefix(string(ctx.RequestURI()), "/instagram?") {
			re := regexp.MustCompile((`(?:reel(?:s?)|p)/([A-Za-z0-9_-]+)`))
			matches = re.FindStringSubmatch(url)
		} else if strings.HasPrefix(string(ctx.RequestURI()), "/tiktok?") {
			re := regexp.MustCompile((`/(?:video|v)/(\d+)`))
			matches = re.FindStringSubmatch(utils.GetRedirectURL(url))
		} else {
			errorMessage := "URL Invalid"
			ctx.Error(errorMessage, fasthttp.StatusMethodNotAllowed)
			return
		}

		if len(matches) > 1 {
			cachedResponse, err := cache.GetRedisClient().Get(context.Background(), matches[1]).Bytes()

			if err == nil {
				ctx.Response.Header.Add("Content-Type", "application/json")
				ctx.Write(cachedResponse)
			} else {
				next(ctx)
			}
		} else {
			errorMessage := "No URL"
			ctx.Error(errorMessage, fasthttp.StatusMethodNotAllowed)
			return
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
			cacheMiddleware(twitter.TwitterIndexer)(ctx)
		case "/x":
			cacheMiddleware(twitter.TwitterIndexer)(ctx)
		case "/tiktok":
			cacheMiddleware(tiktok.TikTokIndexer)(ctx)
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
