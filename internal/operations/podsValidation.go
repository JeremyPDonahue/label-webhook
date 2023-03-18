package operations

import (
	"strings"

	admission "k8s.io/api/admission/v1"
)

func PodsValidation() Hook {
	return Hook{
		Create: podValidationCreate(),
	}
}

func podValidationCreate() AdmitFunc {
	return func(r *admission.AdmissionRequest) (*Result, error) {
		pod, err := parsePod(r.Object.Raw)
		if err != nil {
			return &Result{Msg: err.Error()}, nil
		}

		for _, c := range pod.Spec.Containers {
			if strings.HasSuffix(c.Image, ":latest") {
				return &Result{Msg: "You cannot use the tag 'latest' in a container."}, nil
			}
		}

		return &Result{Allowed: true}, nil
	}
}
