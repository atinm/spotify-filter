package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"github.com/zmb3/spotify"
	"log"
	"net/http"
	"strings"
)

func HandleUpdate(w http.ResponseWriter, req *http.Request) {
	var update Update
	body, err := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	if err != nil {
		log.Print("[ERROR] ", err)
		return
	}
	if err := json.Unmarshal(body, &update); err != nil {
		log.Print("[ERROR] ", err)
		return
	}

	switch update.Type {
	case "transport-state":
		var s struct {
			Update
			TransportState
		}
		if err := json.Unmarshal(body, &s); err != nil {
			log.Print("[ERROR] ", err)
			return
		}

		bodyStr := fmt.Sprintf("%s", body)
		log.Print("[DEBUG] body: ", bodyStr)
		title := s.Player.State.CurrentTrack.Title
    artist := s.Player.State.CurrentTrack.Artist
		uri := s.Player.State.CurrentTrack.Uri
    roomName = s.Player.RoomName
		log.Println("[DEBUG]", title)
		log.Println("[DEBUG]", uri)

                // "uri": "x-sonos-spotify:spotify%3atrack%3a6tF92PMv01Ug9Dh8Rmy6nH?sid=12&flags=8224&sn=8",
		if strings.HasPrefix(uri, "x-sonos-spotify:spotify%3atrack%3a") {
			// extract the trackid from the uri
			trackId := strings.Split(strings.TrimPrefix(uri, "x-sonos-spotify:spotify%3atrack%3a"), "?")[0]
			track, err = client.GetTrack(spotify.ID(trackId))
			if err != nil {
				log.Printf("[WARN] Could not get track info for '%s' by '%s', trackId: (%s): %v\n", title, artist, trackId, err)
				break
			}

			if Rules(track, roomName) {
        nextUrl := fmt.Sprintf("http://localhost:5005/%s/next", roomName)
        response, err := http.Get(nextUrl)
        defer response.Body.Close()
        if err != nil {
          log.Print("[WARN] Could not skip track '%s': %v", title, artist, err)
          return
        }
        log.Printf("[INFO] Skipped '%s' by '%s' playing on '%s'\n", title, artist, roomName)
      }
		}

	case "volume-change":
	case "mute-change":
	case "topology-change":
	default:
	}
}
