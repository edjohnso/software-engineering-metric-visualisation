package webserver

import (
	"testing"
	"net/http/httptest"
	"net/http"
	"os"
	"github.com/gorilla/mux"
)

func TestUnauthHandler(t *testing.T) {
	srv, err := setupTestServer()
	if err != nil { t.Fatalf("Unable to setup HTTP test server: %v", err) }

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/", srv.unauthHandler)
	router.ServeHTTP(rr, request)

	assertResponseRecorder(t, rr, http.StatusOK, loginHTML)
}

func TestOAuthHandler(t *testing.T) {
	srv, err := setupTestServer()
	if err != nil { t.Fatalf("Unable to setup HTTP test server: %v", err) }

	// FIXME: I don't think it's possible to test a successful OAuth code exchange
	//        so this is just testing if it can handle exchange failure
	testCases := []struct { name string; code string; ok bool } {
		{ "Invalid code", "foo", false },
		//{ "Valid code", "TODO", true },
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/?code=" + testCase.code, nil)
			router := mux.NewRouter()
			router.HandleFunc("/", srv.oauthHandler).Queries("code", "{code}")
			router.ServeHTTP(rr, request)

			var expectedStatus int
			var expectedBody string
			if testCase.ok { expectedStatus = http.StatusOK } else { expectedStatus = http.StatusUnauthorized }
			if testCase.ok { expectedBody = graphHTML } else { expectedBody = loginHTML }
			assertResponseRecorder(t, rr, expectedStatus, expectedBody)
		})
	}
}

func TestHandleWSRequest(t *testing.T) {
	srv, err := setupTestServer()
	if err != nil { t.Fatalf("Unable to setup HTTP test server: %v", err) }

	pat := os.Getenv("GHO_PAT")
	if pat == "" {
		t.Fatalf("GHO_PAT environment variable not set")
	}

	testCases := []struct { name string; token string; addCookie bool; status int; body string } {
		{ "No access token cookie", "", false, http.StatusUnauthorized, errorHTML },
		{ "Invalid access token", "", true, http.StatusUnauthorized, errorHTML },
		{ "Invalid WebSocket connection", pat, true, http.StatusBadRequest, "Bad Request" },
		// TODO: testing Websocket connection
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/", nil)
			if testCase.addCookie {
				request.AddCookie(&http.Cookie { Name: "gho", Value: testCase.token })
			}
			router := mux.NewRouter()
			router.HandleFunc("/", srv.wsHandler)
			router.ServeHTTP(rr, request)
			assertResponseRecorder(t, rr, testCase.status, testCase.body)
		})
	}
}

func TestHandleUserRequest(t *testing.T) {
	srv, err := setupTestServer()
	if err != nil { t.Fatalf("Unable to setup HTTP test server: %v", err) }

	pat := os.Getenv("GHO_PAT")
	if pat == "" {
		t.Fatalf("GHO_PAT environment variable not set")
	}

	testCases := []struct { name string; token string; addCookie bool; ok bool } {
		{ "No access token cookie", "", false, false },
		{ "Invalid access token", "", true, false },
		{ "Valid access token (PAT)", pat, true, true },
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/", nil)
			if testCase.addCookie {
				request.AddCookie(&http.Cookie { Name: "gho", Value: testCase.token })
			}
			router := mux.NewRouter()
			router.HandleFunc("/", srv.userHandler)
			router.ServeHTTP(rr, request)

			var expectedStatus int
			var expectedBody string
			if testCase.ok { expectedStatus = http.StatusOK } else { expectedStatus = http.StatusUnauthorized }
			if testCase.ok { expectedBody = graphHTML } else { expectedBody = loginHTML }
			assertResponseRecorder(t, rr, expectedStatus, expectedBody)
		})
	}
}

func TestErrorResponse(t *testing.T) {
	srv, err := setupTestServer()
	if err != nil { t.Fatalf("Unable to setup HTTP test server: %v", err) }

	testCases := []struct { name string; status int } {
		{ "200 OK", http.StatusOK },
		{ "404 Not Found", http.StatusNotFound },
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			srv.errorResponse(rr, testCase.status)
			assertResponseRecorder(t, rr, testCase.status, errorHTML)
		})
	}
}

func TestExecuteTemplate(t *testing.T) {
	srv, err := setupTestServer()
	if err != nil { t.Fatalf("Unable to setup HTTP test server: %v", err) }

	rr := httptest.NewRecorder()
	srv.executeTemplate(rr, "this template doesn't exist", nil)
	assertResponseRecorder(t, rr, http.StatusInternalServerError, errorHTML)

	rr = httptest.NewRecorder()
	srv.executeTemplate(rr, "login.html", nil)
	assertResponseRecorder(t, rr, http.StatusOK, loginHTML)
}
