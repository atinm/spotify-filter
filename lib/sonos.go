package lib

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/atinm/go-sonos"
	"github.com/atinm/go-sonos/didl"
	"github.com/atinm/go-sonos/ssdp"
	"github.com/atinm/go-sonos/upnp"
	"github.com/zmb3/spotify"
)

var (
	sonosDevices []*sonos.Sonos
	exit_chan    = make(chan bool)
)

// GetLocalInterfaceName returns the first interface name that has the non loopback local IPv4 addr of the host
func getLocalInterfaceName() string {
	list, err := net.Interfaces()
	if err != nil {
		panic(err)
	}

	for _, iface := range list {
		addrs, err := iface.Addrs()
		if err != nil {
			panic(err)
		}
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					return iface.Name
				}
			}
		}
	}
	return ""
}

func HandleUpdate(w http.ResponseWriter, req *http.Request) {
	var update Update
	body, err := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	if err != nil {
		log.Print("[ERROR] ", err)
		return
	}
	if err = json.Unmarshal(body, &update); err != nil {
		log.Print("[ERROR] ", err)
		return
	}

	switch update.Type {
	case "transport-state":
		var s struct {
			Update
			TransportState
		}
		if err = json.Unmarshal(body, &s); err != nil {
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
		// CurrentTrackURI: "x-sonos-spotify:spotify%3atrack%3a4Ro98RCK90oHqqSZUnTFq5?sid=12&flags=8224&sn=8"
		// "uri": "x-sonos-spotify:spotify%3atrack%3a6tF92PMv01Ug9Dh8Rmy6nH?sid=12&flags=8224&sn=8",
		if strings.HasPrefix(uri, "x-sonos-spotify:spotify%3atrack%3a") {
			// extract the trackid from the uri
			trackID := strings.Split(strings.TrimPrefix(uri, "x-sonos-spotify:spotify%3atrack%3a"), "?")[0]
			track, err = client.GetTrack(spotify.ID(trackID))
			if err != nil {
				log.Printf("[WARN] Could not get track info for '%s' by '%s', trackId: (%s): %v\n", title, artist, trackID, err)
				break
			}

			if Rules(track, roomName) {
				nextURL := fmt.Sprintf("http://localhost:5005/%s/next", roomName)
				response, err := http.Get(nextURL)
				if err != nil {
					log.Printf("[WARN] Could not skip track '%s' by '%s': %v", title, artist, err)
					return
				}
				defer response.Body.Close()
				log.Printf("[INFO] Skipped '%s' by '%s' playing on '%s'\n", title, artist, roomName)
			}
		}

	case "volume-change":
	case "mute-change":
	case "topology-change":
	default:
	}
}

func getTriggeredSonos(svc *upnp.Service) (sonos *sonos.Sonos) {
	for _, s := range sonosDevices {
		if s.AVTransport.Svc == svc {
			sonos = s
			break
		}
	}
	return
}

func handleAVTransportEvents(reactor upnp.Reactor, c chan bool) {
	for {
		select {
		case evt := <-reactor.Channel():
			switch evt.Type() {
			case upnp.AVTransport_EventType:
				var artist, title, album string

				b := evt.(upnp.AVTransportEvent)
				log.Printf("[DEBUG] TransportState: %v", b.LastChange.InstanceID.TransportState.Val)
				if b.LastChange.InstanceID.TransportState.Val != "PLAYING" {
					continue
				}
				uri := b.LastChange.InstanceID.CurrentTrackURI.Val
				log.Printf("[DEBUG] CurrentTrackURI: %v", b.LastChange.InstanceID.CurrentTrackURI.Val)
				log.Printf("[DEBUG] CurrentTrackMetadata: %s", b.LastChange.InstanceID.CurrentTrackMetaData.Val)
				if strings.HasPrefix(uri, "x-sonos-spotify:spotify%3atrack%3a") {
					log.Print("Playing Spotify track")
					if b.LastChange.InstanceID.CurrentTrackMetaData.Val != "" {
						var doc didl.Lite
						err := xml.Unmarshal([]byte(b.LastChange.InstanceID.CurrentTrackMetaData.Val), &doc)
						if err != nil {
							log.Printf("[ERROR] Could not unmarshal %s: %v", b.LastChange.InstanceID.CurrentTrackMetaData.Val, err)
						}
						for _, item := range doc.Item {
							artist = item.Creator[0].Value
							album = item.Album[0].Value
							title = item.Title[0].Value
							log.Printf("[DEBUG] title: %s, artist: %s, album: %s", title, artist, album)
							break
						}
					}

					// extract the trackid from the uri
					trackID := strings.Split(strings.TrimPrefix(uri, "x-sonos-spotify:spotify%3atrack%3a"), "?")[0]
					track, err := client.GetTrack(spotify.ID(trackID))
					if err != nil {
						log.Printf("[WARN] Could not get track info for trackId: (%s): %v\n", trackID, err)
						continue
					}
					player := getTriggeredSonos(b.Svc)
					if player == nil {
						log.Printf("[WARN] Could not skip track '%s' by '%s': (did not find player)", title, artist)
						continue
					}

					if Rules(track, player.Player.RoomName) {
						posInfo, err := player.GetPositionInfo(0)
						if nil != err {
							panic(err)
						}
						log.Printf("[DEBUG] Position.TrackURI: %s", posInfo.TrackURI)
						log.Printf("[DEBUG] Position.TrackDuration: %s", posInfo.TrackDuration)
						log.Printf("[DEBUG] Position.RelTime: %s", posInfo.RelTime)

						err = player.Next(0)
						if err != nil {
							log.Printf("[WARN] Could not skip track '%s' by '%s': %v", title, artist, err)
							continue
						}
						log.Printf("[INFO] Skipped '%s' by '%s' playing on '%s'\n", title, artist, player.Player.RoomName)
					}
				} else {
					log.Printf("Not playing Spotify track: %s", uri)
				}
			default:
				log.Panicf("[ERROR] Unexpected event %#v", evt)
			}
		}
	}
}

func SetupEvents(mgr ssdp.Manager) {
	// Startup and listen to events
	reactor := sonos.MakeReactor(port)
	go handleAVTransportEvents(reactor, exit_chan)
	sonosDevices = sonos.ConnectAll(mgr, reactor, sonos.SVC_AV_TRANSPORT)
	log.Printf("[DEBUG] Set up event handler for Sonos, found %d devices", len(sonosDevices))
	for _, player := range sonosDevices {
		log.Printf("Found %s", player.RoomName)
	}
}

func InitializeSonos() {
	mgr := ssdp.MakeManager()
	ifname := getLocalInterfaceName()
	log.Printf("Discovering devices over %s...", ifname)
	if err := mgr.Discover(ifname, port, false); nil != err {
		panic(err)
	} else {
		SetupEvents(mgr)
	}
}
