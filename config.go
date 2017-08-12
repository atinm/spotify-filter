package main

import (
	"github.com/hashicorp/logutils"
)

type Config struct {
	Ignored      []string          `json:"ignored"`
	ClientId     string            `json:"client_id"`
	ClientSecret string            `json:"client_secret"`
	RedirectURI  string            `json:"redirect_uri"`
	LogLevel     logutils.LogLevel `json:"log_level"`
}
