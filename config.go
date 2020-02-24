package main

import (
	"encoding/json"
	"log"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	defaultServerInfoPath = "ocs/v2.php/apps/serverinfo/api/v1/info"
)

type internalConfig struct {
	TimeoutSeconds          float64 `json:"timeout"`
	NextcloudURL            string  `json:"nextcloud_url"`
	AppendDefaultServerInfo bool    `json:"append_default_serverinfo_path"`
	Listen                  string  `json:"listen"`
}

type Config struct {
	Timeout time.Duration
	InfoURL *url.URL
	Listen  string
}

func NewConfig(filename string) *Config {
	fd, err := os.Open(filename)
	if err != nil {
		log.Fatalln(err)
	}
	defer fd.Close()

	config := &internalConfig{
		TimeoutSeconds:          10,
		Listen:                  ":9091",
		AppendDefaultServerInfo: true,
	}
	if err := json.NewDecoder(fd).Decode(config); err != nil {
		log.Fatalln(err)
	}

	infoURL, err := url.Parse(config.NextcloudURL)
	if err != nil {
		log.Fatalf("Unable to parse info_url: %s", err)
	}
	if config.AppendDefaultServerInfo {
		if !strings.HasSuffix(infoURL.Path, "/") {
			infoURL.Path += "/"
		}
		infoURL.Path += defaultServerInfoPath
	}

	return &Config{
		Timeout: time.Duration(config.TimeoutSeconds * float64(time.Second)),
		InfoURL: infoURL,
		Listen:  config.Listen,
	}
}
