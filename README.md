# spotify-filter
Skip explicit songs from playing on Spotify devices (Speakers including Sonos, Computers, Smartphones)

# Building and Running

Start by registering your application at [Spotify Application Registration](https://developer.spotify.com/my-applications/).

Set the Redirect URI to be `http://localhost:5009/callback` which is the port
where you will run the authorization server [`https://github.com/atinm/spotify-auth-server`](https://github.com/atinm/spotify-auth-server)
on, and Save. You'll get a client ID and client secret key for your application.

Note: You need to run the authorization server [`https://github.com/atinm/spotify-auth-server`](https://github.com/atinm/spotify-auth-server) before
you run this as the spotify-filter uses the authorization server to get the Spotify access_token and
refresh_token from the Spotify authorization server.

Export the `SPOTIFY_ID` environment variable set to the client id you created above at application
registration to make them available to the program (or you can use the configuration file `config.json`
to save it as described under the Configuration section).

The program listens using HTTPS for the authentication token from the authorization server, and therefore you
need to provide the cert and key pem files. You can generate these from the [generate_cert program in crypto/tls](https://golang.org/src/crypto/tls/generate_cert.go).

    go get github.com/atinm/spotify-filter
    cd desktop
    go build -o spotify-filter
    export SPOTIFY_ID=<the client id from the Spotify application registration>
    ./spotify-filter

The program will pop up the browser to authenticate the user to allow
the program to read the user player state (read the songs playing, and
to skip to the next song). Accept and the program will continue. Leave
it running as long as you want it to keep filtering explicit content.

# Configuration

You may have a config.json file in the same directory as the program:

    {
        "client_id": "<the client id from the Spotify application registration>,
        "ignore": ["Pixel", "Basement"],
        "log_level": "DEBUG|INFO|WARN(default)|ERROR",
        "cert": "cert.pem",
        "key": "key.pem"
    }

# Exclude Devices

To ignore certain devices, add the names as seen in the Spotify
application devices list to the config.json file:

    {
        "ignore": ["Pixel", "Basement"]
    }

# Desktop Tray Application

The desktop version of the program starts a tray application that allows you to
toggle Parental Controls and Quit. If you are able to access to server over the
the network (e.g. if you are on the same LAN), go to `http://localhost:5005/filter`
in the browser (or curl) to toggle the parental controls.

# TBD

Replacing explicit songs with Radio Edits (this is blocked on Spotify
not exposing the play queue for editing) in the Sonos queue.

A workaround for the Spotify app itself that copies a playlist, replaces
explicit songs with radio edits if available can be done for when playing
from playlists, but that isn't a complete solution. That has to wait for
Spotify to give access to its queue.

# Support

None. I subscribed to Google Play Music. It provides this
functionality already, has better search capabilities (Podcasts
search anyone?) and allows me to upload music to the cloud to listen
on any supported device including Sonos.
