package webserver

import (
	"log"
	"sync"
	"net/url"
	"net/http"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type userCollaboratorsFormat struct {
	Username string `json:"username"`
	Collaborators []userFormat `json:"collaborators"`
}

type rootFormat struct {
	Root userFormat `json:"root"`
}

type statusFormat struct {
	Working bool `json:"working"`
	Paused bool `json:"paused"`
	Depth int `json:"depth"`
	MaxDepth int `json:"max_depth"`
}

func (srv *server) wsHandler(w http.ResponseWriter, r *http.Request) {

	// Get auth token from cookie
	authCookie, err := r.Cookie("gho")
	if err != nil {
		log.Printf("Failed to get gho cookie: %v", err)
		srv.errorResponse(w, http.StatusUnauthorized)
		return
	}
	auth := authCookie.Value

	// Attempt to get this users details with their auth token
	resp := srv.requestOK(w, auth, http.MethodGet, "https://api.github.com/user")
	if (resp.Status >= 400) {
		log.Printf("GET /user returned %d", resp.Status)
		srv.errorResponse(w, http.StatusUnauthorized)
		return
	}
	var user userFormat
	json.Unmarshal(resp.Body, &user)

	// Add user to graph if not already
	if _, ok := srv.collabGraph[user.Login]; !ok {
		srv.collabGraph[user.Login] = userEntry { 99, []string{} }
	}

	// Upgrade HTTP connection to WS
	var upgrader = websocket.Upgrader{}
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WS upgrade error: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer ws.Close()

	rootData := rootFormat { user }
	ws.WriteJSON(rootData)

	// Send any loaded collaborators up to requested depth using Breadth-First Traversal
	log.Printf("Sending loaded collaborators to WebSocket client...")
	queue := []string { user.Login }
	links := map[string]string { user.Login: "" }
	for depth := 0; depth <= srv.collabGraph[user.Login].RequestedDepth && len(queue) != 0; depth++ {
		for range queue {

			// Dequeue next and send
			username := queue[0]
			queue = queue[1:]

			// Link and enqueue unique collaborators
			if entry, ok := srv.collabGraph[username]; ok {
				uniques := []string{}

				for _, collaborator := range entry.Collaborators {
					if _, ok = links[collaborator]; !ok {
						links[collaborator] = username
						queue = append(queue, collaborator)
						uniques = append(uniques, collaborator)
					}
				}

				srv.sendUserCollaborators(w, ws, auth, username, uniques)
			}
		}
	}

	// Sync variables
	quit := false
	paused := true
	m := &sync.Mutex{}
	c := sync.NewCond(m)
	depth := 0
	working := false

	// Listen for commands from client
	go func() {
		log.Printf("Listening for commands from WebSocket client...")
		for {
			var data struct { Data string `json:"command"` }
			if err := ws.ReadJSON(&data); err != nil {
				log.Printf("Unable to read data: %v", err)
				c.L.Lock()
				quit = true;
				c.L.Unlock()
				c.Signal()
				return
			}
			c.L.Lock()
			switch data.Data {
			case "plus":
				entry := srv.collabGraph[user.Login]
				entry.RequestedDepth++
				srv.collabGraph[user.Login] = entry
			case "minus":
				entry := srv.collabGraph[user.Login]
				entry.RequestedDepth--
				srv.collabGraph[user.Login] = entry
			case "pause":
				paused = true
			case "continue":
				paused = false
			}
			status := statusFormat { working, paused, depth, srv.collabGraph[user.Login].RequestedDepth }
			ws.WriteJSON(status)
			c.L.Unlock()
			c.Signal()
		}
	}()

	// Grow collaborators graph
	queue = []string { user.Login }
	links = map[string]string { user.Login: "" }
	for depth = 0; len(queue) != 0; depth++ {

		for range queue {

			// Wait until not paused and depth <= max depth
			c.L.Lock()
			for !quit && (paused || depth > srv.collabGraph[user.Login].RequestedDepth) {
				log.Printf("Stopped search (Paused: %t, depth == %d).", paused, depth)
				working = false
				status := statusFormat { working, paused, depth, srv.collabGraph[user.Login].RequestedDepth }
				ws.WriteJSON(status)
				c.Wait()
			}
			working = true
			status := statusFormat { working, paused, depth, srv.collabGraph[user.Login].RequestedDepth }
			ws.WriteJSON(status)
			c.L.Unlock()
			if quit { break }

			// Dequeue next username
			username := queue[0]
			queue = queue[1:]
			srv.addCollaborators(w, auth, username)

			// Link and enqueue unique collaborators
			if entry, ok := srv.collabGraph[username]; ok {
				uniques := []string{}

				for _, collaborator := range entry.Collaborators {
					//if collaborator == "exclude whoever" { continue }
					if paused || depth > srv.collabGraph[user.Login].RequestedDepth || quit { break }
					if _, ok = links[collaborator]; !ok {
						links[collaborator] = username
						queue = append(queue, collaborator)
						uniques = append(uniques, collaborator)
						srv.checkForTarget(collaborator, username, links)
					}
				}

				srv.sendUserCollaborators(w, ws, auth, username, uniques)
			}
		}

		c.L.Lock()
		status := statusFormat { working, paused, depth, srv.collabGraph[user.Login].RequestedDepth }
		ws.WriteJSON(status)
		c.L.Unlock()

		if quit { break }

	}

	log.Printf("Closing WebSocket...")
}

func (srv *server) oauthHandler(w http.ResponseWriter, r *http.Request) {

	// Exchange OAuth code for user access token
	resp := srv.requestOK(
		w, "", http.MethodPost,
		"https://github.com/login/oauth/access_token" +
			"?client_id=" + srv.clientID +
			"&client_secret=" + srv.clientSecret +
			"&code=" + mux.Vars(r)["code"])
	if (resp.Status >= 400) {
		srv.errorResponse(w, resp.Status)
		return
	}

	// Parse user access token from body
	query, err := url.ParseQuery(string(resp.Body))
	if err != nil {
		log.Printf("Unable to parse exchange response body: %v", err)
		srv.errorResponse(w, http.StatusInternalServerError)
		return
	} else if query.Has("error") || !query.Has("access_token") {
		w.WriteHeader(http.StatusUnauthorized)
		srv.unauthHandler(w, r)
		return
	}

	// Add auth token cookie before redirecting
	http.SetCookie(w, &http.Cookie { Name: "gho", Value: query.Get("access_token") })
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (srv *server) userHandler(w http.ResponseWriter, r *http.Request) {

	// Get auth token from cookie
	authCookie, err := r.Cookie("gho")
	if err != nil {
		log.Printf("Failed to get gho cookie: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		srv.unauthHandler(w, r)
		return
	}
	auth := authCookie.Value

	// Attempt to get this users details with their auth token
	resp := srv.requestOK(w, auth, http.MethodGet, "https://api.github.com/user")
	if resp.Status == http.StatusUnauthorized {
		w.WriteHeader(http.StatusUnauthorized)
		srv.unauthHandler(w, r)
		return
	} else if (resp.Status >= 400) {
		srv.errorResponse(w, resp.Status)
		return
	}

	// The client should establish a WebSocket connection upon receiving this
	srv.executeTemplate(w, "graph.html", nil) // TODO: send some template data
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
