package webserver

import (
	"io"
	"log"
	"bytes"
	"net/http"
)

type cacheEntry struct {
	ETag string
	Response []byte
}

var cache = map[string]cacheEntry{}

func request(token, method, url string) (*http.Response, error) {
	log.Printf("Sending request: %s %s", method, url)
	req, err := http.NewRequest(method, url, nil)
	if err != nil { return nil, err }
	if token != "" {
		req.Header.Add("Authorization", "token " + token)
	}

	cacheEntryKey := token + ":" + method + ":" + url
	if entry, ok := cache[cacheEntryKey]; ok {
		req.Header.Add("If-None-Match", entry.ETag)
	}

	var client http.Client
	resp, err := client.Do(req)
	if err != nil { return nil, err }

	if resp.StatusCode == http.StatusNotModified {
		resp.Body = io.NopCloser(bytes.NewBuffer(cache[cacheEntryKey].Response))
		log.Printf("Using cached response.")
	} else {
		if etag := resp.Header.Get("etag"); etag != "" {
			body, err := io.ReadAll(resp.Body)
			if err != nil { return nil, err }
			cache[cacheEntryKey] = cacheEntry { etag, body }
			resp.Body = io.NopCloser(bytes.NewBuffer(body))
			log.Printf("Response cached.")
		}
	}

	return resp, nil
}

func requestAndParse(token, method, url string) (*http.Response, string) {
	resp, err := request(token, method, url)
	if err != nil {
		log.Printf("Unable to make request: %v", err)
		return nil, ""
	}

	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		log.Printf("Unable to read response body: %v", err)
		return resp, ""
	}

	return resp, string(body)
}
