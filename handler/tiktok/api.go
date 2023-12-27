package tiktok

import (
	"context"
	"encoding/json"
	"gSmudgeAPI/cache"
	"gSmudgeAPI/handler"
	"gSmudgeAPI/utils"
	"log"
	"regexp"
	"slices"
	"time"

	"github.com/tidwall/gjson"
	"github.com/valyala/fasthttp"
)

func TikTokIndexer(ctx *fasthttp.RequestCtx) {
	url := string(ctx.QueryArgs().Peek("url"))
	if len(url) == 0 {
		errorMessage := "No URL specified"
		ctx.Error(errorMessage, fasthttp.StatusMethodNotAllowed)
		return
	}

	redirectURL := utils.GetRedirectURL(url)
	VideoID := (regexp.MustCompile((`/(?:video|v)/(\d+)`))).FindStringSubmatch(redirectURL)[1]
	Headers := map[string]string{"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:120.0) Gecko/20100101 Firefox/120.0"}
	Query := map[string]string{
		"aweme_id": string(VideoID),
		"aid":      "1128",
	}
	body := utils.GetHTTPRes("https://api16-normal-c-useast1a.tiktokv.com/aweme/v1/feed/", utils.RequestParams{Query: Query, Headers: Headers}).Body()
	caption := gjson.GetBytes(body, "aweme_list.0.desc").String()
	indexedMedia := &handler.IndexedMedia{}

	if slices.Contains([]int{2, 68, 150}, int(gjson.GetBytes(body, "aweme_list.0.aweme_type").Int())) {
		for _, p := range gjson.GetBytes(body, "aweme_list.0.image_post_info.images").Array() {
			indexedMedia.Medias = append(indexedMedia.Medias, handler.Medias{
				Width:  int(p.Get("display_image.width").Int()),
				Height: int(p.Get("display_image.height").Int()),
				Source: p.Get("display_image.url_list.0").String(),
				Video:  false,
			})
		}
	} else {
		for _, v := range gjson.GetBytes(body, "aweme_list.0.video.play_addr").Array() {
			indexedMedia.Medias = append(indexedMedia.Medias, handler.Medias{
				Width:  int(v.Get("width").Int()),
				Height: int(v.Get("height").Int()),
				Source: v.Get("url_list.0").String(),
				Video:  true,
			})
		}
	}

	ixt := handler.IndexedMedia{
		URL:     redirectURL,
		Medias:  indexedMedia.Medias,
		Caption: caption}

	jsonResponse, _ := json.Marshal(ixt)
	err := cache.GetRedisClient().Set(context.Background(), VideoID, jsonResponse, 24*time.Hour*60).Err()
	if err != nil {
		log.Println("Error setting cache:", err)
	}
	ctx.Response.Header.Add("Content-Type", "application/json")
	json.NewEncoder(ctx).Encode(ixt)
}
