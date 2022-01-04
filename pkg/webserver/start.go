package webserver

import (
	"net/http"
	"html/template"
	"log"
	"os"
	"os/signal"
	"context"
	"errors"
	"time"
	"sync"
	"github.com/gorilla/mux"
)

type server struct {
	http http.Server
	templates *template.Template
	clientID, clientSecret string
	requestCache map[string]requestCacheEntry
	collabGraph map[string]userEntry
	requestMutex *sync.Mutex
}

func Start(address, public, templates, cache string) error {
	log.SetPrefix("[SETUP] ")
	log.Printf("Setting up server...")

	// Initialize server
	var srv server
	var err error
	srv.requestMutex = &sync.Mutex{}
	if srv.clientID, srv.clientSecret, err = loadSecrets(); err != nil { return err }
	if srv.templates, err = loadTemplates(templates); err != nil { return err }
	if srv.requestCache, srv.collabGraph, err = readCacheFromDisk(cache); err != nil { return err }
	srv.setupHTTPServer(address, public)

	log.Printf(
		"Loaded %d cached requests and %d users from cache.",
		len(srv.requestCache), len(srv.collabGraph))

	// Save cache every 30 seconds
	quitChan := make(chan bool, 1)
	go srv.startCacheAutoWriter(cache, quitChan)

	// Shutdown server after receiving a signal
	sigChan := make(chan os.Signal, 1)
	go srv.startSignalHandler(sigChan)

	// Start blocking HTTP server
	log.Printf("Server is up and listening at %s", address)
	log.SetPrefix("")
	err = srv.http.ListenAndServe()
	if err == http.ErrServerClosed { err = nil }

	// Wait for everything to stop
	quitChan <- true
	<-quitChan
	log.Printf(
		"Saved %d cached requests and %d users to cache.",
		len(srv.requestCache), len(srv.collabGraph))

	return err
}

func loadSecrets() (string, string, error) {
	log.Printf("Reading client secrets from environment variables...")
	clientID := os.Getenv("GHO_CLIENT_ID")
	clientSecret := os.Getenv("GHO_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" { return "", "", errors.New("Missing env vars") }
	return clientID, clientSecret, nil
}

func loadTemplates(pattern string) (*template.Template, error) {
	log.Printf("Parsing HTML template files matching %s...", pattern)
	return template.ParseGlob(pattern)
}

func (srv *server) setupHTTPServer(address, public string) {
	log.Printf("Registering HTTP routes...")

	hasGHOCookie := func(r *http.Request, rm *mux.RouteMatch) bool {
		_, err := r.Cookie("gho")
		return err == nil
	}

	isWebSocketRequest := func(r *http.Request, rm *mux.RouteMatch) bool {
		return r.Header.Get("Connection") == "Upgrade" && r.Header.Get("Upgrade") == "websocket"
	}

	r := mux.NewRouter()
	r.HandleFunc("/", srv.oauthHandler).Queries("code", "{code}")
	r.HandleFunc("/", srv.wsHandler).MatcherFunc(hasGHOCookie).MatcherFunc(isWebSocketRequest)
	r.HandleFunc("/", srv.userHandler).MatcherFunc(hasGHOCookie)
	r.HandleFunc("/", srv.unauthHandler)
	r.PathPrefix("/").Handler(http.FileServer(http.Dir(public)))
	srv.http = http.Server { Addr: address, Handler: r }
}

func (srv *server) startCacheAutoWriter(cache string, quitChan chan bool) {
	writeCache := func() {
		err := writeCacheToDisk(cache, srv.requestCache, srv.collabGraph)
		if err != nil { log.Printf("Warning - Error while writing to cache: %v", err) }
	}

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		writeCache()
		select {
		case <-ticker.C:
		case <-quitChan:
			quitChan <- true
			return
		}
	}
}

func (srv *server) startSignalHandler(sigChan chan os.Signal) {
	signal.Notify(sigChan, os.Interrupt)
	<-sigChan
	log.SetPrefix("[SHUTDOWN] ")
	log.Printf("Shutting down server...")
	srv.http.Shutdown(context.Background())
	log.Printf("Server shutdown.")
}
