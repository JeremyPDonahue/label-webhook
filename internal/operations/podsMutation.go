package operations

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	admission "k8s.io/api/admission/v1"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"mutating-webhook/internal/config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func PodsMutation() Hook {
	return Hook{
		Create: podLabelingMutation(),
		Update: podLabelingMutation(),
		// default allow
		Delete: func(r *admission.AdmissionRequest, cfg *config.Config) (*Result, error) {
			return &Result{Allowed: true}, nil
		},
		Connect: func(r *admission.AdmissionRequest, cfg *config.Config) (*Result, error) {
			return &Result{Allowed: true}, nil
		},
	}
}

func podLabelingMutation() AdmitFunc {
	return func(r *admission.AdmissionRequest, cfg *config.Config) (*Result, error) {
		var operations []PatchOperation

		// Skip if labeling is disabled
		if !cfg.EnableLabeling {
			log.Printf("[DEBUG] Custom labeling is disabled")
			return &Result{Allowed: true}, nil
		}

		// Skip if namespace is excluded
		if isNamespaceExcluded(r.Namespace, cfg.ExcludedNamespaces) {
			log.Printf("[DEBUG] Namespace %s is excluded from labeling", r.Namespace)
			return &Result{Allowed: true}, nil
		}

		pod, err := parsePod(r.Object.Raw)
		if err != nil {
			return &Result{Msg: err.Error()}, nil
		}

		// Check if pod is administratively exempt
		if cfg.AllowAdminNoMutate && isAdminExempt(pod) {
			log.Printf("[INFO] Pod %s/%s is administratively exempt from labeling", r.Namespace, pod.Name)
			return &Result{Allowed: true}, nil
		}

		// Check if this is a dry run
		if cfg.DryRun {
			log.Printf("[INFO] DRY RUN: Would apply custom labels to pod %s/%s", r.Namespace, pod.Name)
			return &Result{Allowed: true}, nil
		}

		// Generate custom labels
		customLabels := generateCustomLabels(cfg, r, pod)
		if len(customLabels) == 0 {
			log.Printf("[DEBUG] No custom labels to apply to pod %s/%s", r.Namespace, pod.Name)
			return &Result{Allowed: true}, nil
		}

		// Apply labels
		labelPatches := createLabelPatches(pod, customLabels)
		if len(labelPatches) > 0 {
			operations = append(operations, labelPatches...)
		}

		// Add annotations about the mutation
		annotationPatches := createAnnotationPatches(pod, cfg)
		if len(annotationPatches) > 0 {
			operations = append(operations, annotationPatches...)
		}

		if len(operations) > 0 {
			log.Printf("[INFO] Applied %d custom labels to pod %s/%s", len(customLabels), r.Namespace, pod.Name)
		}

		return &Result{
			Allowed:  true,
			PatchOps: operations,
		}, nil
	}
}

func isNamespaceExcluded(namespace string, excludedNamespaces []string) bool {
	// Always exclude system namespaces
	systemNamespaces := []string{
		"kube-system",
		"kube-public",
		"kube-node-lease",
		"openshift-system",
		"openshift-kube-apiserver",
		"openshift-kube-scheduler",
		"openshift-kube-controller-manager",
		"openshift-etcd",
		"openshift-apiserver",
		"openshift-controller-manager",
		"openshift-authentication",
		"openshift-oauth-apiserver",
		"openshift-service-ca",
		"openshift-network-operator",
		"openshift-cluster-machine-approver",
		"openshift-cluster-samples-operator",
		"openshift-cluster-storage-operator",
		"openshift-cluster-version",
		"openshift-config",
		"openshift-config-managed",
		"openshift-console",
		"openshift-console-operator",
		"openshift-dns",
		"openshift-dns-operator",
		"openshift-image-registry",
		"openshift-ingress",
		"openshift-ingress-operator",
		"openshift-machine-api",
		"openshift-machine-config-operator",
		"openshift-monitoring",
		"openshift-multus",
		"openshift-node",
		"openshift-operator-lifecycle-manager",
		"openshift-operators",
		"openshift-ovn-kubernetes",
		"openshift-sdn",
		"openshift-user-workload-monitoring",
		"openshift-webhook", // Our own namespace
	}

	allExcluded := append(systemNamespaces, excludedNamespaces...)
	for _, excluded := range allExcluded {
		if namespace == excluded {
			return true
		}
	}
	return false
}

func isAdminExempt(pod *core.Pod) bool {
	if pod.Annotations != nil {
		if exempt, exists := pod.Annotations["webhook.openshift.io/exempt"]; exists && exempt == "true" {
			return true
		}
		if exempt, exists := pod.Annotations["admission.webhook/exempt"]; exists && exempt == "true" {
			return true
		}
	}
	return false
}

