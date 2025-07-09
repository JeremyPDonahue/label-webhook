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
	TimeZoneLocal string         `env:"time_zone" default:"UTC"`
	TZoneLocal    *time.Location `ignored:"true"`
	TZoneUTC      *time.Location `ignored:"true"`

	// config file
	ConfigFile string `env:"config_file" default:"/etc/webhook/config.yaml"`

	// logging
	LogLevel int                   `env:"log_level" default:"60"`
	Log      *logutils.LevelFilter `ignored:"true"`

	// webserver
	WebServerPort         int    `env:"webserver_port" default:"8443"`
	WebServerIP           string `env:"webserver_ip" default:"0.0.0.0"`
	WebServerCertificate  string `env:"webserver_cert"`
	WebServerKey          string `env:"webserver_key"`
	WebServerReadTimeout  int    `env:"webserver_read_timeout" default:"30"`
	WebServerWriteTimeout int    `env:"webserver_write_timeout" default:"30"`
	WebServerIdleTimeout  int    `env:"webserver_idle_timeout" default:"120"`

	// admission control configuration
	DryRun               bool     `env:"dry_run" default:"false"`
	EnableMetrics        bool     `env:"enable_metrics" default:"true"`
	MetricsPort          int      `env:"metrics_port" default:"9090"`
	AllowAdminNoMutate   bool     `env:"allow_admin_nomutate" default:"false"`
	ExcludedNamespaces   []string `ignored:"true"`

	// custom labeling configuration
	CustomLabels         map[string]string `ignored:"true"`
	LabelPrefix          string            `env:"label_prefix" default:"managed-by"`
	Organization         string            `env:"organization" default:"default"`
	Environment          string            `env:"environment" default:"production"`
	EnableLabeling       bool              `env:"enable_labeling" default:"true"`
	LabelAllWorkloads    bool              `env:"label_all_workloads" default:"true"`

	// certificate configuration
	CACert         string `env:"ca_cert"`
	CAPrivateKey   string `env:"ca_private_key"`
	CertCert       string `env:"cert_cert"`
	CertPrivateKey string `env:"cert_private_key"`

	// kubernetes configuration
	NameSpace       string `env:"namespace" default:"openshift-webhook"`
	ServiceName     string `env:"service_name" default:"custom-labels-webhook"`
	ClusterName     string `env:"cluster_name" default:"openshift-cluster"`
	WebhookName     string `env:"webhook_name" default:"custom-labels-mutator"`
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
