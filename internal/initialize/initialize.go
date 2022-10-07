package initialize

import (
	"flag"
	"log"
	"os"
	"reflect"
	"strconv"

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

/*
func setLogLevel(l int) {
	switch {
	case l <= 20:
		config.Log.SetMinLevel(logutils.LogLevel("ERROR"))
	case l > 20 && l <= 40:
		config.Log.SetMinLevel(logutils.LogLevel("WARNING"))
	case l > 40 && l <= 60:
		config.Log.SetMinLevel(logutils.LogLevel("INFO"))
	case l > 60 && l <= 80:
		config.Log.SetMinLevel(logutils.LogLevel("DEBUG"))
	case l > 80:
		config.Log.SetMinLevel(logutils.LogLevel("TRACE"))
	}
}
*/

func Init() {
	/*
		var (
			tz  string
			err error
		)
	*/

	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("[FATAL] Unable to initialize.")
	}

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

	if err = cfg.Validate(); err != nil {
		log.Fatalf("[FATAL] %v", err)
	}

	/*
		// log configuration
		flag.IntVar(&config.LogLevel,
			"log",
			getEnvInt("LOG_LEVEL", 50),
			"(LOG_LEVEL)\nlog level")
		// local webserver configuration
		flag.IntVar(&config.WebSrvPort,
			"http-port",
			getEnvInt("HTTP_PORT", 8080),
			"(HTTP_PORT)\nlisten port for internal webserver")
		flag.StringVar(&config.WebSrvIP,
			"http-ip",
			getEnvString("HTTP_IP", ""),
			"(HTTP_IP)\nlisten ip for internal webserver")
		flag.IntVar(&config.WebSrvReadTimeout,
			"http-read-timeout",
			getEnvInt("HTTP_READ_TIMEOUT", 5),
			"(HTTP_READ_TIMEOUT)\ninternal http server read timeout in seconds")
		flag.IntVar(&config.WebSrvWriteTimeout,
			"http-write-timeout",
			getEnvInt("HTTP_WRITE_TIMEOUT", 2),
			"(HTTP_WRITE_TIMEOUT\ninternal http server write timeout in seconds")
		flag.IntVar(&config.WebSrvIdleTimeout,
			"http-idle-timeout",
			getEnvInt("HTTP_IDLE_TIMEOUT", 2),
			"(HTTP_IDLE_TIMEOUT)\ninternal http server idle timeout in seconds")
		// timezone
		flag.StringVar(&tz,
			"timezone",
			getEnvString("TZ", "America/Chicago"),
			"(TZ)\ntimezone")
		// read command line options
		flag.Parse()

		// logging level
		setLogLevel(config.LogLevel)
		log.SetOutput(config.Log)

		// timezone configuration
		config.TimeZoneUTC, _ = time.LoadLocation("UTC")
		if config.TimeZoneLocal, err = time.LoadLocation(tz); err != nil {
			log.Fatalf("[ERROR] Unable to parse timezone string. Please use one of the timezone database values listed here: %s", "https://en.wikipedia.org/wiki/List_of_tz_database_time_zones")
		}

		// print current configuration
		log.Printf("[DEBUG] configuration value set: LOG_LEVEL           = %s\n", strconv.Itoa(config.LogLevel))
		log.Printf("[DEBUG] configuration value set: HTTP_PORT           = %s\n", strconv.Itoa(config.WebSrvPort))
		log.Printf("[DEBUG] configuration value set: HTTP_IP             = %s\n", config.WebSrvIP)
		log.Printf("[DEBUG] configuration value set: HTTP_READ_TIMEOUT   = %s\n", strconv.Itoa(config.WebSrvReadTimeout))
		log.Printf("[DEBUG] configuration value set: HTTP_WRITE_TIMEOUT  = %s\n", strconv.Itoa(config.WebSrvWriteTimeout))
		log.Printf("[DEBUG] configuration value set: HTTP_IDLE_TIMEOUT   = %s\n", strconv.Itoa(config.WebSrvIdleTimeout))
		log.Printf("[DEBUG] configuration value set: TZ                  = %s\n", tz)

		log.Println("[INFO] initialization complete")
	*/
}
