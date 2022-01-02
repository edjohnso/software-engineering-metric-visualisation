package webserver

import (
	"testing"
	"net/http/httptest"
	"net/http"
	"html/template"
	"path/filepath"
	"strings"
	"os"
	"syscall"
	"time"
	"fmt"
	"github.com/gorilla/mux"
)

const loginHTML = `<p>login page!</p>`
const userHTML = `<p>user page!</p>`
const errorHTML = `<p>error page!</p>`

func TestStart(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string { "login.html": loginHTML, "user.html": userHTML, "error.html": errorHTML }
	for file, content := range files {
		if err := os.WriteFile(filepath.Join(dir, file), []byte(content), os.ModePerm); err != nil {
			t.Fatalf("Unable to create %s file: %v", file, err)
		}
	}

	ok := false
	go func() {
		time.Sleep(time.Millisecond * 100)
		syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		time.Sleep(time.Millisecond * 100)
		if !ok {
			fmt.Printf("--- FAIL: Server failed to shutdown after interrupt signal\n\n")
			os.Exit(1)
		}
	}()

	err := Start(":8080", filepath.Join(dir, "*.html"))
	ok = true

	if err != http.ErrServerClosed {
		t.Errorf("Server returned unexpected error: %v", err)
	}
}

func TestLoadSecrets(t *testing.T) {
	testCases := []struct { name string; clientID string; clientSecret string; errorExpected bool } {
		{ "None", "", "", true },
		{ "Client secret only", "", "123", true },
		{ "Client ID only", "123", "", true },
		{ "Both", "123", "123", false },
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Setenv("GHO_CLIENT_ID", testCase.clientID)
			t.Setenv("GHO_CLIENT_SECRET", testCase.clientSecret)
			var srv server
			if err := srv.loadSecrets(); (err != nil) != testCase.errorExpected {
				t.Errorf("Expected an error: %t - Actual error: %v", testCase.errorExpected, err)
			}
		})
	}
}

func TestLoadTemplates(t *testing.T) {
	testCases := []struct { name string; files map[string]string; errorExpected bool } {
		{ "No template files", map[string]string {}, true },
		{ "A single empty template file", map[string]string { "emptyfile": "" }, false },
		{ "A single template file", map[string]string { "testfile": loginHTML }, false },
		{ "Multiple template files", map[string]string { "testfile1": loginHTML, "testfile2": userHTML }, false },
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			dir := t.TempDir()
			for file, content := range testCase.files {
				if err := os.WriteFile(filepath.Join(dir, file), []byte(content), os.ModePerm); err != nil {
					t.Fatalf("Unable to create temp file: %v", err)
				}
			}
			var srv server
			if err := srv.loadTemplates(filepath.Join(dir, "*")); (err != nil) != testCase.errorExpected {
				t.Errorf("Expected an error: %t - Actual error: %v", testCase.errorExpected, err)
			}
		})
	}
}

func TestSetupHTTPServer(t *testing.T) {
	var srv server
	if err := srv.setupHTTPServer("abc"); err != nil {
		t.Errorf("Unexpected error when setting up HTTP server: %v", err)
	}
	if err := srv.setupHTTPServer(":8080"); err != nil {
		t.Errorf("Unexpected error when setting up HTTP server: %v", err)
	}
}

func TestExecuteTemplate(t *testing.T) {
	srv, err := setupSimpleServer()
	if err != nil {
		t.Fatalf("Unable to setup HTTP test server: %v", err)
	}

	testCases := []struct { name string; templateName string; status int; body string } {
		{ "Execute login.html", "login.html", http.StatusOK, loginHTML },
		{ "Execute foo", "foo", http.StatusInternalServerError, errorHTML },
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			responseRecorder := httptest.NewRecorder()
			srv.executeTemplate(responseRecorder, testCase.templateName, nil);
			assertResponse(t, responseRecorder, testCase.status, testCase.body)
		})
	}
}

func TestUnauthHandler(t *testing.T) {
	srv, err := setupSimpleServer()
	if err != nil {
		t.Fatalf("Unable to setup HTTP test server: %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	responseRecorder := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/", srv.unauthHandler)
	router.ServeHTTP(responseRecorder, request)

	assertResponse(t, responseRecorder, http.StatusOK, loginHTML)
}

func setupSimpleServer() (server, error) {
	var srv server
	err := srv.loadSecrets()
	if err != nil { return srv, err }
	if srv.templates, err = template.New("login.html").Parse(loginHTML); err != nil { return srv, err }
	if srv.templates, err = srv.templates.New("user.html").Parse(userHTML); err != nil { return srv, err }
	if srv.templates, err = srv.templates.New("error.html").Parse(errorHTML); err != nil { return srv, err }
	return srv, nil
}

func assertResponse(t *testing.T, rr *httptest.ResponseRecorder, status int, body string) {
	if rr.Code != status {
		t.Errorf("Expected status: %d - Actual status: %d", status, rr.Code)
	}
	if strings.TrimSpace(rr.Body.String()) != body {
		t.Errorf("Expected response: %s - Actual response: %s", body, rr.Body)
	}
}
