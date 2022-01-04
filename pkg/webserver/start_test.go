package webserver

import (
	"testing"
	"os"
	"path/filepath"
	"syscall"
	"time"
	"fmt"
)

func TestStart(t *testing.T) {

	runTest := func(address, public, templates, cache string) error {

		// Create test directory
		dir := t.TempDir()
		files := map[string]string { "login.html": loginHTML, "graph.html": graphHTML, "error.html": errorHTML }
		for file, content := range files {
			if err := os.WriteFile(filepath.Join(dir, file), []byte(content), os.ModePerm); err != nil {
				t.Fatalf("Unable to create %s file: %v", file, err)
			}
		}

		// Force shutdown after 200ms
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

		// Run server wit provided config
		err := Start(
			address,
			filepath.Join(dir, public),
			filepath.Join(dir, templates),
			filepath.Join(dir, cache))
		ok = true
		return err
	}

	t.Run("Valid config", func(t *testing.T) {
		if err := runTest(":8080", "", "*.html", "cache.gz"); err != nil {
			t.Errorf("Expected no error, actually received: %v", err)
		}
	})

	t.Run("Invalid address", func(t *testing.T) {
		if err := runTest("? ? ?", "", "*.html", "cache.gz"); err == nil {
			t.Errorf("Expected error when providing invalid address")
		}
	})

	t.Run("Missing templates", func(t *testing.T) {
		if err := runTest(":8080", "", "none", "cache.gz"); err == nil {
			t.Errorf("Expected error when providing no matching template files")
		}
	})

	t.Run("Invalid cache", func(t *testing.T) {
		if err := runTest(":8080", "", "*.html", "\x00"); err == nil {
			t.Errorf("Expected error when providing invalid cache file")
		}
	})

	t.Run("Missing envvars", func(t *testing.T) {
		t.Setenv("GHO_CLIENT_ID", "")
		if err := runTest(":8080", "", "*.html", "cache.gz"); err == nil {
			t.Errorf("Expected error when not providing a required environment variable")
		}
	})
}
