package main

import (
	"log"
	"os"

	"github.com/atinm/spotify-filter/icon"
	"github.com/atinm/spotify-filter/lib"
	"github.com/getlantern/systray"
	"github.com/hashicorp/logutils"
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
	lib.LogFilter = &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG", "INFO", "WARN", "ERROR"},
		MinLevel: logutils.LogLevel("WARN"),
		Writer:   os.Stderr,
	}
	log.SetOutput(lib.LogFilter)

	lib.LoadConfig("config.json")

	// authenticate against spotify
	lib.Authenticate()
	systray.Run(onReady)
}

func onReady() {
	//systray.SetIcon(icon.Enable)
	//systray.SetTitle("Kid Friendly Spotify")
	systray.SetTooltip("Kid Friendly Spotify")
	mExplicit := systray.AddMenuItem("Parental Controls", "Parental Controls")
	mAbout := systray.AddMenuItem("About", "About")
	mQuit := systray.AddMenuItem("Quit", "Quit the whole app")

	// We can manipulate the systray in other goroutines
	go func() {
		//systray.SetTitle("Kid Friendly Spotify")
		systray.SetTooltip("Kid Friendly Spotify")
		if lib.ParentalControlsEnabled() {
			mExplicit.Check()
			log.Print("[DEBUG] parental controls are enabled")
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
						go lib.Monitor()
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

	// start the server to listen for authentication, never returns
	lib.Server()
}
