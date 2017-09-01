package lib

import (
	"log"
	"time"

	"github.com/atinm/spotify"
	uuid "github.com/satori/go.uuid"
)

const (
	MAX_SLEEP_TIME     = 15 * time.Second
	MIN_SLEEP_TIME     = 5 * time.Second
	DEEP_SLEEP_TIME    = 1 * time.Minute
	DEEP_SLEEP_COUNTER = 20 // 5 minutes
)

var (
	deepSleepCounter = 0
)

func min(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}

func GetAuthURL() string {
	auth = spotify.NewAuthenticator(redirectURI, spotify.ScopeUserReadCurrentlyPlaying, spotify.ScopeUserReadPlaybackState, spotify.ScopeUserModifyPlaybackState)
	if config.ClientId != "" {
		auth.SetAuthInfo(config.ClientId, "")
	}

	auth.SetTokenURL(authTokenURL)

	state = uuid.NewV4().String()
	log.Print("[DEBUG] created state:", state)
	return auth.AuthURL(state) // + "&show_dialog=true"
}

func Authenticate() {

	go func() {
		// wait for auth to complete
		client = <-ch

		// use the client to make calls that require authorization
		user, err := client.CurrentUser()
		if err != nil {
			log.Fatal(err)
		}
		log.Println("[DEBUG] You are logged in as:", user.ID)

		if FiltersEnabled() {
			go Monitor()
		}
	}()
}

func StartMonitor() {
	go Monitor()
}

func Monitor() {
	sleepDuration := MAX_SLEEP_TIME
	lastTrackUri := spotify.URI("")

	log.Println("[INFO] Started monitoring Spotify")
	ticker := time.NewTicker(MIN_SLEEP_TIME)
	for {
		// no filters are enabled, no need to monitor
		if !FiltersEnabled() {
			log.Print("[INFO] Stopped monitoring Spotify")
			monitoring = false
			ticker.Stop()
			break
		}
		monitoring = true

		select {
		case <-ticker.C:
			ticker.Stop()
		}

		playerState, err := client.PlayerState()
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("[DEBUG] Found your %s (%s)\n", playerState.Device.Type, playerState.Device.Name)

		if playerState.Playing {
			log.Print("[DEBUG] playerState is Playing")
			track = playerState.Item
			if playerState.Device.Type != "Smartphone" && playerState.Device.Active && !playerState.Device.Restricted {
				currentTrackUri := track.URI
				if lastTrackUri != currentTrackUri {
					log.Printf("[DEBUG] Found track '%s' by '%s' playing", track.Name, track.Artists[0])
					if Rules(track, playerState.Device.Name) {
						log.Printf("[INFO] Skipped track '%s' by '%s' playing on '%s'", track.Name, track.Artists[0], playerState.Device.Name)
						client.Next()
					}
					// minimum because player is active, could be manually skipped etc
					sleepDuration = MIN_SLEEP_TIME
					lastTrackUri = currentTrackUri
				}
			}
			timeLeft := time.Duration(track.Duration-playerState.Progress) * time.Millisecond
			sleepDuration = min(timeLeft, sleepDuration)
			deepSleepCounter = 0
		} else {
			deepSleepCounter += 1
			if deepSleepCounter >= DEEP_SLEEP_COUNTER {
				sleepDuration = DEEP_SLEEP_TIME
				deepSleepCounter = DEEP_SLEEP_COUNTER
			} else {
				sleepDuration = MAX_SLEEP_TIME
			}
		}

		// restart timer
		if sleepDuration <= 0 {
			sleepDuration = MIN_SLEEP_TIME
		}
		ticker = time.NewTicker(sleepDuration)
	}
}
