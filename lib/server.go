package lib

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/atinm/spotify"
	"github.com/gorilla/mux"
	"github.com/hashicorp/logutils"
	"golang.org/x/oauth2"
)

var (
	config Config
	// redirectURI is the OAuth redirect URI for the application.
	// You must register an application at Spotify's developer portal
	// and enter this value. This is the address where your authorization
	// server runs. The authorization server is the server that contains both
	// the client id and client secret and can get the access token and refresh
	// token from Spotify and return to this application on its own callback as
	// query parameters
	redirectURI  = "https://localhost:5009/callback"
	authTokenURL = "http://localhost:5009/token"
	rule         = Rule{Explicit: true}
	client       *spotify.Client
	track        *spotify.FullTrack
	roomName     string
	auth         spotify.Authenticator
	ch           = make(chan *spotify.Client)
	state        string
	certificate  = "cert.pem"
	key          = "key.pem"
	port         = "5007"
	monitoring   = false
	srv          *http.Server
	LogFilter    *logutils.LevelFilter
)

func GetFilter(w http.ResponseWriter, req *http.Request) {
	json.NewEncoder(w).Encode(rule)
}

func ToggleFilter(w http.ResponseWriter, req *http.Request) {
	rule.Explicit = !rule.Explicit
	json.NewEncoder(w).Encode(rule)
}

func completeAuth(w http.ResponseWriter, req *http.Request) {
	var tok oauth2.Token
	log.Print("[DEBUG] Received callback to completeAuth")

	if st := req.FormValue("state"); st != state {
		http.NotFound(w, req)
		log.Fatalf("State mismatch: received %s != created %s\n", st, state)
	}

	tok.AccessToken = req.FormValue("access_token")
	tok.TokenType = req.FormValue("token_type")
	tok.RefreshToken = req.FormValue("refresh_token")
	expires, _ := strconv.Atoi(req.FormValue("expiry"))
	if expires != 0 {
		tok.Expiry = time.Now().Add(time.Duration(expires) * time.Second)
	}
	log.Print("[DEBUG] Getting new authenticated client")
	// use the token to get an authenticated client
	client := auth.NewClient(&tok)

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, "Monitoring authorization completed! You can close this window now.")
	ch <- &client

	// we are done with the server
	log.Print("[DEBUG] Shutting down authentication listener")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(ctx)

	// now update the token url to handle fresh
	auth.SetTokenURL(authTokenURL)

	// now we can start sonos as well
	InitializeSonos()
}

func StartServer() {
	go Server()
}

func Server() {
	router := mux.NewRouter()

	router.HandleFunc("/callback", completeAuth).Methods("GET")
	router.HandleFunc("/filter", GetFilter).Methods("GET")
	router.HandleFunc("/filter", ToggleFilter).Methods("PUT")

	srv = &http.Server{Addr: ":" + port, Handler: router}
	log.Print("[DEBUG] Listening on " + port)
	srv.ListenAndServeTLS(certificate, key)
}
