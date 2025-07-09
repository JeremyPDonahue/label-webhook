package operations

import (
	"context"
	"fmt"
	"log"
	"strings"

	admission "k8s.io/api/admission/v1"

	"mutating-webhook/internal/config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func PodsMutation() Hook {
	return Hook{
		Create: podAppIDMutation(),
		Update: podAppIDMutation(),
		// default allow
		Delete: func(r *admission.AdmissionRequest, cfg *config.Config) (*Result, error) {
			return &Result{Allowed: true}, nil
		},
		Connect: func(r *admission.AdmissionRequest, cfg *config.Config) (*Result, error) {
			return &Result{Allowed: true}, nil
		},
	}
}

func podAppIDMutation() AdmitFunc {
	return func(r *admission.AdmissionRequest, cfg *config.Config) (*Result, error) {
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

		// Check if this is a dry run
		if cfg.DryRun {
			log.Printf("[INFO] DRY RUN: Would apply appid label to pod %s/%s", r.Namespace, pod.Name)
			return &Result{Allowed: true}, nil
		}

		// Get appid from the namespace
		appid := getAppIDFromNamespace(r.Namespace)
		if appid == "" {
			log.Printf("[DEBUG] No appid found in namespace %s, skipping", r.Namespace)
			return &Result{Allowed: true}, nil
		}

		// Apply only the appid label
		appidLabelKey := fmt.Sprintf("%s/appid", cfg.LabelPrefix)
		var operations []PatchOperation

		// Check if appid label already exists with the same value
		if pod.Labels != nil {
			if existingAppID, exists := pod.Labels[appidLabelKey]; exists && existingAppID == appid {
				log.Printf("[DEBUG] AppID label already exists with correct value for pod %s/%s", r.Namespace, pod.Name)
				return &Result{Allowed: true}, nil
			}
		}

		// Add the appid label
		if pod.Labels == nil {
			// If no labels exist, create the labels map with only appid
			labels := map[string]string{appidLabelKey: appid}
			operations = append(operations, AddPatchOperation("/metadata/labels", labels))
		} else {
			// Add individual appid label
			path := fmt.Sprintf("/metadata/labels/%s", strings.ReplaceAll(appidLabelKey, "/", "~1"))
			operations = append(operations, AddPatchOperation(path, appid))
		}

		log.Printf("[INFO] Applied appid label '%s' to pod %s/%s", appid, r.Namespace, pod.Name)

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
