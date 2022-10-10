package config

import (
	"fmt"
	"log"
	"mutating-webhook/internal/envconfig"
	"os"
	"reflect"
	"strconv"
	"time"

	"github.com/hashicorp/logutils"
)

type Config struct {
	// time configuration
	TimeFormat    string `env:"TIME_FORMAT" default:"2006-01-02 15:04:05"`
	TimeZoneLocal string `env:"TIME_ZONE" default:"America/Chicago"`
	TZoneLocal    *time.Location
	TZoneUTC      *time.Location

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

func DefaultConfig() *Config {
	return &Config{
		Log: &logutils.LevelFilter{
			Levels: []logutils.LogLevel{"TRACE", "DEBUG", "INFO", "WARNING", "ERROR"},
			Writer: os.Stderr,
		},
	}
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

func (cfg *Config) SetLogLevel() {
	switch {
	case cfg.LogLevel <= 20:
		cfg.Log.SetMinLevel(logutils.LogLevel("ERROR"))
	case cfg.LogLevel > 20 && cfg.LogLevel <= 40:
		cfg.Log.SetMinLevel(logutils.LogLevel("WARNING"))
	case cfg.LogLevel > 40 && cfg.LogLevel <= 60:
		cfg.Log.SetMinLevel(logutils.LogLevel("INFO"))
	case cfg.LogLevel > 60 && cfg.LogLevel <= 80:
		cfg.Log.SetMinLevel(logutils.LogLevel("DEBUG"))
	case cfg.LogLevel > 80:
		cfg.Log.SetMinLevel(logutils.LogLevel("TRACE"))
	}
	log.SetOutput(cfg.Log)
}

func (cfg *Config) PrintRunningConfig(cfgInfo []envconfig.StructInfo) {
	log.Printf("[DEBUG] Current Running Configuration Values:")
	for _, info := range cfgInfo {
		switch info.Type.String() {
		case "string":
			p := reflect.ValueOf(cfg).Elem().FieldByName(info.Name).Addr().Interface().(*string)
			log.Printf("[DEBUG]\t%s\t\t= %s\n", info.Alt, *p)
		case "bool":
			p := reflect.ValueOf(cfg).Elem().FieldByName(info.Name).Addr().Interface().(*bool)
			log.Printf("[DEBUG]\t%s\t\t= %s\n", info.Alt, strconv.FormatBool(*p))
		case "int":
			p := reflect.ValueOf(cfg).Elem().FieldByName(info.Name).Addr().Interface().(*int)
			log.Printf("[DEBUG]\t%s\t\t= %s\n", info.Alt, strconv.FormatInt(int64(*p), 10))
		}
	}
}
