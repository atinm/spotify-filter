package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/hashicorp/logutils"
	"github.com/satori/go.uuid"
	"github.com/skratchdot/open-golang/open"
	"github.com/zmb3/spotify"
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
	redirectURI = "https://localhost:5009/callback"
	rule        = Rule{Explicit: true}
	client      *spotify.Client
	track       *spotify.FullTrack
	roomName    string
	auth        spotify.Authenticator
	ch          = make(chan *spotify.Client)
	state       string
	certificate = "cert.pem"
	key         = "key.pem"
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

	if st := req.FormValue("state"); st != state {
		http.NotFound(w, req)
		log.Fatalf("State mismatch: %s != %s\n", st, state)
	}

	tok.AccessToken = req.FormValue("access_token")
	tok.TokenType = req.FormValue("token_type")
	tok.RefreshToken = req.FormValue("refresh_token")
	expires, _ := strconv.Atoi(req.FormValue("expiry"))
	if expires != 0 {
		tok.Expiry = time.Now().Add(time.Duration(expires) * time.Second)
	}

	// use the token to get an authenticated client
	client := auth.NewClient(&tok)

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, "Monitoring authorization completed! You can close this window now.")
	ch <- &client
}

func main() {
	logFilter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG", "INFO", "WARN", "ERROR"},
		MinLevel: logutils.LogLevel("WARN"),
		Writer:   os.Stderr,
	}
	log.SetOutput(logFilter)

	conf, err := os.Open("config.json")
	if err != nil {
		log.Print("[DEBUG] No config file specified, ignoring.")
	} else {
		defer conf.Close()

		decoder := json.NewDecoder(conf)
		err = decoder.Decode(&config)
		if err != nil {
			log.Fatalf("Config file 'config.json could not be read, %v", err)
		}
		if config.LogLevel != "" {
			logFilter.SetMinLevel(config.LogLevel)
		}
		if config.RedirectURI != "" {
			redirectURI = config.RedirectURI
		}
		if config.CertificateFile != "" {
			certificate = config.CertificateFile
		}
		if config.KeyFile != "" {
			key = config.KeyFile
		}
	}

	router := mux.NewRouter()

	router.HandleFunc("/callback", completeAuth).Methods("GET")
	router.HandleFunc("/filter", GetFilter).Methods("GET")
	router.HandleFunc("/filter", ToggleFilter).Methods("PUT")

	sonos := router.PathPrefix("/sonos").Subrouter()
	sonos.HandleFunc("/updates", HandleUpdate).Methods("POST")

	auth = spotify.NewAuthenticator(redirectURI, spotify.ScopeUserReadCurrentlyPlaying, spotify.ScopeUserReadPlaybackState, spotify.ScopeUserModifyPlaybackState)
	go func() {
		if config.ClientId != "" {
			auth.SetAuthInfo(config.ClientId, "")
		}
		state = uuid.NewV4().String()
		url := auth.AuthURL(state)
		err := open.Run(url + "&show_dialog=true")
		if err != nil {
			log.Fatalf("Could not open %s: %v", url, err)
		}

		// wait for auth to complete
		client = <-ch

		// use the client to make calls that require authorization
		user, err := client.CurrentUser()
		if err != nil {
			log.Fatal(err)
		}
		log.Println("[DEBUG] You are logged in as:", user.ID)
		go Monitor()
	}()

	log.Fatal(http.ListenAndServeTLS(":5007", certificate, key, router))
}
