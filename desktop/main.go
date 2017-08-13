package main

import (
	"log"
	"os"

	"github.com/atinm/spotify-filter/icon"
	"github.com/atinm/spotify-filter/lib"
	"github.com/getlantern/systray"
	"github.com/hashicorp/logutils"
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
	mExplicit := systray.AddMenuItem("Parental Control", "Parental Control")
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
