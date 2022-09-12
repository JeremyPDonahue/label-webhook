package main

import (
	"os"
	"time"

	"github.com/hashicorp/logutils"
)

type configStructure struct {
	// time configuration
	TimeFormat    string
	TimeZoneLocal *time.Location
	TimeZoneUTC   *time.Location

	// logging
	LogLevel int
	Log      *logutils.LevelFilter

	// webserver
	WebSrvPort         int
	WebSrvIP           string
	WebSrvReadTimeout  int
	WebSrvWriteTimeout int
	WebSrvIdleTimeout  int
}

type patchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

// Set Defaults
var config = configStructure{
	TimeFormat: "2006-01-02 15:04:05",
	Log: &logutils.LevelFilter{
		Levels: []logutils.LogLevel{"TRACE", "DEBUG", "INFO", "WARNING", "ERROR"},
		Writer: os.Stderr,
	},
	WebSrvReadTimeout:  5,
	WebSrvWriteTimeout: 10,
	WebSrvIdleTimeout:  2,
}
