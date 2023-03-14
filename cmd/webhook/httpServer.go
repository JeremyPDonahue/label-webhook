package main

import (
	"crypto/tls"
	"log"
	"mutating-webhook/internal/certificate"
	"mutating-webhook/internal/config"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const InvalidMethod string = "Invalid http method."

func httpAccessLog(req *http.Request) {
	log.Printf("[TRACE] %s - %s - %s\n", req.Method, req.RemoteAddr, req.RequestURI)
}

func crossSiteOrigin(w http.ResponseWriter) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Add("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization, X-API-Token")
}

func strictTransport(w http.ResponseWriter) {
	w.Header().Add("Strict-Transport-Security", "max-age=63072000")
}

func httpServer(cfg *config.Config) {
	var serverCertificate tls.Certificate
	if config.DefaultConfig().WebServerCertificate == "" || cfg.WebServerKey == "" {
		log.Printf("[INFO] No webserver certificate configured, automatically generating self signed certificate.")
		serverCertificate = certificate.CreateServerCert()
	} else {
		log.Fatal("[FATAL] Code to support external webserver certificate is not complete yet. ./cmd/webhook/httpServer.go:36")
		// read certificate from files
		// check for errors
	}
	path := http.NewServeMux()

	connection := &http.Server{
		Addr:         cfg.WebServerIP + ":" + strconv.FormatInt(int64(cfg.WebServerPort), 10),
		Handler:      path,
		ReadTimeout:  time.Duration(cfg.WebServerReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.WebServerWriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.WebServerIdleTimeout) * time.Second,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			},
			Certificates: []tls.Certificate{
				serverCertificate,
			},
		},
	}

	// healthcheck
	path.HandleFunc("/healthcheck", webHealthCheck)
	// api-endpoint
	path.HandleFunc("/api/v1/mutate", webMutatePod)
	// web root
	path.HandleFunc("/", webRoot)

	if err := connection.ListenAndServeTLS("", ""); err != nil {
		log.Fatalf("[ERROR] %s\n", err)
	}
}

func webRoot(w http.ResponseWriter, r *http.Request) {
	httpAccessLog(r)
	crossSiteOrigin(w)
	strictTransport(w)

	switch {
	case strings.ToLower(r.Method) != "get":
		log.Printf("[DEBUG] Request to '/' was made using the wrong method: expected %s, got %s", "GET", strings.ToUpper(r.Method))
		tmpltError(w, http.StatusBadRequest, InvalidMethod)
	case r.URL.Path != "/":
		log.Printf("[DEBUG] Unable to locate requested path: '%s'", r.URL.Path)
		tmpltError(w, http.StatusNotFound, "Requested path not found.")
	default:
		tmpltWebRoot(w)
	}
}

func webHealthCheck(w http.ResponseWriter, r *http.Request) {
	httpAccessLog(r)
	crossSiteOrigin(w)
	strictTransport(w)

	if strings.ToLower(r.Method) == "get" {
		tmpltHealthCheck(w)
	} else {
		log.Printf("[DEBUG] Request to '/healthcheck' was made using the wrong method: expected %s, got %s", "GET", strings.ToUpper(r.Method))
		tmpltError(w, http.StatusBadRequest, InvalidMethod)
	}
}
