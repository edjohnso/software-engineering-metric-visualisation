package webserver

import (
	"log"
	"os"
	"errors"
	"compress/gzip"
	"encoding/gob"
)

type diskCacheFormat struct {
	Requests map[string]requestCacheEntry
	CollabGraph map[string]userEntry
}

func readCacheFromDisk(file string) (map[string]requestCacheEntry, map[string]userEntry, error) {

	// Open file to read
	f, err := os.Open(file)
	if errors.Is(err, os.ErrNotExist) {
		log.Printf("Cache not found, loaded no data.")
		return map[string]requestCacheEntry{}, map[string]userEntry{}, nil
	}
	if err != nil { return nil, nil, err }
	defer f.Close()

	// Use GZip decompression
	z, err := gzip.NewReader(f)
	if err != nil { return nil, nil, err }
	defer z.Close()

	// Decompress and decode file to cache
	var cache diskCacheFormat
	err = gob.NewDecoder(z).Decode(&cache)
	if err != nil { return nil, nil, err }

	return cache.Requests, cache.CollabGraph, nil
}

func writeCacheToDisk(file string, requests map[string]requestCacheEntry, collabGraph map[string]userEntry) error {

	// Create file to write to
	f, err := os.Create(file)
	if err != nil { return err }
	defer f.Close()

	// Use GZip compression
	z := gzip.NewWriter(f)
	defer z.Close()

	// Encode and compress cache to file
	cache := diskCacheFormat { requests, collabGraph }
	err = gob.NewEncoder(z).Encode(cache)
	if err != nil { return err }

	return nil
}
