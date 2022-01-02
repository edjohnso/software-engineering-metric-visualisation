package webserver

import (
	"testing"
	"path/filepath"
	"os"
)

const loginHTML = `<p>login page!</p>`
const userHTML = `<p>user page!</p>`
const errorHTML = `<p>error page!</p>`

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
			if _, _, err := loadSecrets(); (err != nil) != testCase.errorExpected {
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
			if _, err := loadTemplates(filepath.Join(dir, "*")); (err != nil) != testCase.errorExpected {
				t.Errorf("Expected an error: %t - Actual error: %v", testCase.errorExpected, err)
			}
		})
	}
}

func TestSetupHTTPServer(t *testing.T) {
	if _, err := setupHTTPServer(); err != nil {
		t.Errorf("Failed to setup HTTP server: %v", err)
	}
}
