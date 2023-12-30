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

	"github.com/tidwall/gjson"
	"github.com/valyala/fasthttp"
)

func graphql(postID string, indexedMedia *handler.IndexedMedia) (string, *handler.IndexedMedia) {
	Headers := map[string]string{
		"Sec-Fetch-Mode": "navigate",
		"User-Agent":     "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Safari/537.36",
		"Referer":        fmt.Sprintf("https://www.instagram.com/p/%v/", postID),
	}
	Query := map[string]string{"query_hash": "9f8827793ef34641b2fb195d4d41151c", "variables": fmt.Sprintf(`{"shortcode":"%v"}`, postID)}

	res := utils.GetHTTPRes("https://www.instagram.com/graphql/query/", utils.RequestParams{Query: Query, Headers: Headers, Proxy: true}).Body()
	caption := gjson.GetBytes(res, "data.shortcode_media.edge_media_to_caption.edges.0.node.text").String()
	if gjson.GetBytes(res, "data.shortcode_media.__typename").String() == "GraphSidecar" {
		display_resources := gjson.GetBytes(res, "data.shortcode_media.edge_sidecar_to_children.edges")
		for _, edge := range display_resources.Array() {
			is_video := edge.Get("node.is_video").Bool()
			for _, results := range edge.Get("node.display_resources.@reverse.0").Array() {
				indexedMedia.Medias = append(indexedMedia.Medias, handler.Medias{
					Width:  int(results.Get("config_width").Int()),
					Height: int(results.Get("config_height").Int()),
					Source: results.Get("src").String(),
					Video:  is_video,
				})
			}
		}
	} else {
		is_video := gjson.GetBytes(res, "data.shortcode_media.is_video").Bool()
		for _, results := range gjson.GetBytes(res, "data.shortcode_media.display_resources.@reverse.0").Array() {
			indexedMedia.Medias = append(indexedMedia.Medias, handler.Medias{
				Width:  int(results.Get("config_width").Int()),
				Height: int(results.Get("config_height").Int()),
				Source: results.Get("src").String(),
				Video:  is_video,
			})
		}
	}
	return caption, indexedMedia

}

func InstagramIndexer(ctx *fasthttp.RequestCtx) {
	url := string(ctx.QueryArgs().Peek("url"))
	if len(url) == 0 {
		errorMessage := "No URL specified"
		ctx.Error(errorMessage, fasthttp.StatusMethodNotAllowed)
		return
	}
	indexedMedia := &handler.IndexedMedia{}
	var caption string

	PostID := (regexp.MustCompile((`(?:reel|p)/([A-Za-z0-9_-]+)`))).FindStringSubmatch(url)[1]

	Headers := map[string]string{
		"accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9",
		"accept-language":           "en-US,en;q=0.9",
		"cache-control":             "max-age=0",
		"connection":                "close",
		"sec-fetch-mode":            "navigate",
		"upgrade-insecure-requests": "1",
		"referer":                   "https://www.instagram.com/",
		"User-Agent":                "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Safari/537.36",
		"viewport-width":            "1280",
	}

	res := utils.GetHTTPRes(fmt.Sprintf("https://www.instagram.com/p/%v/embed/captioned/", PostID), utils.RequestParams{Headers: Headers}).Body()
	r := regexp.MustCompile(`\\\"gql_data\\\":([\s\S]*)\}\"\}`)
	match := r.FindStringSubmatch(string(res))
	if len(match) == 2 {
		rJson := utils.UnescapeJSON(match[1])
		caption = gjson.Get(rJson, "shortcode_media.edge_media_to_caption.edges.0.node.text").String()
		result := gjson.Get(rJson, "shortcode_media.edge_sidecar_to_children.edges")
		if !result.Exists() {
			display_resources := gjson.Get(rJson, "shortcode_media.display_resources.@reverse.0")
			is_video := gjson.Get(rJson, "shortcode_media.is_video").Bool()
			for _, results := range display_resources.Array() {
				indexedMedia.Medias = append(indexedMedia.Medias, handler.Medias{
					Width:  int(results.Get("config_width").Int()),
					Height: int(results.Get("config_height").Int()),
					Source: strings.ReplaceAll(results.Get("src").String(), `\/`, `/`),
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
					Source: strings.ReplaceAll(results.Get("src").String(), `\/`, `/`),
					Video:  is_video,
				})
			}
		}

	} else {
		// Media
		re := regexp.MustCompile(`class="Content(.*?)src="(.*?)"`)
		mainMediaData := re.FindAllStringSubmatch(string(res), -1)
		mainMediaURL := (strings.ReplaceAll(mainMediaData[0][2], "amp;", ""))

		// Caption
		re = regexp.MustCompile(`(?s)class="Caption"(.*?)class="CaptionUsername"(.*?)<\/a>(.*?)<div`)
		captionData := re.FindAllStringSubmatch(string(res), -1)
		if len(captionData) > 0 && len(captionData[0]) > 2 {
			re = regexp.MustCompile(`<[^>]*>`)
			caption = strings.TrimSpace(re.ReplaceAllString(captionData[0][3], ""))
		}

		width, height := utils.GetImageDimension(mainMediaURL)
		indexedMedia.Medias = append(indexedMedia.Medias, handler.Medias{
			Width:  width,
			Height: height,
			Source: mainMediaURL,
			Video:  false,
		})
	}

	if indexedMedia.Medias == nil {
		caption, indexedMedia = graphql(PostID, &handler.IndexedMedia{})
		for i := 0; caption == "" && i < 15; i++ {
			caption, indexedMedia = graphql(PostID, &handler.IndexedMedia{})
		}
	}

	ixt := handler.IndexedMedia{
		URL:     url,
		Medias:  indexedMedia.Medias,
		Caption: caption}

	jsonResponse, _ := json.Marshal(ixt)

	if indexedMedia.Medias != nil {
		err := cache.GetRedisClient().Set(context.Background(), PostID, jsonResponse, 24*time.Hour*60).Err()
		if err != nil {
			log.Println("Error setting cache:", err)
		}
	}
	ctx.Response.Header.Add("Content-Type", "application/json")
	json.NewEncoder(ctx).Encode(ixt)
}
