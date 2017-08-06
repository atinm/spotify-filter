package main

import (
	"github.com/hashicorp/logutils"
)

type Update struct {
	Type      string   `json:"type"`
}

type Equalizer struct {
	Bass int `json:"bass"`
	Treble int `json:"treble"`
	SpeechEnhancement bool `json:"speechEnhancement"`
	NightMode bool `json:"nightMode"`
	Loudness bool `json:"loudness"`
}

type Track struct {
	Artist string `json:"artist"`
	Title string `json:"title"`
	Album string `json:"album"`
	AlbumArtUri string `json:"albumArtUri"`
	Duration int `json:"duration"`
	Uri string `json:"uri"`
	Type string `json:"type"`
	StationName string `json"stationName"`
	AbsoluteAlbumArtUri string `json:"absoluteAlbumArtUri"`
}

type PlayMode struct {
	Repeat string `json:"repeat"`
	Shuffle bool `json:"shuffle"`
	CrossFade bool `json:"crossfade"`
}

type State struct {
	CurrentTrack Track `json:"currentTrack"`
	NextTrack Track `json:"nextTrack"`
	PlayMode PlayMode `json:"playMode"`
	PlaylistName string `json:"playlistName"`
	RelTime int `json:"relTime"`
	StateTime int `json:"stateTime"`
	Volume int `json:"volume"`
	Mute bool `json:"mute"`
	TrackNo int `json:"trackNo"`
	PlaybackState string `json:"playbackState"`
	Equalizer Equalizer `json:"equalizer"`
	ElapsedTime int `json:"elapsedTime"`
	ElapsedTimeFormatted string `json:"elapsedTimeFormatted"`
}

type GroupState struct {
	Volume int `json:"volume"`
	Mute bool `json:"mute"`
}

type Player struct {
	Uuid string `json:"uuid"`
	Coordinator string `json:"coordinator"`
	RoomName string `json:"roomName"`
	State State `json:"state"`
	GroupState GroupState `json:"groupState"`
	AvTransportUri string `json:"avTransportUri"`
	AvTransportUriMetadata string `json:"avTransportUriMetadata"`
}

type TransportState struct {
	Player Player `json:"data"`
}

type VolumeChange struct {
	Uuid string `json:"uuid"`
	PreviousVolume int `json:"previousVolume"`
	NewVolume int `json:"newVolume"`
	RoomName string `json:"roomName"`
}

type MuteChange struct {
	Uuid string `json:"uuid"`
	PreviousMute int `json:"previousMute"`
	NewMute int `json:"newMute"`
	RoomName string `json:"roomName"`
}

type Rule struct {
	Explicit bool
}

type Config struct {
	Ignored []string `json:"ignored"`
	LogLevel logutils.LogLevel `json:"log_level"`
}
