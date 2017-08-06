package main

import (
	"github.com/zmb3/spotify"
	"log"
	"time"
)

const (
	MAX_SLEEP_TIME = int64(5000)
	MIN_SLEEP_TIME = int64(1000)
)

func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func Monitor() {
	sleepDuration := MAX_SLEEP_TIME
	lastTrackUri := spotify.URI("")
	log.Println("[INFO] Started monitoring Spotify")

	for {
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
				sleepDuration = min(int64(track.Duration - playerState.Progress), MAX_SLEEP_TIME)
			}
		}
		time.Sleep(time.Duration(sleepDuration) * time.Millisecond)
	}
}
