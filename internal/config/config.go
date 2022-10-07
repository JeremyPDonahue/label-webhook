package config

import (
	"fmt"
	"time"

	"github.com/hashicorp/logutils"
)

type Config struct {
	// time configuration
	TimeFormat    string `env:"TIME_FORMAT" default:"2006-01-02 15:04:05"`
	TimeZoneLocal *time.Location
	TimeZoneUTC   *time.Location

	// logging
	LogLevel int `env:"LOG_LEVEL" default:"50"`
	Log      *logutils.LevelFilter

	// webserver
	WebServerPort         int    `env:"WEBSERVER_PORT" default:"8080"`
	WebServerIP           string `env:"WEBSERVER_IP" default:"0.0.0.0"`
	WebServerReadTimeout  int    `env:"WEBSERVER_READ_TIMEOUT" default:"5"`
	WebServerWriteTimeout int    `env:"WEBSERVER_WRITE_TIMEOUT" default:"1"`
	WebServerIdleTimeout  int    `env:"WEBSERVER_IDLE_TIMEOUT" default:"2"`
}

func NewConfig() (*Config, error) {
	cfg := &Config{}

	return cfg, nil
}

func (cfg *Config) Validate() error {
	checks := []struct {
		bad    bool
		errMsg string
	}{
		{cfg.TimeFormat == "", "no TimeFormat specified"},
		{cfg.LogLevel == 0, "no LogLevel specified"},
		{cfg.WebServerPort == 0, "no WebServer.Port specified"},
		{cfg.WebServerIP == "", "no WebServer.IP specified"},
		{cfg.WebServerReadTimeout == 0, "no WebServer.ReadTimeout specified"},
		{cfg.WebServerWriteTimeout == 0, "no WebServer.WriteTimeout specified"},
		{cfg.WebServerIdleTimeout == 0, "no WebServer.IdleTimeout specified"},
	}

	for _, check := range checks {
		if check.bad {
			return fmt.Errorf("invalid config: %s", check.errMsg)
		}
	}

	return nil
}
