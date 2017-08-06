# spotify-filter
Skip explicit songs from playing on Spotify devices (Speakers including Sonos, Computers, Smartphones)

# Exclude Devices
Create a config.json file to exclude particular devices from
monitoring:

    {
        "ignore": ["Pixel", "Basement"]
    }

The names are from the Device list shown in the Spotify application
devices list.

# Toggling Filtering of Explicit Content

You can either just kill the program to stop filtering, or you have to
be on the same LAN as the program and go to
`http://localhost:5005/filter` in the browser (or curl).

# Building and Running

    go get github.com/atinm/spotify-filter
    go build
    ./spotify-filter

The program will pop up the browser to authenticate the user to allow
the program to read the user player state (read the songs playing, and
to skip to the next song). Accept and the program will continue. Leave
it running as long as you want it to keep filtering explicit content.

# Sonos only

Filtering Spotify explicit songs from playing on Sonos requires
running https://github.com/jishi/node-sonos-http-api locally on port
5005 (default). Follow the instructions on how to start it. You will
need to create a settings.json file for node-sonos-http-api with:

    {
        "webhook": "http://localhost:5007/sonos/updates"
    }

to get updates from Sonos into the spotify-filter program.

# TBD

Replacing explicit songs with Radio Edits (this is blocked on Spotify not exposing the play queue for editing). A workaround that copies a playlist, replaces explicit songs with radio edits if available can be done for when playing from playlists, but that isn't a complete solution.

