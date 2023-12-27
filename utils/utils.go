package utils

import (
	"image"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"os"

	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpproxy"
)

type Header struct {
	key   string
	value string
}

type RequestParams struct {
	Query   map[string]string
	Headers map[string]string
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

	return g.Dx(), g.Dy()
}

func GetHTTPRes(Link string, params RequestParams) *fasthttp.Response {
	req := fasthttp.AcquireRequest()
	res := fasthttp.AcquireResponse()

	var client *fasthttp.Client

	if os.Getenv("SOCKS_PROXY") == "" {
		client = &fasthttp.Client{}
	} else {
		client = &fasthttp.Client{
			Dial: fasthttpproxy.FasthttpSocksDialer(os.Getenv("SOCKS_PROXY")),
		}
	}

	req.Header.SetMethod("GET")
	for key, value := range params.Headers {
		req.Header.Set(key, value)
	}

	req.SetRequestURI(Link)
	for key, value := range params.Query {
		req.URI().QueryArgs().Add(key, value)
	}

	client.Do(req, res)

	return res
}
