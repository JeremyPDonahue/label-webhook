package config

import (
	"flag"
	"log"
	"os"
	"reflect"
	"strings"
	"time"
)

func getOSEnv(env, def string) string {
	if val, ok := os.LookupEnv(strings.ToUpper(env)); ok {
		return val
	}
	return def
}

// Init initializes the application configuration by reading default values from the struct's tags
// and environment variables. Tags processed by this process are as follows:
// `ignored:"true" env:"ENVIRONMENT_VARIABLE" default:"default value"`
func Init() *Config {
	cfg := DefaultConfig()

	cfgInfo, err := getStructInfo(cfg)
	if err != nil {
		log.Fatalf("[FATAL] %v", err)
	}

	for _, info := range cfgInfo {
		switch info.Type.String() {
		case "string":
			var dv string
			if info.DefaultValue != nil {
				dv = info.DefaultValue.(string)
			}
			p := reflect.ValueOf(cfg).Elem().FieldByName(info.Name).Addr().Interface().(*string)
			flag.StringVar(p, info.Name, dv, "("+info.Key+")")

		case "bool":
			var dv bool
			if info.DefaultValue != nil {
				dv = info.DefaultValue.(bool)
			}
			p := reflect.ValueOf(cfg).Elem().FieldByName(info.Name).Addr().Interface().(*bool)
			flag.BoolVar(p, info.Name, dv, "("+info.Key+")")

		case "int":
			var dv int
			if info.DefaultValue != nil {
				dv = int(info.DefaultValue.(int64))
			}
			p := reflect.ValueOf(cfg).Elem().FieldByName(info.Name).Addr().Interface().(*int)
			flag.IntVar(p, info.Name, dv, "("+info.Key+")")
		}
	}
	flag.Parse()

	// set logging level
	cfg.setLogLevel()

	// timezone & format configuration
	cfg.TZoneUTC, _ = time.LoadLocation("UTC")
	if err != nil {
		log.Fatalf("[ERROR] Unable to parse timezone string. Please use one of the timezone database values listed here: %s", "https://en.wikipedia.org/wiki/List_of_tz_database_time_zones")
	}
	cfg.TZoneLocal, err = time.LoadLocation(cfg.TimeZoneLocal)
	if err != nil {
		log.Fatalf("[ERROR] Unable to parse timezone string. Please use one of the timezone database values listed here: %s", "https://en.wikipedia.org/wiki/List_of_tz_database_time_zones")
	}
	time.Now().Format(cfg.TimeFormat)

	// print running config
	cfg.printRunningConfig(cfgInfo)

	// read config file
	configFileData, err := getConfigFileData(cfg.ConfigFile)
	if err != nil {
		log.Fatalf("[FATAL] Unable to read configuration file")
	}
	if cfg.AllowAdminNoMutate == false {
		cfg.AllowAdminNoMutate = configFileData.AllowAdminNoMutate
	}
	if cfg.AllowAdminNoMutateToggle == "2d77b689-dc14-40a5-8971-34c62999335c" {
		cfg.AllowAdminNoMutateToggle = configFileData.AllowAdminNoMutateToggle
	}
	if cfg.DockerhubRegistry == "registry.hub.docker.com" {
		cfg.DockerhubRegistry = configFileData.DockerhubRegistry
	}
	cfg.MutateIgnoredImages = configFileData.MutateIgnoredImages

	log.Println("[INFO] initialization sequence complete")
	return cfg
}
