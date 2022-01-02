package webserver

import (
	"testing"
	"net/http/httptest"
	"net/http"
	"io"
	"fmt"
	"strings"
)

func TestRequestHandling(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if r.Method != http.MethodGet || path != "/" {
			w.WriteHeader(http.StatusNotFound)
		}
		fmt.Fprintf(w, "%s %s %s", r.Header.Get("Authorization"), r.Method, path)
	}))
	defer srv.Close()

	testCases := []struct { name string; token string; method string; url string; status int; body string; errorExpected bool } {
		{ "Invalid request",   "",    "xxx",           "xxx",  0,                   "",           true },
		{ "Failing request",   "",    http.MethodGet,  "000",  0,                   "",           true },
		{ "GET /",             "",    http.MethodGet,  "/",    http.StatusOK,       "GET /",      false },
		{ "GET /foo/bar",      "",    http.MethodGet,  "/foo", http.StatusNotFound, "GET /foo",   false },
		{ "GET / with token",  "123", http.MethodGet,  "/",    http.StatusOK,       "token 123 GET /",  false },
		{ "POST / with token", "123", http.MethodPost, "/",    http.StatusNotFound, "token 123 POST /", false },
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			resp, err := request(testCase.token, testCase.method, srv.URL + testCase.url);
			if testCase.errorExpected {
				if err == nil {
					t.Errorf("An error was expected but none was returned")
				}
			} else {
				if err != nil {
					t.Errorf("An unexpected error occurred: %v", err)
				} else if resp == nil {
					t.Errorf("No response was returned")
				} else {
					assertResponse(t, resp, testCase.status, testCase.body)
				}
			}
		})
	}
}

func TestRequestCaching(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if etag := r.Header.Get("If-None-Match"); etag == "" {
			w.Header().Set("etag", "xyz")
			w.Write([]byte("abc"))
		} else {
			w.WriteHeader(http.StatusNotModified)
		}
	}))
	defer srv.Close()

	testCases := []struct { name string; token string; method string; url string; status int; body string } {
		{ "Request 1",          "",    http.MethodGet,  "/",    http.StatusOK,          "abc" },
		{ "Request 2",          "",    http.MethodGet,  "/",    http.StatusNotModified, "abc" },
		{ "Different Token 1",  "123", http.MethodGet,  "/",    http.StatusOK,          "abc" },
		{ "Different Token 2",  "123", http.MethodGet,  "/",    http.StatusNotModified, "abc" },
		{ "Different Method 1", "",    http.MethodPost, "/",    http.StatusOK,          "abc" },
		{ "Different Method 2", "",    http.MethodPost, "/",    http.StatusNotModified, "abc" },
		{ "Different URL 1",    "",    http.MethodGet,  "/foo", http.StatusOK,          "abc" },
		{ "Different URL 2",    "",    http.MethodGet,  "/foo", http.StatusNotModified, "abc" },
		{ "Original request",   "",    http.MethodGet,  "/",    http.StatusNotModified, "abc" },
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			resp, err := request(testCase.token, testCase.method, srv.URL + testCase.url);
			if err != nil {
				t.Errorf("An unexpected error occurred: %v", err)
			} else {
				assertResponse(t, resp, testCase.status, testCase.body)
			}
		})
	}
}

func TestRequestAndParse(t *testing.T) {
	testCases := []struct { name string; url string; status int; body string; handler func(w http.ResponseWriter, r *http.Request) } {
		{
			"Bad request", "xxx", 0, "",
			func(w http.ResponseWriter, r *http.Request) {},
		},
		{
			"Error response", "/", http.StatusNotFound, "error",
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte("error"))
			},
		},
		{
			"OK response", "/", http.StatusOK, "hello",
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("hello"))
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(testCase.handler))
			defer srv.Close()
			resp, body := requestAndParse("", http.MethodGet, srv.URL + testCase.url)
			if resp == nil {
				if testCase.status != 0 {
					t.Errorf("Request failed unexpectedly")
				}
			} else {
				if resp.StatusCode != testCase.status {
					t.Errorf("Expected status: %d - Actual status: %d", testCase.status, resp.StatusCode)
				}
				if strings.TrimSpace(body) != testCase.body {
					t.Errorf("Expected response: %s - Actual response: %s", testCase.body, body)
				}
			}
		})
	}
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
