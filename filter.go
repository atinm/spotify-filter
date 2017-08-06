package main

import (
	"github.com/zmb3/spotify"
)

func ignored(device string) bool {
	for _, name := range config.Ignored {
		if name == device {
			return true
		}
	}
	return false
}

func Rules(track *spotify.FullTrack, device string) bool {
	// filter explicit
	if track != nil && rule.Explicit && track.Explicit && !ignored(device) {
		return true
	}
	return false
}
