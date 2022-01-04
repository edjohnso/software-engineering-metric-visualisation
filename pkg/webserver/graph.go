package webserver

import (
	"log"
	"net/http"
	"encoding/json"
	"sort"
	"github.com/gorilla/websocket"
)

type userEntry struct {
	RequestedDepth int
	Collaborators []string
}

func (srv *server) sendUserCollaborators(w http.ResponseWriter, ws *websocket.Conn, auth, username string, collaborators []string) {

	data := userCollaboratorsFormat {
		username, make([]userFormat, len(collaborators)),
	}

	// Get users data
	for i, collaborator := range collaborators {
		resp := srv.requestOK(w, auth, http.MethodGet, "https://api.github.com/users/" + collaborator)
		if (resp.Status >= 400) { log.Printf("GET /users/%s returned %d", collaborator, resp.Status); return }
		json.Unmarshal(resp.Body, &data.Collaborators[i])
	}

	// Send data
	if err := ws.WriteJSON(data); err != nil {
		log.Printf("Unable to write data: %v", err)
	}
}

func (srv *server) addCollaborators(w http.ResponseWriter, auth, username string) {
	log.Printf("Scanning for collaborators of %s...", username)

	// Find users repositories
	resp := srv.requestOK(w, auth, http.MethodGet, "https://api.github.com/users/" + username + "/repos")
	if (resp.Status >= 400) { log.Printf("GET /users/%s/repos returned %d", username, resp.Status) }

	var repos reposFormat
	json.Unmarshal(resp.Body, &repos)

	// Find every contributor to every one of their repositories
	collaborators := map[string]string{}
	for _, repo := range repos {
		resp = srv.requestOK(w, auth, http.MethodGet, "https://api.github.com/repos/" + username + "/" + repo.Name + "/contributors")
		if (resp.Status >= 400) { log.Printf("GET /repos/%s/%s returned %d", username, repo.Name, resp.Status) }

		var contributors contributorsFormat
		json.Unmarshal(resp.Body, &contributors)

		for _, contributor := range contributors {
			collaborators[contributor.Login] = repo.Name
		}
	}

	// Add these collaborators to the graph
	// TODO: do something with repos
	keys := make([]string, 0, len(collaborators))
	for k := range collaborators {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	entry := srv.collabGraph[username]
	entry.Collaborators = keys
	srv.collabGraph[username] = entry
}

func (srv *server) checkForTarget(collaborator string, username string, links map[string]string) {
	log.Printf("Found a unique collaborator: %s is a contributor to a repo belonging to %s", collaborator, username)
	/*
	// if we're looking for a target and we find it
	if collaborator == target {
		path := []string { target }
		for username != rootUsername {
			path = append(path, username)
			username = links[username]
			// TODO: should remove username from links to prevent looping
			// TODO: should username, ok := links[username]
		}

		// TODO: ws.WriteJSON()
	}
	*/
}

