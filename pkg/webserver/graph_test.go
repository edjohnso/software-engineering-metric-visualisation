package webserver

import (
	"testing"
	"net/http"
	"net/http/httptest"
	"os"
)

func TestAddCollaborators(t *testing.T) {

	pat := os.Getenv("GHO_PAT")

	t.Run("Attempt to add collaborators of invalid user", func(t *testing.T) {
		srv, err := setupTestServer()
		if err != nil { t.Fatalf("Failed to setup test server: %v", err) }
		rr := httptest.NewRecorder()
		srv.addCollaborators(rr, pat, "not_a_real_username_so_this_should_error")
		if _, ok := srv.collabGraph["edjohnso"]; ok {
			t.Errorf("Added user entry for not_a_real_username_so_this_should_error")
		}
		assertResponseRecorder(t, rr, http.StatusNotFound, errorHTML)
	})

	t.Run("Add edjohnso collaborators", func(t *testing.T) {
		srv, err := setupTestServer()
		if err != nil { t.Fatalf("Failed to setup test server: %v", err) }
		rr := httptest.NewRecorder()
		srv.addCollaborators(rr, pat, "edjohnso")
		if entry, ok := srv.collabGraph["edjohnso"]; !ok {
			t.Errorf("Failed to set user entry for edjohnso")
		} else if entry.Collaborators == nil {
			t.Errorf("Failed to set collaborators for edjohnso")
		} else {
			collaborators := entry.Collaborators
			expected := []string { "edjohnso", "tedski999" }
			match := len(collaborators) == len(expected)
			for i := range expected {
				if match && collaborators[i] != expected[i] {
					match = false
				}
			}
			if !match {
				t.Errorf("Expected collaborators do not match. Expected: %v. Actual: %v", expected, collaborators)
			}
		}
	})
}

func TestCheckForTarget(t *testing.T) {
	srv, err := setupTestServer()
	if err != nil { t.Fatalf("Failed to setup test server: %v", err) }
	srv.checkForTarget("", "", nil)
}
