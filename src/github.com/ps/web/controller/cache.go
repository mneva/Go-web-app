package controller

import (
	"bytes"
	"flag"
	"io"
	"net/http"
	"strconv"
)

var cachServiceURL = flag.String("cacheservice", "https://172.18.0.13:5000", "Address of the caching service provider")

func getFromCache(key string) (io.ReadCloser, bool) {
	resp, err := http.Get(*cachServiceURL + "/?key=" + key)
	if err != nil || resp.StatusCode != http.StatusOK {
		println("get fail")
		return nil, false
	}
	return resp.Body, true
}

func saveToCache(key string, duration int64, data []byte) {
	req, _ := http.NewRequest(http.MethodPost, *cachServiceURL+"/?key="+key,
		bytes.NewBuffer(data))
	req.Header.Add("cache-control", "maxage="+strconv.FormatInt(duration, 10))
	http.DefaultClient.Do(req)
}

func invalidateCacheEntry(key string) {
	http.Get(*cachServiceURL + "/invalidate?key=" + key)
}
