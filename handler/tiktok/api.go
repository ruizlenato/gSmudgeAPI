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

	"github.com/valyala/fasthttp"
)

func TikTokIndexer(ctx *fasthttp.RequestCtx) {
	url := string(ctx.QueryArgs().Peek("url"))
	if len(url) == 0 {
		errorMessage := "No URL specified"
		ctx.Error(errorMessage, fasthttp.StatusMethodNotAllowed)
		return
	}

	VideoID := (regexp.MustCompile((`/(?:video|photo|v)/(\d+)`))).FindStringSubmatch(url)[1]
	Headers := map[string]string{"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:120.0) Gecko/20100101 Firefox/120.0"}
	Query := map[string]string{
		"aweme_id": string(VideoID),
		"aid":      "1128",
	}
	body := utils.GetHTTPRes("https://api16-normal-c-useast1a.tiktokv.com/aweme/v1/feed/", utils.RequestParams{Query: Query, Headers: Headers}).Body()

	var tikTokData TikTokData
	err := json.Unmarshal(body, &tikTokData)
	if err != nil {
		log.Println(err)
	}

	indexedMedia := &handler.IndexedMedia{}

	if slices.Contains([]int{2, 68, 150}, tikTokData.AwemeList[0].AwemeType) {
		for _, media := range tikTokData.AwemeList[0].ImagePostInfo.Images {
			indexedMedia.Medias = append(indexedMedia.Medias, handler.Medias{
				Height: media.DisplayImage.Height,
				Width:  media.DisplayImage.Width,
				Source: media.DisplayImage.URLList[1],
				Video:  false,
			})
		}
	} else {
		indexedMedia.Medias = append(indexedMedia.Medias, handler.Medias{
			Height: tikTokData.AwemeList[0].Video.PlayAddr.Height,
			Width:  tikTokData.AwemeList[0].Video.PlayAddr.Width,
			Source: tikTokData.AwemeList[0].Video.PlayAddr.URLList[0],
			Video:  true,
		})
	}

	ixt := handler.IndexedMedia{
		URL:     url,
		Medias:  indexedMedia.Medias,
		Caption: tikTokData.AwemeList[0].Desc}

	jsonResponse, _ := json.Marshal(ixt)
	err = cache.GetRedisClient().Set(context.Background(), VideoID, jsonResponse, 24*time.Hour*60).Err()
	if err != nil {
		log.Println("Error setting cache:", err)
	}
	ctx.Response.Header.Add("Content-Type", "application/json")
	json.NewEncoder(ctx).Encode(ixt)
}
