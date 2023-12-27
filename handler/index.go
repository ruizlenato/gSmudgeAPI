package handler

import (
	"encoding/json"

	"github.com/valyala/fasthttp"
)

type Medias struct {
	Width  int    `json:"width,omitempty"`
	Height int    `json:"height,omitempty"`
	Source string `json:"src"`
	Video  bool   `json:"is_video"`
}

type IndexedMedia struct {
	URL     string   `json:"url"`
	Medias  []Medias `json:"medias"`
	Caption string   `json:"caption,omitempty"`
}

func HandlerIndex(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.Add("Content-Type", "application/json")
	json.NewEncoder(ctx).Encode(map[string]interface{}{
		"/instagram": map[string]interface{}{
			"method":       "GET",
			"description":  "Scrapper for instagram",
			"query_params": map[string]string{"url": "url"},
		},
		"/twitter": map[string]interface{}{
			"method":       "GET",
			"description":  "Scrapper for twitter/x",
			"query_params": map[string]string{"url": "url"},
		},
		"/tiktok": map[string]interface{}{
			"method":       "GET",
			"description":  "Scrapper for tiktok",
			"query_params": map[string]string{"url": "url"},
		},
	})
}
