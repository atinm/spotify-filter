package lib

import (
	"encoding/json"
	"log"
	"os"

	"github.com/hashicorp/logutils"
)

type Config struct {
	Ignored         []string          `json:"ignored"`
	ClientId        string            `json:"client_id"`
	ClientSecret    string            `json:"client_secret"`
	RedirectURI     string            `json:"redirect_uri"`
	AuthTokenURI    string            `json:"auth_token_uri"`
	CertificateFile string            `json:"cert"`
	KeyFile         string            `json:"key"`
	Port            string            `json:"port"`
	LogLevel        logutils.LogLevel `json:"log_level"`
}

func LoadConfig(file string) {
	LogFilter = &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG", "INFO", "WARN", "ERROR"},
		MinLevel: logutils.LogLevel("WARN"),
		Writer:   os.Stderr,
	}
	log.SetOutput(LogFilter)

	conf, err := os.Open(file)
	if err != nil {
		log.Print("[DEBUG] No config file specified, ignoring.")
	} else {
		defer conf.Close()

		decoder := json.NewDecoder(conf)
		err = decoder.Decode(&config)
		if err != nil {
			log.Fatalf("Config file 'config.json could not be read, %v", err)
		}
		if config.LogLevel != "" {
			LogFilter.SetMinLevel(config.LogLevel)
		}
		if config.RedirectURI != "" {
			redirectURI = config.RedirectURI
		}
		if config.AuthTokenURI != "" {
			authTokenURL = config.AuthTokenURI
		}
		if config.CertificateFile != "" {
			certificate = config.CertificateFile
		}
		if config.KeyFile != "" {
			key = config.KeyFile
		}
		if config.Port != "" {
			port = config.Port
		}
	}
}
