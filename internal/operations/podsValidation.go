package operations

import (
	"log"
	"strings"

	"mutating-webhook/internal/config"

	admission "k8s.io/api/admission/v1"
)

func PodsValidation() Hook {
	return Hook{
		Create: podValidationCreate(),
		// default allow
		Delete: func(r *admission.AdmissionRequest, cfg config.Config) (*Result, error) {
			return &Result{Allowed: true}, nil
		},
		Update: func(r *admission.AdmissionRequest, cfg config.Config) (*Result, error) {
			return &Result{Allowed: true}, nil
		},
		Connect: func(r *admission.AdmissionRequest, cfg config.Config) (*Result, error) {
			return &Result{Allowed: true}, nil
		},
	}
}

func podValidationCreate() AdmitFunc {
	return func(r *admission.AdmissionRequest, cfg config.Config) (*Result, error) {
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
