package config

import (
	"log"
	"os"
	"reflect"
	"strconv"
	"time"

	"github.com/hashicorp/logutils"
)

type Config struct {
	// time configuration
	TimeFormat    string         `env:"time_format" default:"2006-01-02 15:04:05"`
	TimeZoneLocal string         `env:"time_zone" default:"America/Chicago"`
	TZoneLocal    *time.Location `ignored:"true"`
	TZoneUTC      *time.Location `ignored:"true"`

	// config file
	ConfigFile string `env:"config_file" default:"./config.yaml"`

	// logging
	LogLevel int                   `env:"log_level" default:"50"`
	Log      *logutils.LevelFilter `ignored:"true"`

	// webserver
	WebServerPort         int    `env:"webserver_port" default:"8443"`
	WebServerIP           string `env:"webserver_ip" default:"0.0.0.0"`
	WebServerCertificate  string `env:"webserver_cert"`
	WebServerKey          string `env:"webserver_key"`
	WebServerReadTimeout  int    `env:"webserver_read_timeout" default:"5"`
	WebServerWriteTimeout int    `env:"webserver_write_timeout" default:"1"`
	WebServerIdleTimeout  int    `env:"webserver_idle_timeout" default:"2"`

	// mutation configuration
	AllowAdminNoMutate       bool     `env:"allow_admin_nomutate" default:"false"`
	AllowAdminNoMutateToggle string   `env:"allow_admin_nomutate_toggle" default:"7b068a99-c02b-410a-bd59-3514bac85e7a"`
	DockerhubRegistry        string   `env:"dockerhub_registry" default:"registry.hub.docker.com"`
	MutateIgnoredImages      []string `ignored:"true"`

	// certificate configuration
	CACert         string `env:"ca_cert"`
	CAPrivateKey   string `env:"ca_private_key"`
	CertCert       string `env:"cert_cert"`
	CertPrivateKey string `env:"cert_private_key"`
}

// DefaultConfig initializes the config variable for use with a prepared set of defaults.
func DefaultConfig() Config {
	return Config{
		Log: &logutils.LevelFilter{
			Levels: []logutils.LogLevel{"TRACE", "DEBUG", "INFO", "WARNING", "ERROR"},
			Writer: os.Stderr,
		},
	}
}

func setLogLevel(cfg Config) {
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

func printRunningConfig(cfg *Config, cfgInfo []StructInfo) {
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
