package utils

import (
	"image"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"net/http"
	"regexp"

	"github.com/valyala/fasthttp"
)

type RequestParams struct {
	Query   map[string]string
	Headers map[string]string
}

func UnescapeJSON(str string) string {
	re := regexp.MustCompile(`\\(u[0-9a-fA-F]{4}|"|\\)`)
	return re.ReplaceAllStringFunc(str, func(s string) string {
		switch s {
		case `\"`:
			return `"`
		case `\\`:
			return `\`
		case `\u0022`:
			return `"`
		default:
			return s
		}
	})
}

func GetImageDimension(url string) (int, int) {
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

	return g.Dy(), g.Dx()
}

func GetRedirectURL(url string) string {
	res, _ := http.Get(url)
	return res.Request.URL.String()

}

func GetHTTPRes(Link string, params RequestParams) *fasthttp.Response {
	req := fasthttp.AcquireRequest()
	res := fasthttp.AcquireResponse()

	client := &fasthttp.Client{ReadBufferSize: 8192}

	req.Header.SetMethod("GET")
	for key, value := range params.Headers {
		req.Header.Set(key, value)
	}

	req.SetRequestURI(Link)
	for key, value := range params.Query {
		req.URI().QueryArgs().Add(key, value)
	}

	err := client.Do(req, res)
	if err != nil {
		log.Fatal(err)
	}

	return res
}
