package main

import (
	"flag"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/hashicorp/logutils"
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

func initialize() {
	var (
		tz  string
		err error
	)

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
}
