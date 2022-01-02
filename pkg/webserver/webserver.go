package webserver

import (
	"net/http"
	"html/template"
	"log"
	"os"
	"os/signal"
	"context"
	"errors"
	"github.com/gorilla/mux"
)

type server struct {
	http http.Server
	templates *template.Template
	clientID string
	clientSecret string
}

func Start(address, templates string) error {
	var srv server
	if err := srv.loadSecrets(); err != nil { return err }
	if err := srv.loadTemplates(templates); err != nil { return err }
	if err := srv.setupHTTPServer(address); err != nil { return err }

	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, os.Interrupt)
	go func() {
		<-sigChan
		srv.http.Shutdown(context.Background())
		log.Printf("Server shutdown")
	}()

	log.Printf("Server is up and listening at %s", address)
	return srv.http.ListenAndServe()
}

func (srv *server) loadSecrets() error {
	log.Printf("| Reading client secrets from environment variables...")
	srv.clientID = os.Getenv("GHO_CLIENT_ID")
	srv.clientSecret = os.Getenv("GHO_CLIENT_SECRET")
	if srv.clientID == "" || srv.clientSecret == "" {
		return errors.New("Missing required environment variable")
	}
	return nil
}

func (srv *server) loadTemplates(pattern string) error {
	log.Printf("| Parsing HTML template files matching %s...", pattern)
	var err error
	srv.templates, err = template.ParseGlob(pattern)
	return err
}

func (srv *server) setupHTTPServer(address string) error {
	log.Printf("| Registering HTTP routes...")
	r := mux.NewRouter()
	r.HandleFunc("/", srv.oauthHandler).Queries("code", "{code}")
	r.HandleFunc("/", srv.userHandler).Queries("u", "{token}")
	r.HandleFunc("/", srv.unauthHandler)
	srv.http = http.Server { Addr: address, Handler: r }
	return nil
}

func (srv *server) oauthHandler(w http.ResponseWriter, r *http.Request) {

}

func (srv *server) userHandler(w http.ResponseWriter, r *http.Request) {

}

func (srv *server) unauthHandler(w http.ResponseWriter, r *http.Request) {

}
