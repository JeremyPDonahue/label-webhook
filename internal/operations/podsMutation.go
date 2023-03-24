package operations

import (
	"fmt"
	"log"
	"regexp"

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
		var (
			operations []PatchOperation
			mutated    bool
		)

		pod, err := parsePod(r.Object.Raw)
		if err != nil {
			return &Result{Msg: err.Error()}, nil
		}

		// if pod is administratively exempt
		if func(serviceEnabled bool, pod *core.Pod) bool {
			if serviceEnabled {
				for label, value := range pod.Annotations {
					if label == "AdminNoMutate" && value == "true" {
						return false
					}
				}
			}
			return true
		}(cfg.AllowAdminNoMutate, pod) {
			for i, p := range pod.Spec.Containers {
				img, mutationOccurred, err := customDockerRegistry(p.Image, cfg)
				if err != nil {
					return &Result{Msg: err.Error()}, nil
				}
				if mutationOccurred {
					mutated = true
					path := fmt.Sprintf("/spec/containers/%d/image", i)
					operations = append(operations, ReplacePatchOperation(path, img))
					log.Printf("[TRACE] Image has been mutated: %s -> %s", p.Image, img)
				} else {
					log.Printf("[TRACE] No mutation required for image: %s", p.Image)
				}
			}
		} else {
			log.Printf("[TRACE] Mutations administratively disabled.")
		}

		if mutated {
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
		}

		return &Result{
			Allowed:  true,
			PatchOps: operations,
		}, nil
	}
}

func customDockerRegistry(imgPath string, cfg *config.Config) (string, bool, error) {
	if len(cfg.DockerhubRegistry) == 0 {
		return imgPath, false, nil
	}

	// regex match official project
	reg, err := regexp.Compile(`^([a-z]|\.|_|-)+\:([a-zA-Z0-9]|_|\.|-)+$`)
	if err != nil {
		return "", false, fmt.Errorf("Unable to parse regex: %v", err)
	}
	if reg.MatchString(imgPath) {
		log.Printf("Official docker image detected: %s", imgPath)
		return fmt.Sprintf("%s/library/%s", cfg.DockerhubRegistry, imgPath), true, nil
	}
	return "", false, nil
}
