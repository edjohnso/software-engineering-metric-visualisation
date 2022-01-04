package webserver

import (
	"testing"
	"net/http/httptest"
	"net/http"
	"fmt"
)

func TestRequestHandling(t *testing.T) {
	srv, err := setupTestServer()
	if err != nil { t.Fatalf("Failed to setup test server: %v", err) }

	testAPIServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if r.Method != http.MethodGet || path != "/" {
			w.WriteHeader(http.StatusNotFound)
		}
		fmt.Fprintf(w, "%s %s %s", r.Header.Get("Authorization"), r.Method, path)
	}))
	defer testAPIServer.Close()

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
			resp, err := srv.request(testCase.token, testCase.method, testAPIServer.URL + testCase.url);
			if testCase.errorExpected {
				if err == nil {
					t.Errorf("An error was expected but none was returned")
				}
			} else {
				if err != nil {
					t.Errorf("An unexpected error occurred: %v", err)
				} else {
					assertWebserverResponse(t, resp, testCase.status, testCase.body)
				}
			}
		})
	}
}

func TestRequestCaching(t *testing.T) {
	srv, err := setupTestServer()
	if err != nil { t.Fatalf("Failed to setup test server: %v", err) }

	testAPIServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if etag := r.Header.Get("If-None-Match"); etag == "" {
			w.Header().Set("etag", "xyz")
			w.Write([]byte("abc"))
		} else {
			w.WriteHeader(http.StatusNotModified)
		}
	}))
	defer testAPIServer.Close()

	testCases := []struct { name string; token string; method string; url string; status int; body string } {
		{ "Request 1",          "",    http.MethodGet,  "/",    http.StatusOK, "abc" },
		{ "Request 2",          "",    http.MethodGet,  "/",    http.StatusOK, "abc" },
		{ "Different Token 1",  "123", http.MethodGet,  "/",    http.StatusOK, "abc" },
		{ "Different Token 2",  "123", http.MethodGet,  "/",    http.StatusOK, "abc" },
		{ "Different Method 1", "",    http.MethodPost, "/",    http.StatusOK, "abc" },
		{ "Different Method 2", "",    http.MethodPost, "/",    http.StatusOK, "abc" },
		{ "Different URL 1",    "",    http.MethodGet,  "/foo", http.StatusOK, "abc" },
		{ "Different URL 2",    "",    http.MethodGet,  "/foo", http.StatusOK, "abc" },
		{ "Original request",   "",    http.MethodGet,  "/",    http.StatusOK, "abc" },
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			resp, err := srv.request(testCase.token, testCase.method, testAPIServer.URL + testCase.url);
			if err != nil {
				t.Errorf("An unexpected error occurred: %v", err)
			} else {
				assertWebserverResponse(t, resp, testCase.status, testCase.body)
			}
		})
	}
}

func TestRequestOK(t *testing.T) {
	srv, err := setupTestServer()
	if err != nil { t.Fatalf("Failed to setup test server: %v", err) }

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
			testAPIServer := httptest.NewServer(http.HandlerFunc(testCase.handler))
			defer testAPIServer.Close()
			rr := httptest.NewRecorder()
			resp := srv.requestOK(rr, "", http.MethodGet, testAPIServer.URL + testCase.url)
			if testCase.status == 0 {
				assertResponseRecorder(t, rr, http.StatusInternalServerError, errorHTML)
			} else {
				assertWebserverResponse(t, resp, testCase.status, testCase.body)
			}
		})
	}
}
