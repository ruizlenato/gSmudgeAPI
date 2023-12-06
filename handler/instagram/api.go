package instagram

import (
	"encoding/json"
	"fmt"
	"gSmudgeAPI/handler"
	"net/http"
	"regexp"
	"strings"

	"github.com/gocolly/colly"
	"github.com/tidwall/gjson"
)

func InstagramIndexer(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query().Get("url")
	if len(url) == 0 {
		response := "No URL specified"
		http.Error(w, response, http.StatusMethodNotAllowed)
		return
	}

	PostID := (regexp.MustCompile((`(?:reel|p)/([A-Za-z0-9_-]+)`))).FindStringSubmatch(url)[1]

	c := colly.NewCollector()

	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2228.0 Safari/537.36")
		r.Headers.Set("accept-language", "en-US,en;q=0.9")
	})

	indexedMedia := &handler.IndexedMedia{}
	var caption string
	c.OnHTML("div[data-media-type=GraphImage] img.EmbeddedMediaImage", func(h *colly.HTMLElement) {
		indexedMedia.Medias = append(indexedMedia.Medias, handler.Medias{Source: h.Attr("src")})
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
			for _, results := range display_resources.Array() {
				indexedMedia.Medias = append(indexedMedia.Medias, handler.Medias{
					Width:  int(results.Get("config_width").Int()),
					Height: int(results.Get("config_height").Int()),
					Source: results.Get("src").String(),
				})
			}
		}
		for _, results := range result.Array() {
			display_resources := results.Get("node.display_resources.@reverse.0")
			for _, results := range display_resources.Array() {
				indexedMedia.Medias = append(indexedMedia.Medias, handler.Medias{
					Width:  int(results.Get("config_width").Int()),
					Height: int(results.Get("config_height").Int()),
					Source: results.Get("src").String(),
				})
			}
		}
	})

	c.Visit(fmt.Sprintf("https://www.instagram.com/p/%v/embed/captioned/", PostID))

	ixt := handler.IndexedMedia{
		URL:     url,
		Medias:  indexedMedia.Medias,
		Caption: caption}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ixt)
}
