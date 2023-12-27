package instagram

import (
	"context"
	"encoding/json"
	"fmt"
	"gSmudgeAPI/cache"
	"gSmudgeAPI/handler"
	"gSmudgeAPI/utils"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/gocolly/colly"
	"github.com/tidwall/gjson"
	"github.com/valyala/fasthttp"
)

func InstagramIndexer(ctx *fasthttp.RequestCtx) {
	url := string(ctx.QueryArgs().Peek("url"))
	if len(url) == 0 {
		errorMessage := "No URL specified"
		ctx.Error(errorMessage, fasthttp.StatusMethodNotAllowed)
		return
	}

	PostID := (regexp.MustCompile((`(?:reel|p)/([A-Za-z0-9_-]+)`))).FindStringSubmatch(url)[1]

	c := colly.NewCollector()

	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:120.0) Gecko/20100101 Firefox/120.0")
		r.Headers.Set("accept-language", "en-US,en;q=0.9")
	})

	indexedMedia := &handler.IndexedMedia{}
	var caption string
	c.OnHTML("div[data-media-type=GraphImage] img.EmbeddedMediaImage", func(h *colly.HTMLElement) {
		width, height := utils.GetImageDimension(h.Attr("src"))
		indexedMedia.Medias = append(indexedMedia.Medias, handler.Medias{
			Width:  int(width),
			Height: int(height),
			Source: h.Attr("src"),
		})
	})

	c.OnHTML("div[class=Caption]", func(h *colly.HTMLElement) {
		r := regexp.MustCompile(`.*</a><br/><br/>(.*)<div class="CaptionComments">.*`)
		match := r.FindStringSubmatch(fmt.Sprint(h.DOM.Html()))
		caption = match[1]
	})

	c.OnHTML("script", func(h *colly.HTMLElement) {
		r := regexp.MustCompile(`\\\"gql_data\\\":([\s\S]*)\}\"\}\]\]\,\[\"NavigationMetrics`)
		match := r.FindStringSubmatch(h.Text)

		if len(match) < 2 {
			return
		}

		s := strings.ReplaceAll(match[1], `\"`, `"`)
		s = strings.ReplaceAll(s, `\n`, ``)
		s = strings.ReplaceAll(s, `\\/`, `/`)
		s = strings.ReplaceAll(s, `\\`, `\`)

		result := gjson.Get(s, "shortcode_media.edge_sidecar_to_children.edges")
		if !result.Exists() {
			display_resources := gjson.Get(s, "shortcode_media.display_resources.@reverse.0")
			is_video := gjson.Get(s, "shortcode_media.is_video").Bool()
			for _, results := range display_resources.Array() {
				indexedMedia.Medias = append(indexedMedia.Medias, handler.Medias{
					Width:  int(results.Get("config_width").Int()),
					Height: int(results.Get("config_height").Int()),
					Source: results.Get("src").String(),
					Video:  is_video,
				})
			}
		}
		for _, results := range result.Array() {
			is_video := results.Get("node.is_video").Bool()
			display_resources := results.Get("node.display_resources.@reverse.0")
			for _, results := range display_resources.Array() {
				indexedMedia.Medias = append(indexedMedia.Medias, handler.Medias{
					Width:  int(results.Get("config_width").Int()),
					Height: int(results.Get("config_height").Int()),
					Source: results.Get("src").String(),
					Video:  is_video,
				})
			}
		}
	})
	c.Visit(fmt.Sprintf("https://www.instagram.com/p/%v/embed/captioned/", PostID))

	if indexedMedia.Medias == nil {
		Headers := map[string]string{
			"Sec-Fetch-Mode": "navigate",
			"User-Agent":     "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Safari/537.36",
			"Referer":        fmt.Sprintf("https://www.instagram.com/p/%v/", PostID),
		}
		Query := map[string]string{"query_hash": "9f8827793ef34641b2fb195d4d41151c", "variables": fmt.Sprintf(`{"shortcode":"%v"}`, PostID)}

		json := utils.GetHTTPRes("https://www.instagram.com/graphql/query/", utils.RequestParams{Query: Query, Headers: Headers}).Body()
		caption = gjson.GetBytes(json, "data.shortcode_media.edge_media_to_caption.edges.0.node.text").String()
		display_resources := gjson.GetBytes(json, "data.shortcode_media.display_resources.@reverse.0")
		is_video := gjson.GetBytes(json, "data.shortcode_media.is_video").Bool()
		for _, results := range display_resources.Array() {
			indexedMedia.Medias = append(indexedMedia.Medias, handler.Medias{
				Width:  int(results.Get("config_width").Int()),
				Height: int(results.Get("config_height").Int()),
				Source: results.Get("src").String(),
				Video:  is_video,
			})
		}

	}

	ixt := handler.IndexedMedia{
		URL:     url,
		Medias:  indexedMedia.Medias,
		Caption: caption}

	jsonResponse, _ := json.Marshal(ixt)

	err := cache.GetRedisClient().Set(context.Background(), PostID, jsonResponse, 24*time.Hour*60).Err()
	if err != nil {
		log.Println("Error setting cache:", err)
	}
	ctx.Response.Header.Add("Content-Type", "application/json")
	json.NewEncoder(ctx).Encode(ixt)
}