func generateCustomLabels(cfg *config.Config, r *admission.AdmissionRequest, pod *core.Pod) map[string]string {
	labels := make(map[string]string)

	// Get appid from the namespace
	appid := getAppIDFromNamespace(r.Namespace)
	if appid != "" {
		labels[fmt.Sprintf("%s/appid", cfg.LabelPrefix)] = appid
	}

	// Add base labels
	labels[fmt.Sprintf("%s/webhook", cfg.LabelPrefix)] = "custom-labels-mutator"
	labels[fmt.Sprintf("%s/organization", cfg.LabelPrefix)] = cfg.Organization
	labels[fmt.Sprintf("%s/environment", cfg.LabelPrefix)] = cfg.Environment
	labels[fmt.Sprintf("%s/cluster", cfg.LabelPrefix)] = cfg.ClusterName
	labels[fmt.Sprintf("%s/timestamp", cfg.LabelPrefix)] = time.Now().UTC().Format("2006-01-02T15:04:05Z")

	// Add namespace and workload info
	labels[fmt.Sprintf("%s/namespace", cfg.LabelPrefix)] = r.Namespace
	if r.UserInfo.Username != "" {
		labels[fmt.Sprintf("%s/created-by", cfg.LabelPrefix)] = sanitizeLabelValue(r.UserInfo.Username)
	}

	// Add owner reference information if available
	if pod.OwnerReferences != nil && len(pod.OwnerReferences) > 0 {
		owner := pod.OwnerReferences[0] // Primary owner
		labels[fmt.Sprintf("%s/workload-type", cfg.LabelPrefix)] = strings.ToLower(owner.Kind)
		labels[fmt.Sprintf("%s/workload-name", cfg.LabelPrefix)] = sanitizeLabelValue(owner.Name)
	}

	// Add custom labels from configuration
	for key, value := range cfg.CustomLabels {
		labels[sanitizeLabelKey(key)] = sanitizeLabelValue(value)
	}

	return labels
}

func getAppIDFromNamespace(namespace string) string {
	// Create in-cluster client
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Printf("[ERROR] Failed to create in-cluster config: %v", err)
		return ""
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Printf("[ERROR] Failed to create clientset: %v", err)
		return ""
	}

	// Get the namespace object
	ns, err := clientset.CoreV1().Namespaces().Get(context.TODO(), namespace, metav1.GetOptions{})
	if err != nil {
		log.Printf("[ERROR] Failed to get namespace %s: %v", namespace, err)
		return ""
	}

	// Check for appid in annotations
	if ns.Annotations != nil {
		if appid, exists := ns.Annotations["appid"]; exists {
			log.Printf("[DEBUG] Found appid '%s' in namespace '%s'", appid, namespace)
			return appid
		}
	}

	// Check for appid in labels as fallback
	if ns.Labels != nil {
		if appid, exists := ns.Labels["appid"]; exists {
			log.Printf("[DEBUG] Found appid '%s' in namespace labels '%s'", appid, namespace)
			return appid
		}
	}

	log.Printf("[DEBUG] No appid found in namespace '%s'", namespace)
	return ""
}

func createLabelPatches(pod *core.Pod, customLabels map[string]string) []PatchOperation {
	var patches []PatchOperation

	if pod.Labels == nil {
		// If no labels exist, create the labels map
		patches = append(patches, AddPatchOperation("/metadata/labels", customLabels))
	} else {
		// Add individual labels
		for key, value := range customLabels {
			// Check if label already exists
			if existingValue, exists := pod.Labels[key]; !exists || existingValue != value {
				path := fmt.Sprintf("/metadata/labels/%s", strings.ReplaceAll(key, "/", "~1"))
				patches = append(patches, AddPatchOperation(path, value))
			}
		}
	}

	return patches
}

func createAnnotationPatches(pod *core.Pod, cfg *config.Config) []PatchOperation {
	var patches []PatchOperation
	annotations := make(map[string]string)

	annotations["webhook.openshift.io/mutated-by"] = cfg.WebhookName
	annotations["webhook.openshift.io/mutated-at"] = time.Now().UTC().Format(time.RFC3339)
	annotations["webhook.openshift.io/version"] = "v1.0.0"

	if pod.Annotations == nil {
		patches = append(patches, AddPatchOperation("/metadata/annotations", annotations))
	} else {
		for key, value := range annotations {
			path := fmt.Sprintf("/metadata/annotations/%s", strings.ReplaceAll(key, "/", "~1"))
			patches = append(patches, AddPatchOperation(path, value))
		}
	}

	return patches
}

func sanitizeLabelKey(key string) string {
	// Kubernetes label keys must be valid DNS subdomains
	// Replace invalid characters with hyphens
	sanitized := strings.ToLower(key)
	sanitized = strings.ReplaceAll(sanitized, "_", "-")
	sanitized = strings.ReplaceAll(sanitized, " ", "-")
	sanitized = strings.ReplaceAll(sanitized, ".", "-")
	return sanitized
}

func sanitizeLabelValue(value string) string {
	// Kubernetes label values must be valid
	// Remove or replace invalid characters
	sanitized := strings.ToLower(value)
	sanitized = strings.ReplaceAll(sanitized, "_", "-")
	sanitized = strings.ReplaceAll(sanitized, " ", "-")
	sanitized = strings.ReplaceAll(sanitized, "@", "-at-")
	sanitized = strings.ReplaceAll(sanitized, ":", "-")
	
	// Truncate if too long (max 63 characters)
	if len(sanitized) > 63 {
		sanitized = sanitized[:63]
	}
	
	return sanitized
}
