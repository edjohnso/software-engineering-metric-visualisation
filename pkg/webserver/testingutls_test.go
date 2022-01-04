package webserver

import (
	"testing"
	"html/template"
	"net/http/httptest"
	"net/http"
	"strings"
	"io"
	"sync"
)

const loginHTML = `<p>login page!</p>`
const graphHTML = `<p>graph page!</p>`
const errorHTML = `<p>error page!</p>`

func setupTestServer() (*server, error) {
	var srv server
	var err error
	if srv.clientID, srv.clientSecret, err = loadSecrets(); err != nil { return nil, err }
	if srv.templates, err = template.New("login.html").Parse(loginHTML); err != nil { return &srv, err }
	if srv.templates, err = srv.templates.New("graph.html").Parse(graphHTML); err != nil { return &srv, err }
	if srv.templates, err = srv.templates.New("error.html").Parse(errorHTML); err != nil { return &srv, err }
	srv.requestCache = map[string]requestCacheEntry{}
	srv.collabGraph = map[string]userEntry{}
	srv.requestMutex = &sync.Mutex{}
	return &srv, nil
}

func assertResponse(t *testing.T, r *http.Response, status int, body string) {
	if r.StatusCode != status {
		t.Errorf("Expected status: %d - Actual status: %d", status, r.StatusCode)
	}
	buf, err := io.ReadAll(r.Body)
	if err != nil {
		t.Fatalf("Unable to read response body")
	}
	if strings.TrimSpace(string(buf)) != body {
		t.Errorf("Expected response: %s - Actual response: %s", body, r.Body)
	}
}

func assertWebserverResponse(t *testing.T, resp response, status int, body string) {
	if resp.Status != status {
		t.Errorf("Expected status: %d - Actual status: %d", status, resp.Status)
	}
	respbody := string(resp.Body)
	if strings.TrimSpace(respbody) != body {
		t.Errorf("Expected response: %s - Actual response: %s", body, respbody)
	}
}

func assertResponseRecorder(t *testing.T, rr *httptest.ResponseRecorder, status int, body string) {
	if rr.Code != status {
		t.Errorf("Expected status: %d - Actual status: %d", status, rr.Code)
	}
	if strings.TrimSpace(rr.Body.String()) != body {
		t.Errorf("Expected response: %s - Actual response: %s", body, rr.Body)
	}
}
