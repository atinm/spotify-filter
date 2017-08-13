package main

import (
	"log"
	"time"

	"github.com/zmb3/spotify"
)

const (
	MAX_SLEEP_TIME = 5 * time.Second
	MIN_SLEEP_TIME = 1 * time.Second
)

func min(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
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

		if playerState.Device.Type != "Smartphone" && playerState.Device.Active && !playerState.Device.Restricted && playerState.Playing {
			track = playerState.Item
			currentTrackUri := track.URI
			if lastTrackUri != currentTrackUri {
				log.Printf("[DEBUG] Found track '%s' by '%s' playing", track.Name, track.Artists[0])
				if Rules(track, playerState.Device.Name) {
					log.Printf("[INFO] Skipped track '%s' by '%s' playing on '%s'", track.Name, track.Artists[0], playerState.Device.Name)
					client.Next()
					sleepDuration = MIN_SLEEP_TIME
				}

				lastTrackUri = currentTrackUri
			} else {
				timeLeft := time.Duration(track.Duration-playerState.Progress) * time.Millisecond
				sleepDuration = min(timeLeft, MAX_SLEEP_TIME)
			}
		} else {
			sleepDuration = MAX_SLEEP_TIME
		}

		// restart timer
		ticker = time.NewTicker(sleepDuration)
	}
}
