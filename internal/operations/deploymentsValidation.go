package operations

import (
	"mutating-webhook/internal/config"

	admission "k8s.io/api/admission/v1"
)

func DeploymentsValidation() Hook {
	return Hook{
		// default allow
		Create: func(r *admission.AdmissionRequest, cfg config.Config) (*Result, error) {
			return &Result{Allowed: true}, nil
		},
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
