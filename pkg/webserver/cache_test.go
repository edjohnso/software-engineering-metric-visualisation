package webserver

import (
	"testing"
	"os"
	"path/filepath"
	"time"
)

func TestReadCacheFromDisk(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "cache.gz")

	if err := os.WriteFile(file, []byte(""), os.ModePerm); err != nil {
		t.Fatalf("Unable to create %s file: %v", file, err)
	}
	if _, _, err := readCacheFromDisk(file); err == nil {
		t.Errorf("Expected error when reading from invalid file")
	}

	requests := map[string]requestCacheEntry{}
	requests["testrequest"] = requestCacheEntry { time.Now(), "xyz", response { 200, nil, []byte("testresponse") } }
	collabGraph := map[string]userEntry{}
	collabGraph["testuser"] = userEntry { 0, []string { "testcollaborator" } }

	if err := writeCacheToDisk(file, requests, collabGraph); err != nil {
		t.Fatalf("Unable to write test cache file %s: %v", file, err)
	}
	if requestsCached, collabGraphCached, err := readCacheFromDisk(file); err != nil {
		t.Errorf("Expected no error, actually received: %v", err)
	} else {
		if len(requestsCached) != len(requests) {
			t.Errorf("Cached requests contains %d entries but %d were expected", len(requestsCached), len(requests))
		}
		for request, cacheEntry := range requests {
			if cacheEntryCached, ok := requestsCached[request]; !ok {
				t.Errorf("Cached requests missing %s", request)
			} else {
				if cacheEntryCached.ETag != cacheEntry.ETag {
					t.Errorf("Expected ETag: %s - Actual ETag: %s", cacheEntry.ETag, cacheEntryCached.ETag)
				}
				assertWebserverResponse(
					t, cacheEntryCached.Response,
					cacheEntry.Response.Status,
					string(cacheEntry.Response.Body))
			}
		}

		if len(collabGraphCached) != len(collabGraph) {
			t.Errorf("Cached collab graph contains %d entries but %d were expected", len(collabGraphCached), len(collabGraph))
		}
		for user, entry := range collabGraph {
			if cachedEntry, ok := collabGraphCached[user]; !ok {
				t.Errorf("Cached collab graph missing %s", user)
			} else {
				if entry.RequestedDepth != cachedEntry.RequestedDepth {
					t.Errorf("Expected depth: %d - Actual depth: %d", entry.RequestedDepth, cachedEntry.RequestedDepth)
				}
				match := len(entry.Collaborators) == len(cachedEntry.Collaborators)
				for i := range entry.Collaborators {
					if match && cachedEntry.Collaborators[i] != entry.Collaborators[i] {
						match = false
					}
				}
				if !match {
					t.Errorf(
						"Expected collaborators do not match. Expected: %v. Actual: %v",
						entry.Collaborators, cachedEntry.Collaborators)
				}
			}
		}
	}
}

func TestWriteCacheToDisk(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "cache.gz")

	requests := map[string]requestCacheEntry{}
	requests["testrequest"] = requestCacheEntry { time.Now(), "xyz", response { 200, nil, []byte("testresponse") } }
	collabGraph := map[string]userEntry{}
	collabGraph["testuser"] = userEntry { 0, []string { "testcollaborator" } }

	if err := writeCacheToDisk(file, requests, collabGraph); err != nil {
		t.Errorf("Unable to write cache file %s: %v", file, err)
	}
	if _, err := os.Stat(file); err != nil {
		t.Errorf("Unable to access new cache file %s: %v", file, err)
	}
}
