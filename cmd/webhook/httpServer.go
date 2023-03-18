package main

import (
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	"crypto/tls"
	"encoding/json"
	"net/http"

	"mutating-webhook/internal/certificate"
	"mutating-webhook/internal/config"
	"mutating-webhook/internal/operations"

	admission "k8s.io/api/admission/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
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
			MinVersion: tls.VersionTLS13,
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

	ah := &admissionHandler{
		decoder: serializer.NewCodecFactory(runtime.NewScheme()).UniversalDeserializer(),
		config:  cfg,
	}

	// healthcheck
	path.HandleFunc("/healthcheck", webHealthCheck)
	// pod admission
	path.HandleFunc("/api/v1/admit/pod", ah.Serve(operations.PodsValidation()))
	// deployment admission
	path.HandleFunc("/api/v1/admit/deployment", ah.Serve(operations.DeploymentsValidation()))
	// pod mutation
	path.HandleFunc("/api/v1/mutate/pod", ah.Serve(operations.PodsMutation()))
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

type admissionHandler struct {
	decoder runtime.Decoder
	config  *config.Config
}

// Serve returns a http.HandlerFunc for an admission webhook
func (h *admissionHandler) Serve(hook operations.Hook) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		httpAccessLog(r)
		crossSiteOrigin(w)
		strictTransport(w)

		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			msg := fmt.Sprintf("incorrect method: got request type %s, expected request type %s", r.Method, http.MethodPost)
			log.Printf("[TRACE] %s", msg)
			tmpltError(w, http.StatusMethodNotAllowed, msg)
			return
		}

		if contentType := r.Header.Get("Content-Type"); contentType != "application/json" {
			msg := "only content type 'application/json' is supported"
			log.Printf("[TRACE] %s", msg)
			tmpltError(w, http.StatusBadRequest, msg)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			msg := fmt.Sprintf("could not read request body: %v", err)
			log.Printf("[TRACE] %s", msg)
			tmpltError(w, http.StatusBadRequest, msg)
			return
		}

		var review admission.AdmissionReview
		if _, _, err := h.decoder.Decode(body, nil, &review); err != nil {
			msg := fmt.Sprintf("could not deserialize request: %v", err)
			log.Printf("[TRACE] %s", msg)
			tmpltError(w, http.StatusBadRequest, msg)
			return
		}

		if review.Request == nil {
			msg := "malformed admission review: request is nil"
			log.Printf("[TRACE] %s", msg)
			tmpltError(w, http.StatusBadRequest, msg)
			return
		}

		result, err := hook.Execute(review.Request, h.config)
		if err != nil {
			msg := err.Error()
			log.Printf("[ERROR] Internal Server Error: %s", msg)
			tmpltError(w, http.StatusInternalServerError, msg)
			return
		}

		admissionResponse := admission.AdmissionReview{
			Response: &admission.AdmissionResponse{
				UID:     review.Request.UID,
				Allowed: result.Allowed,
				Result:  &meta.Status{Message: result.Msg},
			},
		}

		// set the patch operations for mutating admission
		if len(result.PatchOps) > 0 {
			patchBytes, err := json.Marshal(result.PatchOps)
			if err != nil {
				msg := fmt.Sprintf("could not marshal JSON patch: %v", err)
				log.Printf("[ERROR] %s", msg)
				tmpltError(w, http.StatusInternalServerError, msg)
			}
			admissionResponse.Response.Patch = patchBytes
		}

		res, err := json.Marshal(admissionResponse)
		if err != nil {
			msg := fmt.Sprintf("could not marshal response: %v", err)
			log.Printf("[ERROR] %s", msg)
			tmpltError(w, http.StatusInternalServerError, msg)
			return
		}

		log.Printf("[INFO] Webhook [%s] - Allowed: %t", review.Request.Operation, result.Allowed)
		w.WriteHeader(http.StatusOK)
		w.Write(res)
	}
}
