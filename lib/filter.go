package lib

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

func FiltersEnabled() bool {
	// we only have one rule, but this could be a check whether any are enabled
	return ParentalControlsEnabled()
}

func ParentalControlsEnabled() bool {
	return rule.Explicit
}

func Rules(track *spotify.FullTrack, device string) bool {
	// filter explicit
	if track != nil && rule.Explicit && track.Explicit && !ignored(device) {
		return true
	}
	return false
}

func SetParentalControls(b bool) {
	rule.Explicit = b
}
