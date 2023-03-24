package operations

//https://github.com/douglasmakey/admissioncontroller

import (
	"fmt"

	admission "k8s.io/api/admission/v1"

	"mutating-webhook/internal/config"
)

// Result contains the result of an admission request
type Result struct {
	Allowed  bool
	Msg      string
	PatchOps []PatchOperation
}

// AdmitFunc defines how to process an admission request
type AdmitFunc func(request *admission.AdmissionRequest, cfg *config.Config) (*Result, error)

// Hook represents the set of functions for each operation in an admission webhook.
type Hook struct {
	Create  AdmitFunc
	Delete  AdmitFunc
	Update  AdmitFunc
	Connect AdmitFunc
}

// Execute evaluates the request and try to execute the function for operation specified in the request.
func (h *Hook) Execute(r *admission.AdmissionRequest, cfg *config.Config) (*Result, error) {
	switch r.Operation {
	case admission.Create:
		return wrapperExecution(h.Create, r, cfg)
	case admission.Update:
		return wrapperExecution(h.Update, r, cfg)
	case admission.Delete:
		return wrapperExecution(h.Delete, r, cfg)
	case admission.Connect:
		return wrapperExecution(h.Connect, r, cfg)
	}

	return &Result{Msg: fmt.Sprintf("Invalid operation: %s", r.Operation)}, nil
}

func wrapperExecution(fn AdmitFunc, r *admission.AdmissionRequest, cfg *config.Config) (*Result, error) {
	if fn == nil {
		return nil, fmt.Errorf("operation %s is not registered", r.Operation)
	}

	return fn(r, cfg)
}
