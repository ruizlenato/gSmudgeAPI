package utils

import (
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
)

type Header struct {
	key   string
	value string
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

func GetResBody(Link string, Query map[string]string, Headers map[string]string) []byte {
	req, err := http.NewRequest("GET", Link, nil)
	if err != nil {
		panic(err)
	}

	query := req.URL.Query()
	for key, value := range Query {
		query.Add(key, value)
	}

	req.URL.RawQuery = query.Encode()
	for key, value := range Headers {
		req.Header.Add(key, value)
	}

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
