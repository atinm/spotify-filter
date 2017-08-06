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

Start by registering your application at the following page:

    https://developer.spotify.com/my-applications/

Set the Redirect URI to be `http://localhost:5007/callback` (the port
that the program listens on) and Save. You'll get a client ID and
secret key for your application.

Export the `SPOTIFY_ID` and `SPOTIFY_SECRET` environment variables set
to the client id and secret you created above at application
registration make them available to the program.

    go get github.com/atinm/spotify-filter
    go build
    export SPOTIFY_ID=<the client id from the Spotify application registration>
    export SPOTIFY_SECRET=<the client secret from the Spotify application registration>
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

