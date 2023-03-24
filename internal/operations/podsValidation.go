package operations

import (
	"log"
	"strings"

	admission "k8s.io/api/admission/v1"

	"mutating-webhook/internal/config"
)

func PodsValidation() Hook {
	return Hook{
		// default allow
		Create: func(r *admission.AdmissionRequest, cfg *config.Config) (*Result, error) {
			return &Result{Allowed: true}, nil
		},
		Delete: func(r *admission.AdmissionRequest, cfg *config.Config) (*Result, error) {
			return &Result{Allowed: true}, nil
		},
		Update: func(r *admission.AdmissionRequest, cfg *config.Config) (*Result, error) {
			return &Result{Allowed: true}, nil
		},
		Connect: func(r *admission.AdmissionRequest, cfg *config.Config) (*Result, error) {
			return &Result{Allowed: true}, nil
		},
	}
}

func podValidationCreate() AdmitFunc {
	return func(r *admission.AdmissionRequest, cfg *config.Config) (*Result, error) {
		pod, err := parsePod(r.Object.Raw)
		if err != nil {
			return &Result{Msg: err.Error()}, nil
		}

		for _, c := range pod.Spec.Containers {
			if strings.HasSuffix(c.Image, ":latest") {
				msg := "You cannot use the tag 'latest' in a container."
				log.Printf("[TRACE] Request Rejectd: %s", msg)
				return &Result{Msg: msg}, nil
			}
		}

		return &Result{Allowed: true}, nil
	}
}
