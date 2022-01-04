package webserver

import (
	"io"
	"net/http"
	"time"
)

type response struct {
	Status int
	Header http.Header
	Body []byte
}

type requestCacheEntry struct {
	Time time.Time
	ETag string
	Response response
}

func (srv *server) request(auth, method, url string) (response, error) {
	srv.requestMutex.Lock()
	defer srv.requestMutex.Unlock()

	now := time.Now()
	key := auth + ":" + method + ":" + url
	etag := ""

	// Check if the request is cached
	if entry, ok := srv.requestCache[key]; ok {
		if now.Sub(entry.Time) > 24 * time.Hour { // TODO: check this works
			etag = entry.ETag
		} else {
			r := entry.Response
			r.Header = r.Header.Clone()
			return r, nil
		}
	}

	// Otherwise, create a new request
	req, err := http.NewRequest(method, url, nil)
	if err != nil { return response{}, err }
	// Add an auth token if provided
	if auth != "" { req.Header.Add("Authorization", "token " + auth) }
	// Add the ETag if the cached response is due a check
	if etag != "" { req.Header.Add("If-None-Match", etag) }
	// Send the request
	var client http.Client
	resp, err := client.Do(req)
	if err != nil { return response{}, err }

	// If cached request is not modified, use the cached response
	var r response
	if resp.StatusCode == http.StatusNotModified {
		r = srv.requestCache[key].Response
		r.Header = r.Header.Clone()
	} else {
		body, err := io.ReadAll(resp.Body)
		if err != nil { return response{}, err }
		r = response { resp.StatusCode, resp.Header, body }
		etag = resp.Header.Get("etag")
	}

	// Cache the request and return a copy
	srv.requestCache[key] = requestCacheEntry { now, etag, r }
	r = srv.requestCache[key].Response
	r.Header = r.Header.Clone()
	return r, nil
}

func (srv *server) requestOK(w http.ResponseWriter, auth, method, url string) response {
	resp, err := srv.request(auth, method, url)
	if err != nil { srv.errorResponse(w, http.StatusInternalServerError) }
	return resp
}
