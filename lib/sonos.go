package lib

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/atinm/go-sonos"
	"github.com/atinm/go-sonos/didl"
	"github.com/atinm/go-sonos/ssdp"
	"github.com/atinm/go-sonos/upnp"
	"github.com/atinm/spotify"
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
				log.Printf("[DEBUG] CurrentTrack: %s", b.LastChange.InstanceID.CurrentTrack.Val)
				currentTrack, _ := strconv.Atoi(b.LastChange.InstanceID.CurrentTrack.Val)
				if strings.HasPrefix(uri, "x-sonos-spotify:spotify%3atrack%3a") {
					log.Print("Playing Spotify track")
					if b.LastChange.InstanceID.CurrentTrackMetaData.Val != "" {
						var doc didl.Lite
						err := xml.Unmarshal([]byte(b.LastChange.InstanceID.CurrentTrackMetaData.Val), &doc)
						if err != nil {
							log.Printf("[ERROR] Could not unmarshal %s: %v", b.LastChange.InstanceID.CurrentTrackMetaData.Val, err)
							continue
						}
						for _, item := range doc.Item {
							for _, a := range item.Creator {
								artist = a.Value
							}
							for _, l := range item.Album {
								album = l.Value
							}
							for _, t := range item.Title {
								title = t.Value
							}
							log.Printf("[DEBUG] title: %s, artist: %s, album: %s", title, artist, album)
							break
						}
					}

					// extract the trackid from the uri
					parsedTrackURI := strings.Split(strings.TrimPrefix(uri, "x-sonos-spotify:spotify%3atrack%3a"), "?")
					trackID := parsedTrackURI[0]
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
						var query string
						query = fmt.Sprintf("title:%s artist:%s album:%s", title, artist, album)
						log.Printf("[DEBUG] Querying spotify for %s", query)
						// search for radio friendly track for artist, title, album, explicit=false instead
						if res, err := client.Search(query, spotify.SearchTypeTrack); err != nil || len(res.Tracks.Tracks) == 0 {
							log.Printf("[DEBUG] Could not find '%s' by '%s' in search: %v", title, artist, err)
							player.Next(0)
							if err != nil {
								log.Printf("[WARN] Could not skip track '%s' by '%s': %v", title, artist, err)
								continue
							}
							log.Printf("[INFO] Skipped '%s' by '%s' playing on '%s'\n", title, artist, player.Player.RoomName)
						} else {
							var complete = false
							for _, track := range res.Tracks.Tracks {
								log.Print("[DEBUG] Found tracks searching for alternates")
								if !track.Explicit {
									log.Print("[DEBUG] Found radio-friendly track")
									uri := fmt.Sprintf("x-sonos-spotify:spotify%%3atrack%%3a%s?%s", track.ID, parsedTrackURI[1])
									//var uri = "x-sonos-spotify:spotify%3atrack%3a" + trackID
									var meta = "<DIDL-Lite xmlns:dc=\"http://purl.org/dc/elements/1.1/\" xmlns:upnp=\"urn:schemas-upnp-org:metadata-1-0/upnp/\" xmlns:r=\"urn:schemas-rinconnetworks-com:metadata-1-0/\" xmlns=\"urn:schemas-upnp-org:metadata-1-0/DIDL-Lite/\"><item id=\"-1\" parentID=\"-1\" restricted=\"true\"><res protocolInfo=\"sonos.com-spotify:*:audio/x-spotify:*\" duration=\"" + posInfo.TrackDuration + "\">x-sonos-spotify:spotify%3atrack%3a" + trackID + "?sid=12&amp;flags=8224&amp;sn=8</res><r:streamContent></r:streamContent><r:radioShowMd></r:radioShowMd><upnp:albumArtURI>/getaa?s=1&amp;u=x-sonos-spotify%3aspotify%253atrack%253a" + trackID + "%3fsid%3d12%26flags%3d8224%26sn%3d8</upnp:albumArtURI><dc:title>" + title + "</dc:title><upnp:class>object.item.audioItem.musicTrack</upnp:class><dc:creator>" + artist +
										"</dc:creator><upnp:album>" + album + "</upnp:album><r:tags>1</r:tags></item></DIDL-Lite>"
									req := upnp.AddURIToQueueIn{
										EnqueuedURI:                     uri,
										EnqueuedURIMetaData:             meta,
										DesiredFirstTrackNumberEnqueued: uint32(currentTrack + 1),
									}
									log.Printf("[DEBUG] Queued: %s", meta)
									if _, err := player.AddURIToQueue(0 /*instanceId*/, &req); nil != err {
										log.Printf("[INFO] Couldn't add radio-friendly version '%s', '%s' of '%s' by '%s', skipped playing on '%s': %v\n",
											track.ID, track.URI, title, artist, player.Player.RoomName, err)
										err = player.Next(0)
										if err != nil {
											log.Printf("[WARN] Could not skip track '%s' by '%s': %v", title, artist, err)
											break
										}
										complete = true
									} else {
										err = player.Next(0)
										if err != nil {
											log.Printf("[WARN] Could not skip track '%s' by '%s': %v", title, artist, err)
											break
										}
										log.Printf("[INFO] Skipped to radio-friendly version '%s' of '%s' by '%s' playing on '%s'", uri, title, artist, player.Player.RoomName)
										complete = true
									}
									break
								}
							}

							if !complete {
								err = player.Next(0)
								if err != nil {
									log.Printf("[WARN] Could not skip track '%s' by '%s': %v", title, artist, err)
									continue
								}
								log.Printf("[INFO] Skipped '%s' by '%s' playing on '%s'\n", title, artist, player.Player.RoomName)
							}
						}
					}
				} else {
					log.Printf("[DEBUG] Not playing Spotify track: %s", uri)
				}
			case upnp.ContentDirectory_EventType:

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
	sonosDevices = sonos.ConnectAll(mgr, reactor, sonos.SVC_AV_TRANSPORT|sonos.SVC_CONTENT_DIRECTORY)
	log.Printf("[DEBUG] Set up event handler for Sonos, found %d devices", len(sonosDevices))
	for _, player := range sonosDevices {
		log.Printf("[DEBUG] Found %s", player.RoomName)
	}
}

func InitializeSonos() {
	mgr := ssdp.MakeManager()
	ifname := getLocalInterfaceName()
	log.Printf("[DEBUG] Discovering devices over %s...", ifname)
	if err := mgr.Discover(ifname, port, false); nil != err {
		panic(err)
	} else {
		SetupEvents(mgr)
	}
}
