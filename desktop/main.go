package main

import (
	"log"
	"os"

	"github.com/atinm/spotify-filter/icon"
	"github.com/atinm/spotify-filter/lib"
	"github.com/getlantern/systray"
	"github.com/skratchdot/open-golang/open"
)

func updateIcon() {
	if lib.FiltersEnabled() {
		systray.SetIcon(icon.Enable)
	} else {
		systray.SetIcon(icon.Disable)
	}
}

func main() {
	lib.LoadConfig("config.json")
	// authenticate against spotify
	lib.Authenticate()
	systray.Run(onReady)
}

func onReady() {
	systray.SetTooltip("Kid Friendly Spotify")
	mExplicit := systray.AddMenuItem("Parental Controls", "Parental Controls")
	mAbout := systray.AddMenuItem("About", "About")
	mQuit := systray.AddMenuItem("Quit", "Quit the whole app")

	// We can manipulate the systray in other goroutines
	go func() {
		systray.SetTooltip("Kid Friendly Spotify")
		if lib.ParentalControlsEnabled() {
			mExplicit.Check()
			log.Print("[DEBUG] parental controls are enabled")
			// start the server to listen for authentication
			lib.StartServer()
		} else {
			mExplicit.Uncheck()
			log.Print("[DEBUG] parental controls are disabled")
		}
		updateIcon()

		for {
			select {
			case <-mExplicit.ClickedCh:
				if mExplicit.Checked() {
					mExplicit.Uncheck()
					lib.SetParentalControls(false)
					log.Print("[DEBUG] Disabled parental controls")
				} else {
					mExplicit.Check()
					lib.SetParentalControls(true)
					log.Print("[DEBUG] Enabled parental controls")
					if lib.FiltersEnabled() {
						lib.StartMonitor()
					}
				}
				updateIcon()
			case <-mAbout.ClickedCh:
				open.Run("https://github.com/atinm/spotify-filter/blob/master/README.md")
			}
		}
	}()

	go func() {
		<-mQuit.ClickedCh
		systray.Quit()
		log.Print("[DEBUG] Quitting")
		os.Exit(0)
	}()
}
