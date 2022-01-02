package webserver

import (
	"net/http"
	"net/url"
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
	log.Printf("Setting up server...")
	var srv server
	if err := srv.loadSecrets(); err != nil { return err }
	if err := srv.loadTemplates(templates); err != nil { return err }
	if err := srv.setupHTTPServer(address); err != nil { return err }

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	go func() {
		<-sigChan
		srv.http.Shutdown(context.Background())
		log.Printf("Server shutdown")
	}()

	log.Printf("Server is up and listening at %s", address)
	if err := srv.http.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}

	return nil
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
	resp, body := requestAndParse(
		"", http.MethodPost,
		"https://github.com/login/oauth/access_token" +
			"?client_id=" + srv.clientID +
			"&client_secret=" + srv.clientSecret +
			"&code=" + mux.Vars(r)["code"])
	if (resp.StatusCode >= 400) {
		srv.errorResponse(w, resp.StatusCode)
		return
	}

	rq, err := url.ParseQuery(body)
	if err != nil {
		log.Printf("Unable to parse exchange response body: %v", err)
		srv.errorResponse(w, http.StatusInternalServerError)
	} else if rq.Has("error") || !rq.Has("access_token") {
		w.WriteHeader(http.StatusUnauthorized)
		srv.unauthHandler(w, r)
	} else {
		http.Redirect(w, r, "/?u=" + rq.Get("access_token"), http.StatusSeeOther)
	}
}

	http.Redirect(w, r, "/?u=" + rq.Get("access_token"), http.StatusSeeOther)
}

func (srv *server) userHandler(w http.ResponseWriter, r *http.Request) {
	resp, body := requestAndParse(mux.Vars(r)["token"], http.MethodGet, "https://api.github.com/user")
	if resp.StatusCode == http.StatusUnauthorized {
		w.WriteHeader(http.StatusUnauthorized)
		srv.unauthHandler(w, r)
		return
	} else if (resp.StatusCode >= 400) {
		srv.errorResponse(w, resp.StatusCode)
		return
	}

	// TODO: decode JSON body

	srv.executeTemplate(w, "user.html", body)
}

func (srv *server) unauthHandler(w http.ResponseWriter, r *http.Request) {
	srv.executeTemplate(w, "login.html", struct { ClientID string }{ srv.clientID })
}

func (srv *server) executeTemplate(w http.ResponseWriter, name string, data interface{}) {
	if err := srv.templates.ExecuteTemplate(w, name, data); err != nil {
		log.Printf("Failed to execute '%s' template: %v", name, err)
		srv.errorResponse(w, http.StatusInternalServerError)
	}
}

func (srv *server) errorResponse(w http.ResponseWriter, status int) {
	w.WriteHeader(status)
	msg := http.StatusText(status)
	srv.executeTemplate(w, "error.html", struct { Code int; Message string } { status, msg })
}
