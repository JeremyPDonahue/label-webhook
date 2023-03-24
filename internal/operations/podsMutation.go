package operations

import (
	"fmt"

	admission "k8s.io/api/admission/v1"
	core "k8s.io/api/core/v1"

	"mutating-webhook/internal/config"
)

func PodsMutation() Hook {
	return Hook{
		Create: podMutationCreate(),
		// default allow
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

func podMutationCreate() AdmitFunc {
	return func(r *admission.AdmissionRequest, cfg *config.Config) (*Result, error) {
		var operations []PatchOperation
		pod, err := parsePod(r.Object.Raw)
		if err != nil {
			return &Result{Msg: err.Error()}, nil
		}

		// if pod is administratively exempt
		if cfg.AllowAdminNoMutate && func(pod *core.Pod) bool {
			for label, value := range pod.Annotations {
				if label == "AdminNoMutate" && value == "true" {
					return false
				}
			}
			return true
		}(pod) {
			// mutate pod (annotation)
			metadata := map[string]string{
				"mutation-status": "pod mutated by mutation-controller",
			}
			// add original image to annotations
			for _, p := range pod.Spec.Containers {
				metadata[fmt.Sprintf("mutation-original-image-%s", p.Name)] = p.Image
			}
			// add annotation stating that the pos had been mutated
			operations = append(operations, AddPatchOperation("/metadata/annotations", metadata))

			// add image mutation
		}

		return &Result{
			Allowed:  true,
			PatchOps: operations,
		}, nil
	}
}
