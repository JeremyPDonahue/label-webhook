package initialize

import (
	"flag"
	"log"
	"os"
	"reflect"
	"strconv"
	"time"

	"mutating-webhook/internal/config"
	"mutating-webhook/internal/envconfig"
)

// getEnvString returns string from environment variable
func getEnvString(env, def string) (val string) { //nolint:deadcode
	val = os.Getenv(env)

	if val == "" {
		return def
	}

	return
}

// getEnvInt returns int from environment variable
func getEnvInt(env string, def int) (ret int) {
	val := os.Getenv(env)

	if val == "" {
		return def
	}

	ret, err := strconv.Atoi(val)
	if err != nil {
		log.Fatalf("[ERROR] Environment variable is not numeric: %v\n", env)
	}

	return
}

// getEnvBool returns boolean from environment variable
func getEnvBool(env string, def bool) bool {
	var (
		err    error
		retVal bool
		val    = os.Getenv(env)
	)

	if len(val) == 0 {
		return def
	} else {
		retVal, err = strconv.ParseBool(val)
		if err != nil {
			log.Fatalf("[ERROR] Environment variable is not boolean: %v\n", env)
		}
	}

	return retVal
}

func Init() *config.Config {
	cfg := config.DefaultConfig()

	cfgInfo, err := envconfig.GetStructInfo(cfg)
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
			flag.StringVar(p, info.Name, getEnvString(info.Name, dv), "("+info.Key+")")
		case "bool":
			var dv bool

			if info.DefaultValue != nil {
				dv = info.DefaultValue.(bool)
			}
			p := reflect.ValueOf(cfg).Elem().FieldByName(info.Name).Addr().Interface().(*bool)
			flag.BoolVar(p, info.Name, getEnvBool(info.Name, dv), "("+info.Key+")")
		case "int":
			var dv int

			if info.DefaultValue != nil {
				dv = int(info.DefaultValue.(int64))
			}
			p := reflect.ValueOf(cfg).Elem().FieldByName(info.Name).Addr().Interface().(*int)
			flag.IntVar(p, info.Name, getEnvInt(info.Name, dv), "("+info.Key+")")
		}
	}
	flag.Parse()

	// set logging level
	cfg.SetLogLevel()

	// 
	if err = cfg.Validate(); err != nil {
		log.Fatalf("[FATAL] %v", err)
	}

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
	cfg.PrintRunningConfig(cfgInfo)

	log.Println("[INFO] initialization complete")
	return cfg
}
