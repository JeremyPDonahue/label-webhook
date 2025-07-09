package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// Webhook admission metrics
	admissionRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "webhook_admission_requests_total",
			Help: "Total number of admission requests processed by the webhook",
		},
		[]string{"operation", "resource", "namespace", "allowed"},
	)

	admissionRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "webhook_admission_request_duration_seconds",
			Help:    "Duration of admission requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation", "resource"},
	)

	// Labeling metrics
	labelsAppliedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "webhook_labels_applied_total",
			Help: "Total number of labels applied by the webhook",
		},
		[]string{"namespace", "workload_type"},
	)

	mutationsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "webhook_mutations_total",
			Help: "Total number of mutations performed by the webhook",
		},
		[]string{"namespace", "mutation_type", "success"},
	)

	// Error metrics
	errorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "webhook_errors_total",
			Help: "Total number of errors encountered by the webhook",
		},
		[]string{"error_type", "operation"},
	)

	// Health metrics
	webhookUp = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "webhook_up",
			Help: "Whether the webhook is up and running",
		},
	)

	certificateExpiryTime = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "webhook_certificate_expiry_timestamp",
			Help: "Timestamp when the webhook certificate expires",
		},
	)
)

func init() {
	// Register metrics
	prometheus.MustRegister(
		admissionRequestsTotal,
		admissionRequestDuration,
		labelsAppliedTotal,
		mutationsTotal,
		errorsTotal,
		webhookUp,
		certificateExpiryTime,
	)

	// Set webhook as up
	webhookUp.Set(1)
}

// RecordAdmissionRequest records metrics for admission requests
func RecordAdmissionRequest(operation, resource, namespace string, allowed bool, duration time.Duration) {
	admissionRequestsTotal.WithLabelValues(
		operation,
		resource,
		namespace,
		strconv.FormatBool(allowed),
	).Inc()

	admissionRequestDuration.WithLabelValues(
		operation,
		resource,
	).Observe(duration.Seconds())
}

// RecordLabelsApplied records metrics for applied labels
func RecordLabelsApplied(namespace, workloadType string, count int) {
	labelsAppliedTotal.WithLabelValues(
		namespace,
		workloadType,
	).Add(float64(count))
}

// RecordMutation records metrics for mutations
func RecordMutation(namespace, mutationType string, success bool) {
	mutationsTotal.WithLabelValues(
		namespace,
		mutationType,
		strconv.FormatBool(success),
	).Inc()
}

// RecordError records error metrics
func RecordError(errorType, operation string) {
	errorsTotal.WithLabelValues(
		errorType,
		operation,
	).Inc()
}

// SetCertificateExpiry sets the certificate expiry timestamp
func SetCertificateExpiry(expiryTime time.Time) {
	certificateExpiryTime.Set(float64(expiryTime.Unix()))
}

// Handler returns the metrics HTTP handler
func Handler() http.Handler {
	return promhttp.Handler()
}

// SetWebhookDown marks the webhook as down
func SetWebhookDown() {
	webhookUp.Set(0)
}

// SetWebhookUp marks the webhook as up
func SetWebhookUp() {
	webhookUp.Set(1)
}
