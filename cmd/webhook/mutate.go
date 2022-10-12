package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	admission "k8s.io/api/admission/v1"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

type result struct {
	Allowed  bool
	Msg      string
	PatchOps []patchOperation
}

type admitFunc func(request *admission.AdmissionRequest) (*result, error)

type hook struct {
	Create  admitFunc
	Delete  admitFunc
	Update  admitFunc
	Connect admitFunc
}

func webMutatePod(w http.ResponseWriter, r *http.Request) {
	//https://github.com/douglasmakey/admissioncontroller

	podsValidation := hook{
		Create: validateCreate(),
	}

	admissionHandler := &struct {
		decoder runtime.Decoder
	}{
		decoder: serializer.NewCodecFactory(runtime.NewScheme()).UniversalDeserializer(),
	}

	// read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		tmpltError(w, http.StatusBadRequest, "No data in request body.")
		return
	}

	// see if request body can be decoded
	var review admission.AdmissionReview
	if _, _, err := admissionHandler.decoder.Decode(body, nil, &review); err != nil {
		tmpltError(w, http.StatusBadRequest, "Unable to decode request body.")
		return
	}

	var o *result
	switch review.Request.Operation {
	case admission.Create:
		if podsValidation.Create == nil {
			tmpltError(w, http.StatusBadRequest, fmt.Sprintf("operation %s is not registered", review.Request.Operation))
			return
		}
		o, _ = podsValidation.Create(review.Request)
	case admission.Update:
		if podsValidation.Update == nil {
			tmpltError(w, http.StatusBadRequest, fmt.Sprintf("operation %s is not registered", review.Request.Operation))
			return
		}
		o, _ = podsValidation.Update(review.Request)
	case admission.Delete:
		if podsValidation.Delete == nil {
			tmpltError(w, http.StatusBadRequest, fmt.Sprintf("operation %s is not registered", review.Request.Operation))
			return
		}
		o, _ = podsValidation.Delete(review.Request)
	case admission.Connect:
		if podsValidation.Connect == nil {
			tmpltError(w, http.StatusBadRequest, fmt.Sprintf("operation %s is not registered", review.Request.Operation))
			return
		}
		o, _ = podsValidation.Connect(review.Request)
	}

	admissionResult := admission.AdmissionReview{
		Response: &admission.AdmissionResponse{
			UID:     review.Request.UID,
			Allowed: o.Allowed,
			Result: &meta.Status{
				Message: o.Msg,
			},
		},
	}

	resp, _ := json.Marshal(admissionResult)
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func validateCreate() admitFunc {
	return func(r *admission.AdmissionRequest) (*result, error) {
		pod, err := parsePod(r.Object.Raw)
		if err != nil {
			return &result{Msg: err.Error()}, nil
		}

		for _, c := range pod.Spec.Containers {
			if strings.HasSuffix(c.Image, ":latest") {
				return &result{Msg: "You cannot use the tag 'latest' in a container."}, nil
			}
		}

		return &result{Allowed: true}, nil
	}
}

func parsePod(object []byte) (*core.Pod, error) {
	var pod core.Pod
	if err := json.Unmarshal(object, &pod); err != nil {
		return nil, err
	}

	return &pod, nil
}
