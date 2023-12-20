package instagram

import (
	"context"
	"encoding/json"
	"fmt"
	"gSmudgeAPI/cache"
	"gSmudgeAPI/handler"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gocolly/colly"
	"github.com/tidwall/gjson"
)

func getImageDimension(url string) (int, int) {
	res, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	m, _, err := image.Decode(res.Body)
	if err != nil {
		panic(err)
	}
	g := m.Bounds()

	return g.Dx(), g.Dy()
}

func rgraphql(PostID string) []byte {
	req, err := http.NewRequest("GET", `https://www.instagram.com/graphql/query/?query_hash=b3055c01b4b222b8a47dc12b090e4e64`, nil)
	if err != nil {
		panic(err)
	}
	q := req.URL.Query()
	q.Add("variables", fmt.Sprintf(`{"shortcode":"%v"}`, PostID))
	req.URL.RawQuery = q.Encode()

	req.Header.Add("Accept-Language", "en-US,en;q=0.9")
	req.Header.Add("Connection", "close")
	req.Header.Add("Sec-Fetch-Mode", "navigate")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:120.0) Gecko/20100101 Firefox/120.0")
	req.Header.Add("Referer", fmt.Sprintf("https://www.instagram.com/p/%v/", PostID))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	return body
}

func InstagramIndexer(w http.ResponseWriter, r *http.Request) {
	print(r.URL.String())
	url := r.URL.Query().Get("url")
	if len(url) == 0 {
		response := "No URL specified"
		http.Error(w, response, http.StatusMethodNotAllowed)
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
		width, height := getImageDimension(h.Attr("src"))
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
		json := rgraphql(PostID)
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ixt)
}
