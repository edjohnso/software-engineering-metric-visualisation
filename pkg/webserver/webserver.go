package webserver

import (
	"net/http"
	"html/template"
	"log"
	"os"
	"errors"
	"github.com/gorilla/mux"
)

type server struct {
	http http.Server
	templates *template.Template
	clientID string
	clientSecret string
}

func Start() error {
	return errors.New("Not implemented!")
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

func (srv *server) setupHTTPServer() error {
	log.Printf("| Registering HTTP routes...")
	r := mux.NewRouter()
	r.HandleFunc("/", srv.oauthHandler).Queries("code", "{code}")
	r.HandleFunc("/", srv.userHandler).Queries("u", "{token}")
	r.HandleFunc("/", srv.unauthHandler)
	srv.http = http.Server { Addr: ":http", Handler: r }
	return nil
}

func (srv *server) oauthHandler(w http.ResponseWriter, r *http.Request) {

}

func (srv *server) userHandler(w http.ResponseWriter, r *http.Request) {

}

func (srv *server) unauthHandler(w http.ResponseWriter, r *http.Request) {

}
