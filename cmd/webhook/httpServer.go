package main

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"crypto/tls"
	"encoding/json"

	"mutating-webhook/internal/config"
	"mutating-webhook/internal/metrics"
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
	// Parse and validate certificate
	serverCertificate, err := tls.X509KeyPair(append([]byte(cfg.CertCert), []byte(cfg.CACert)...), []byte(cfg.CertPrivateKey))
	if err != nil {
		log.Fatalf("[ERROR] Failed to load server certificate: %v", err)
	}

	// Extract certificate expiry for metrics
	setCertificateExpiryMetrics(cfg.CertCert)

	// Setup webhook server
	webhookMux := http.NewServeMux()
	ah := &admissionHandler{
		decoder: serializer.NewCodecFactory(runtime.NewScheme()).UniversalDeserializer(),
		config:  cfg,
	}

	// Webhook endpoints
	webhookMux.HandleFunc("/api/v1/admit/pod", ah.ahServe(operations.PodsValidation()))
	webhookMux.HandleFunc("/api/v1/admit/deployment", ah.ahServe(operations.DeploymentsValidation()))
	webhookMux.HandleFunc("/api/v1/mutate/pod", ah.ahServe(operations.PodsMutation()))
	webhookMux.HandleFunc("/healthz", healthzHandler())
	webhookMux.HandleFunc("/readyz", readyzHandler())
	webhookMux.HandleFunc("/", webServe())

	webhookServer := &http.Server{
		Addr:         cfg.WebServerIP + ":" + strconv.FormatInt(int64(cfg.WebServerPort), 10),
		Handler:      webhookMux,
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
			ClientAuth: tls.NoClientCert,
		},
	}

	// Start metrics server if enabled
	if cfg.EnableMetrics {
		go startMetricsServer(cfg.MetricsPort)
	}

	log.Printf("[INFO] Starting webhook server on %s:%d", cfg.WebServerIP, cfg.WebServerPort)
	if err := webhookServer.ListenAndServeTLS("", ""); err != nil {
		metrics.SetWebhookDown()
		log.Fatalf("[ERROR] Webhook server failed: %s\n", err)
	}
}

func startMetricsServer(port int) {
	metricsServer := &http.Server{
		Addr:    ":" + strconv.Itoa(port),
		Handler: metrics.Handler(),
	}

	log.Printf("[INFO] Starting metrics server on port %d", port)
	if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Printf("[ERROR] Metrics server failed: %v", err)
	}
}

func setCertificateExpiryMetrics(certPEM string) {
	block, _ := pem.Decode([]byte(certPEM))
	if block == nil {
		return
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return
	}

	metrics.SetCertificateExpiry(cert.NotAfter)
}

func healthzHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}
}

func readyzHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ready"))
	}
}

func webServe() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		httpAccessLog(r)
		crossSiteOrigin(w)
		strictTransport(w)

		switch {
		case r.Method != http.MethodGet:
			msg := fmt.Sprintf("incorrect method: got request type %s, expected request type %s", r.Method, http.MethodPost)
			log.Printf("[DEBUG] %s", msg)
			tmpltError(w, http.StatusMethodNotAllowed, msg)
		case r.URL.Path == "/api/v1/admin":
			tmpltAdminToggle(w, r.URL.Query())
		case r.URL.Path == "/healthcheck":
			tmpltHealthCheck(w)
		case r.URL.Path == "/":
			tmpltWebRoot(w)
		default:
			msg := fmt.Sprintf("Unable to locate requested path: '%s'", r.URL.Path)
			log.Printf("[DEBUG] %s", msg)
			tmpltError(w, http.StatusNotFound, msg)
		}
	}
}

type admissionHandler struct {
	decoder runtime.Decoder
	config  *config.Config
}

func (h *admissionHandler) ahServe(hook operations.Hook) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		httpAccessLog(r)
		crossSiteOrigin(w)
		strictTransport(w)

		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			msg := fmt.Sprintf("incorrect method: got request type %s, expected request type %s", r.Method, http.MethodPost)
			log.Printf("[DEBUG] %s", msg)
			metrics.RecordError("method_not_allowed", "admission")
			tmpltError(w, http.StatusMethodNotAllowed, msg)
			return
		}

		if contentType := r.Header.Get("Content-Type"); contentType != "application/json" {
			msg := "only content type 'application/json' is supported"
			log.Printf("[DEBUG] %s", msg)
			metrics.RecordError("invalid_content_type", "admission")
			tmpltError(w, http.StatusBadRequest, msg)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			msg := fmt.Sprintf("could not read request body: %v", err)
			log.Printf("[DEBUG] %s", msg)
			metrics.RecordError("body_read_error", "admission")
			tmpltError(w, http.StatusBadRequest, msg)
			return
		}

		var review admission.AdmissionReview
		if _, _, err := h.decoder.Decode(body, nil, &review); err != nil {
			msg := fmt.Sprintf("could not deserialize request: %v", err)
			log.Printf("[DEBUG] %s", msg)
			metrics.RecordError("decode_error", "admission")
			tmpltError(w, http.StatusBadRequest, msg)
			return
		}

		if review.Request == nil {
			msg := "malformed admission review: request is nil"
			log.Printf("[DEBUG] %s", msg)
			metrics.RecordError("nil_request", "admission")
			tmpltError(w, http.StatusBadRequest, msg)
			return
		}

		// Record admission request metrics
		namespace := review.Request.Namespace
		if namespace == "" {
			namespace = "cluster-scope"
		}
		resource := review.Request.Kind.Kind
		operation := string(review.Request.Operation)

		result, err := hook.Execute(review.Request, h.config)
		if err != nil {
			msg := err.Error()
			log.Printf("[ERROR] Internal Server Error: %s", msg)
			metrics.RecordError("hook_execution_error", "admission")
			metrics.RecordAdmissionRequest(operation, resource, namespace, false, time.Since(startTime))
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
				metrics.RecordError("patch_marshal_error", "admission")
				tmpltError(w, http.StatusInternalServerError, msg)
				return
			}
			admissionResponse.Response.Patch = patchBytes
			
			// Record mutation metrics
			metrics.RecordMutation(namespace, "labels", true)
			metrics.RecordLabelsApplied(namespace, resource, len(result.PatchOps))
		}

		res, err := json.Marshal(admissionResponse)
		if err != nil {
			msg := fmt.Sprintf("could not marshal response: %v", err)
			log.Printf("[ERROR] %s", msg)
			metrics.RecordError("response_marshal_error", "admission")
			tmpltError(w, http.StatusInternalServerError, msg)
			return
		}

		// Record successful admission request
		metrics.RecordAdmissionRequest(operation, resource, namespace, result.Allowed, time.Since(startTime))

		log.Printf("[DEBUG] Webhook [%s] - Resource: %s - Namespace: %s - Allowed: %t - Patches: %d", 
			review.Request.Operation, resource, namespace, result.Allowed, len(result.PatchOps))
		w.WriteHeader(http.StatusOK)
		w.Write(res)
	}
}
