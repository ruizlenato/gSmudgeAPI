package instagram

import (
	"context"
	"encoding/json"
	"fmt"
	"gSmudgeAPI/cache"
	"gSmudgeAPI/handler"
	"gSmudgeAPI/utils"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/tidwall/gjson"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpproxy"
)

func graphql(postID string, indexedMedia *handler.IndexedMedia) string {
	response := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(response)

	request := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(request)

	//Set header
	headers := map[string]string{
		`User-Agent`:         `Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:120.0) Gecko/20100101 Firefox/120.0`,
		`Accept`:             `*/*`,
		`Accept-Language`:    `pt-BR,pt;q=0.8,en-US;q=0.5,en;q=0.3`,
		`Content-Type`:       `application/x-www-form-urlencoded`,
		`X-FB-Friendly-Name`: `PolarisPostActionLoadPostQueryQuery`,
		`X-CSRFToken`:        `-m5n6c-w1Z9RmrGqkoGTMq`,
		`X-IG-App-ID`:        `936619743392459`,
		`X-FB-LSD`:           `AVp2LurCmJw`,
		`X-ASBD-ID`:          `129477`,
		`DNT`:                `1`,
		`Sec-Fetch-Dest`:     `empty`,
		`Sec-Fetch-Mode`:     `cors`,
		`Sec-Fetch-Site`:     `same-origin`,
	}
	request.Header.SetMethod(fasthttp.MethodPost)
	request.SetRequestURI("https://www.instagram.com/api/graphql")
	for key, value := range headers {
		request.Header.Set(key, value)
	}

	params := []string{
		`av=0`,
		`__d=www`,
		`__user=0`,
		`__a=1`,
		`__req=3`,
		`__hs=19734.HYP:instagram_web_pkg.2.1..0.0`,
		`dpr=1`,
		`__ccg=UNKNOWN`,
		`__rev=1010782723`,
		`__s=qg5qgx:efei15:ng6310`,
		`__hsi=7323030086241513400`,
		`__dyn=7xeUjG1mxu1syUbFp60DU98nwgU29zEdEc8co2qwJw5ux609vCwjE1xoswIwuo2awlU-cw5Mx62G3i1ywOwv89k2C1Fwc60AEC7U2czXwae4UaEW2G1NwwwNwKwHw8Xxm16wUxO1px-0iS2S3qazo7u1xwIwbS1LwTwKG1pg661pwr86C1mwrd6goK68jxe6V8`,
		`__csr=gps8cIy8WTDAqjWDrpda9SoLHhaVeVEgvhaJzVQ8hF-qEPBV8O4EhGmciDBQh1mVuF9V9d2FHGicAVu8GAmfZiHzk9IxlhV94aKC5oOq6Uhx-Ku4Kaw04Jrx64-0oCdw0MXw1lm0EE2Ixcjg2Fg1JEko0N8U421tw62wq8989EMw1QpV60CE02BIw`,
		`__comet_req=7`,
		`lsd=AVp2LurCmJw`,
		`jazoest=2989`,
		`__spin_r=1010782723`,
		`__spin_b=trunk`,
		`__spin_t=1705025808`,
		`fb_api_caller_class=RelayModern`,
		`fb_api_req_friendly_name=PolarisPostActionLoadPostQueryQuery`,
		fmt.Sprintf(`variables={"shortcode": "%v","fetch_comment_count":40,"fetch_related_profile_media_count":3,"parent_comment_count":24,"child_comment_count":3,"fetch_like_count":10,"fetch_tagged_user_count":null,"fetch_preview_comment_count":2,"has_threaded_comments":true,"hoisted_comment_id":null,"hoisted_reply_id":null}`, postID),
		`server_timestamps=true`,
		`doc_id=10015901848480474`,
	}

	reqBody := strings.Join(params, "&")
	request.SetBodyString(reqBody)

	client := &fasthttp.Client{
		ReadBufferSize:  16 * 1024,
		MaxConnsPerHost: 1024,
		Dial:            fasthttpproxy.FasthttpSocksDialer(os.Getenv("SOCKS_PROXY")),
	}

	err := client.Do(request, response)
	fasthttp.ReleaseRequest(request)
	if err != nil {
		log.Fatal(err)
	}

	return string(response.Body())
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

	PostID := (regexp.MustCompile((`(?:reel(?:s?)|p)/([A-Za-z0-9_-]+)`))).FindStringSubmatch(url)[1]

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
	graphql(PostID, &handler.IndexedMedia{})

	if len(match) == 2 {
		rJson := utils.UnescapeJSON(match[1])
		caption = gjson.Get(rJson, "shortcode_media.edge_media_to_caption.edges.0.node.text").String()
		result := gjson.Get(rJson, "shortcode_media.edge_sidecar_to_children.edges")
		if !result.Exists() {
			display_resources := gjson.Get(rJson, "shortcode_media.display_resources.@reverse.0")
			is_video := gjson.Get(rJson, "shortcode_media.is_video").Bool()
			for _, results := range display_resources.Array() {
				indexedMedia.Medias = append(indexedMedia.Medias, handler.Medias{
					Height: int(results.Get("config_height").Int()),
					Width:  int(results.Get("config_width").Int()),
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
					Height: int(results.Get("config_height").Int()),
					Width:  int(results.Get("config_width").Int()),
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

		height, width := utils.GetImageDimension(mainMediaURL)
		indexedMedia.Medias = append(indexedMedia.Medias, handler.Medias{
			Height: height,
			Width:  width,
			Source: mainMediaURL,
			Video:  false,
		})
	}

	if indexedMedia.Medias == nil {
		rjson := graphql(PostID, &handler.IndexedMedia{})
		result := gjson.Get(rjson, "data.xdt_shortcode_media")
		caption = result.Get("edge_media_to_caption.edges.0.node.text").String()
		if strings.Contains(result.Get("__typename").String(), "Video") {
			dimensions := result.Get("dimensions")
			for _, results := range dimensions.Array() {
				indexedMedia.Medias = append(indexedMedia.Medias, handler.Medias{
					Height: int(results.Get("height").Int()),
					Width:  int(results.Get("width").Int()),
					Source: strings.ReplaceAll(result.Get("video_url").String(), `\/`, `/`),
					Video:  true,
				})
			}
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
