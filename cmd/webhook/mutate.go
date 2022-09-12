package main

import (
	"fmt"
	"log"
	"strings"

	"encoding/json"
	"io/ioutil"
	"net/http"

	admission "k8s.io/api/admission/v1"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

var (
	codecs = serializer.NewCodecFactory(runtime.NewScheme())
)

func admissionReviewFromRequest(r *http.Request, deserializer runtime.Decoder) (*admission.AdmissionReview, error) {
	// Validate that the incoming content type is correct.
	if r.Header.Get("Content-Type") != "application/json" {
		return nil, fmt.Errorf("expected application/json content-type")
	}

	// Get the body data, which will be the AdmissionReview
	// content for the request.
	var body []byte
	if r.Body != nil {
		requestData, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}
		body = requestData
	}

	// Decode the request body into
	admissionReviewRequest := &admission.AdmissionReview{}
	if _, _, err := deserializer.Decode(body, nil, admissionReviewRequest); err != nil {
		return nil, err
	}

	return admissionReviewRequest, nil
}

func webMutatePod(w http.ResponseWriter, r *http.Request) {
	httpAccessLog(r)

	deserializer := codecs.UniversalDeserializer()

	// Parse the AdmissionReview from the http request.
	admissionReviewRequest, err := admissionReviewFromRequest(r, deserializer)
	if err != nil {
		msg := fmt.Sprintf("error getting admission review from request: %v", err)
		log.Printf("[ERROR] %v", msg)
		tmpltError(w, http.StatusBadRequest, msg)
		return
	}

	// Do server-side validation that we are only dealing with a pod resource. This
	// should also be part of the MutatingWebhookConfiguration in the cluster, but
	// we should verify here before continuing.
	podResource := meta.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}
	if admissionReviewRequest.Request.Resource != podResource {
		msg := fmt.Sprintf("did not receive pod, got %s", admissionReviewRequest.Request.Resource.Resource)
		log.Printf("[ERROR] %v", msg)
		tmpltError(w, http.StatusBadRequest, msg)
		return
	}

	// Decode the pod from the AdmissionReview.
	rawRequest := admissionReviewRequest.Request.Object.Raw
	pod := core.Pod{}
	if _, _, err := deserializer.Decode(rawRequest, nil, &pod); err != nil {
		msg := fmt.Sprintf("error decoding raw pod: %v", err)
		log.Printf("[ERROR] %v", msg)
		tmpltError(w, http.StatusBadRequest, msg)
		return
	}

	// check to see if mutation is required by looking for a label
	if !mutationRequired(&pod.ObjectMeta) {
		mutationResp(w, admissionReviewRequest, &admission.AdmissionResponse{Allowed: true})
	}

	// Add sidecar
	sidecarContainer := []core.Container{{
		Image: "ca-cert-server:latest",
	}}

	patchBytes, _ := createPatch(&pod, sidecarContainer)

	// respond with patch
	mutationResp(w, admissionReviewRequest, &admission.AdmissionResponse{
		Allowed: true,
		Patch:   patchBytes,
		PatchType: func() *admission.PatchType {
			pt := admission.PatchTypeJSONPatch
			return &pt
		}(),
	})
}

// prepare response
func mutationResp(w http.ResponseWriter, aRRequest *admission.AdmissionReview, aResponse *admission.AdmissionResponse) {
	var aRResponse admission.AdmissionReview
	aRResponse.Response = aResponse
	aRResponse.SetGroupVersionKind(aRRequest.GroupVersionKind())
	aRResponse.Response.UID = aRRequest.Request.UID

	resp, err := json.Marshal(aRResponse)
	if err != nil {
		msg := fmt.Sprintf("error marshalling response json: %v", err)
		log.Printf("[ERROR] %v", msg)
		tmpltError(w, http.StatusBadRequest, msg)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

// create mutation patch for resources
func createPatch(pod *core.Pod, containers []core.Container) ([]byte, error) {
	var (
		patch []patchOperation
		first bool
		value interface{}
	)

	if len(pod.Spec.Containers) == 0 {
		first = true
	}

	for _, add := range containers {
		value = add
		path := "/spec/containers"
		if first {
			first = false
			value = []core.Container{add}
		} else {
			path = path + "/-"
		}
		patch = append(patch, patchOperation{
			Op:    "add",
			Path:  path,
			Value: value,
		})
	}

	return json.Marshal(patch)
}

// Check whether the target resourse needs to be mutated
func mutationRequired(metadata *meta.ObjectMeta) bool {
	var ignoredNamespaces = []string{
		meta.NamespaceSystem,
		meta.NamespacePublic,
	}

	// skip special kubernetes system namespaces
	for _, namespace := range ignoredNamespaces {
		if metadata.Namespace == namespace {
			log.Printf("[TRACE] Skip mutation for %v for it's in special namespace:%v", metadata.Name, metadata.Namespace)
			return false
		}
	}

	annotations := metadata.GetLabels()
	if annotations == nil {
		annotations = map[string]string{}
	}

	// determine whether to perform mutation based on annotation for the target resource
	var required bool
	switch strings.ToLower(annotations["sidecar-injector-webhook/inject"]) {
	case "yes", "y", "true", "t", "on":
		required = true
	default:
		required = false
	}

	log.Printf("[TRACE] Mutation policy for %v/%v: required:%v", metadata.Namespace, metadata.Name, required)
	return required
}
