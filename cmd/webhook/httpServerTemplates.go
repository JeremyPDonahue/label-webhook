package main

import (
	"log"

	"encoding/json"
	"net/http"
	"net/url"
)

const cT string = "Content-Type"
const cTjson string = "application/json"
const marshalErrorMsg string = "[TRACE] Unable to marshal error message: %v."

func tmpltError(w http.ResponseWriter, s int, m string) {
	var (
		output []byte
		o      = struct {
			Error    int    `json:"error" yaml:"error"`
			ErrorMsg string `json:"errorMessage" yaml:"errorMessage"`
		}{
			Error:    s,
			ErrorMsg: m,
		}
		err error
	)

	w.Header().Add(cT, cTjson)

	output, err = json.MarshalIndent(o, "", "  ")
	if err != nil {
		log.Printf(marshalErrorMsg, err)
	}
	w.WriteHeader(s)
	w.Write(output) //nolint:errcheck
}

func tmpltHealthCheck(w http.ResponseWriter) {
	o := struct {
		WebServer bool   `json:"webServerActive" yaml:"webServerActive"`
		Status    string `json:"status" yaml:"status"`
	}{
		WebServer: true,
		Status:    "healthy",
	}

	output, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		log.Printf(marshalErrorMsg, err)
	}

	w.Header().Add(cT, cTjson)
	w.Write(output) //nolint:errcheck
}

func tmpltWebRoot(w http.ResponseWriter) {
	o := struct {
		Application string `json:"application" yaml:"application"`
		Description string `json:"description" yaml:"description"`
		Version     string `json:"version" yaml:"version"`
	}{
		Application: "AppID Labeling Webhook API",
		Description: "Mutating Webhook for AppID Label Application",
		Version:     "v1.0.0",
	}
	w.Header().Add(cT, cTjson)

	output, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		log.Printf(marshalErrorMsg, err)
	}
	w.Write(output) //nolint:errcheck
}

func tmpltAdminToggle(w http.ResponseWriter, urlPrams url.Values) {
	o := struct {
		Application string `json:"application" yaml:"application"`
		Description string `json:"description" yaml:"description"`
		Version     string `json:"version" yaml:"version"`
		AdminNoMutate bool `json:"admin-no-mutate" yaml:"admin-no-mutate"`
	}{
		Application: "AppID Labeling Webhook API",
		Description: "Mutating Webhook for AppID Label Application",
		Version:     "v1.0.0",
	}
	w.Header().Add(cT, cTjson)

	// Admin toggle functionality removed for simplified webhook
	o.AdminNoMutate = cfg.AllowAdminNoMutate

	output, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		log.Printf(marshalErrorMsg, err)
	}
	w.Write(output) //nolint:errcheck
}
